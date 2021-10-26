package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/desertbit/grumble"
)

// WolfPackApp - global value to store the current app context for the grumble app and access things like command execution in the grumble context
var WolfPackApp *grumble.App

// IsWolfPackExec - global value to let grumble run functions determine if the current command is being executed in the context of a
var IsWolfPackExec bool

// CurrentOperator - keeps track of the current user that is actively interacting with the WolfPack server during the request flow
var CurrentOperator string

// WolfPackServerHandler - Handles all Wolfpack server requests over HTTPS by passing data to handler sub-functions
//
// Also sets HTTP server parameters and any other applicable HTTP server level variables.
func WolfPackServerHandler(w http.ResponseWriter, r *http.Request) {
	// Setup webserver attributes like headers and response information
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		handleWolfPackRequests(w, r)
	default:
		// Invalid request type, stay silent don't respond to anything that isn't pre-defined
		return
	}

	return
}

// handleWolfPackRequests - handles any incoming GET requests received by the HTTP(S) listener. Once all values are handled various Implant data update/response routines are executed where relevant based on the provided parameters.
//
// When requests are received, the URL parameters are extracted, validated and stored.
//
// HTTP GET Requests are expected to be provided as URL parameters like any other web request. The following parameters can be provided:
//
// PSK - a string that is the client Pre-Shared Key that the the implant will send to be compared for authentication to the server PSK
//
// Username - a string that is a unique Username that is defined by the operator administering wolfpack server users. This is sent to identify what user is connecting.
//
// Command - a command string (slice) issued by a user to be transmitted and executed by the Lupo server
//
// Polling - boolean status indicator to know if the incoming request was from the Lupo client polling functions or the user
//
// IsSessionShell = a boolean value to indicate whether or not the current shell type being interacted with is a session or the core Lupo CLI. This is required to access session specific functions like executing commands and swapping sessions within the Wolfpack server.
//
// ActiveSession - an integer that is the active session an operator is interacting with when executing commands, only applies to session sub-shell/nested shell commands
//
// FileName - a string value provided by an implant that is the filename for a file being sent to download or upload.
//
// File - a string value that is expected to be a base64 encoded string that is a file to download or upload.
//
// IsGetChatLog - boolean status to determine whether or not a user has just entered the chat to send them the full chat log.
//
// IsChatShell = a boolean value to indicate whether or not the current shell type being interacted with is the chat CLI or the core Lupo CLI. This is required to access the Wolfpack server chat and send/receive messages.

func handleWolfPackRequests(w http.ResponseWriter, r *http.Request) {

	// Construct variables for GET URL paramaters
	getParams := r.URL.Query()
	var getPSK string
	var getUsername string
	var getCommand []string
	var getPolling = false
	var getIsSessionShell = false
	var getIsChatShell = false
	var getActiveSession int
	var getFileName string
	var getFile string
	var getIsGetChatLog = false

	// Get the Remote Address of the Implant from the request
	remoteAddr := r.RemoteAddr

	// Check GET URL parameters and handle errors
	if len(getParams["psk"]) > 0 {
		getPSK = getParams["psk"][0]
	} else {
		errorString := "wolfpack GET Request did not provide PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if len(getParams["user"]) > 0 {
		getUsername = getParams["user"][0]
	}

	if len(getParams["command"]) > 0 {
		getCommand = strings.Split(getParams["command"][0], " ")
	}

	if len(getParams["polling"]) > 0 {
		polling, err := strconv.ParseBool(getParams["polling"][0])

		if err != nil {
			errorString := "wolfpack GET Request could not parse getPolling parameter as bool, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}

		getPolling = polling
	}

	if len(getParams["isSessionShell"]) > 0 {
		isSessionShell, err := strconv.ParseBool(getParams["isSessionShell"][0])

		if err != nil {
			errorString := "wolfpack GET Request could not parse getIsSessionShell parameter as bool, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}

		if len(getParams["activeSession"]) > 0 {

			activeSession, err := strconv.Atoi(getParams["activeSession"][0])

			if err != nil {
				errorString := "wolfpack GET Request could not convert getActiveSession parameter as int, request ignored"
				core.LogData(errorString)
				returnErr := errors.New(errorString)
				ErrorHandler(returnErr)
				return
			}
			getActiveSession = activeSession
		}

		getIsSessionShell = isSessionShell
	}

	if len(getParams["isChatShell"]) > 0 {
		isChatShell, err := strconv.ParseBool(getParams["isChatShell"][0])

		if err != nil {
			errorString := "wolfpack GET Request could not parse getIsChatShell parameter as bool, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}

		getIsChatShell = isChatShell

		if len(getParams["getChatLog"]) > 0 {
			getIsGetChatLog, err = strconv.ParseBool(getParams["getChatLog"][0])

			if err != nil {
				errorString := "wolfpack GET Request could not parse getIsChatShell parameter as bool, request ignored"
				core.LogData(errorString)
				returnErr := errors.New(errorString)
				ErrorHandler(returnErr)
				return
			}

			if getIsGetChatLog {
				chatLog, err := ioutil.ReadFile(".lupo.chat.log") // just pass the file name
				if err != nil {
					fmt.Print(err)
				}

				jsonData := `{"chatData":"` + string(chatLog) + `"}`
				currentWolf := core.Wolves[getUsername]
				core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
			}
		}

		if getIsChatShell && !getIsGetChatLog {

			currentWolf := core.Wolves[getUsername]

			if len(getParams["message"]) > 0 {
				message := getParams["message"][0]
				wolfpackChat(message, currentWolf, w)
			}

			var previousOffset int64 = 0
			file, err := os.Open(".lupo.chat.log")
			if err != nil {
				panic(err)
			}

			defer file.Close()

			reader := bufio.NewReader(file)

			// we need to calculate the size of the last line for file.ReadAt(offset) to work

			// NOTE : not a very effective solution as we need to read
			// the entire file at least for 1 pass :(

			lastLineSize := 0

			for {
				line, _, err := reader.ReadLine()

				if err == io.EOF {
					break
				}

				lastLineSize = len(line)
			}

			fileInfo, err := os.Stat(".lupo.chat.log")

			// make a buffer size according to the lastLineSize
			buffer := make([]byte, lastLineSize)

			// +1 to compensate for the initial 0 byte of the line
			// otherwise, the initial character of the line will be missing

			// instead of reading the whole file into memory, we just read from certain offset

			offset := fileInfo.Size() - int64(lastLineSize+1)
			numRead, err := file.ReadAt(buffer, offset)

			if previousOffset != offset {

				// print out last line content
				buffer = buffer[:numRead]
				jsonData := `{"chatData":"` + string(buffer) + `"}`
				core.BroadcastWolfPackChat(jsonData)
				previousOffset = offset
			}
		}

	}

	if len(getParams["filename"]) > 0 {
		getFileName = getParams["filename"][0]
	}

	if len(getParams["file"]) > 0 {
		getFile = getParams["file"][0]
	}

	if getPSK != core.Wolves[getUsername].WolfPSK {
		errorString := "wolfpack GET Request Invalid PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if getPolling || getIsChatShell {
		currentWolf := core.Wolves[getUsername]

		if currentWolf.Broadcast != "" {

			w.Write([]byte(currentWolf.Broadcast))
			// Clear the response once returned
			core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, "")

		}

	} else if !getPolling {
		core.UpdateWolf(getUsername, remoteAddr)
		CurrentOperator = getUsername
		IsWolfPackExec = true
		core.LogData(getUsername + "@" + remoteAddr + " executed: " + strings.Join(getCommand, " "))

		if getIsSessionShell {

			if getCommand[0] == "back" {
				core.LogData(getUsername + " executed: back")

			} else if getCommand[0] == "session" {

				var sessionID string
				if len(getParams["id"]) > 0 {
					sessionID = getParams["id"][0]
				}

				session, err := strconv.Atoi(sessionID)

				if err != nil {
					errorString := "wolfpack GET Request could not convert session ID to int, request ignored..."
					core.LogData(errorString)
					returnErr := errors.New(errorString)
					ErrorHandler(returnErr)
					return
				}
				sessionExists := core.SessionExists(session)

				if sessionExists {
					core.AssignWolfResponse(CurrentOperator, core.Wolves[CurrentOperator].Rhost, "true")
					core.LogData(CurrentOperator + " executed: session " + sessionID)

				} else {
					core.AssignWolfResponse(CurrentOperator, core.Wolves[CurrentOperator].Rhost, "false")
					core.LogData(CurrentOperator + " executed: session " + sessionID)
				}

			} else if getCommand[0] == "cmd" {

				var cmdString string
				if len(getParams["cmdString"]) > 0 {
					cmdString = getParams["cmdString"][0]
				}
				core.LogData(CurrentOperator + " executed on session " + strconv.Itoa(getActiveSession) + ": cmd " + cmdString)

				if core.Sessions[getActiveSession].CommandQuery != "" {
					data, err := core.ExecuteConnection(core.Sessions[getActiveSession].Rhost, core.Sessions[getActiveSession].Rport, core.Sessions[getActiveSession].Protocol, core.Sessions[getActiveSession].ShellPath, core.Sessions[getActiveSession].CommandQuery, cmdString, core.Sessions[getActiveSession].Query, core.Sessions[getActiveSession].RequestType, "", "")
					if err != nil {
						data = "an error occurred executing the connection, is the shell still up?"
					}

					core.LogData("Session " + strconv.Itoa(getActiveSession) + " returned:\n" + data)

					currentWolf := core.Wolves[CurrentOperator]
					jsonData := `{"data":"` + data + `"}`
					core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)

				} else {
					core.QueueImplantCommand(getActiveSession, cmdString, CurrentOperator)

				}

			} else if getCommand[0] == "kill" {

				var sessionID string
				if len(getParams["id"]) > 0 {
					sessionID = getParams["id"][0]
				}

				session, err := strconv.Atoi(sessionID)

				if err != nil {
					errorString := "wolfpack GET Request could not convert session ID to int, request ignored..."
					core.LogData(errorString)
					returnErr := errors.New(errorString)
					ErrorHandler(returnErr)
					return
				}

				sessionExists := core.SessionExists(session)

				core.LogData(CurrentOperator + " executed: kill " + sessionID)

				var response string

				if !sessionExists {
					response = "Session " + sessionID + " does not exist"
					core.AssignWolfResponse(CurrentOperator, core.Wolves[CurrentOperator].Rhost, response)
				} else {

					delete(core.Sessions, session)

					response = "Session " + sessionID + " has been terminated..."

					core.LogData(response)

					core.AssignWolfResponse(CurrentOperator, core.Wolves[CurrentOperator].Rhost, response)

				}

			} else if getCommand[0] == "load" {

				var sessionID string
				if len(getParams["id"]) > 0 {
					sessionID = getParams["id"][0]
				}

				session, err := strconv.Atoi(sessionID)

				if err != nil {
					errorString := "wolfpack GET Request could not convert session ID to int, request ignored..."
					core.LogData(errorString)
					returnErr := errors.New(errorString)
					ErrorHandler(returnErr)
					return
				}

				response := core.ClientLoadExtendedFunctions(session)

				core.AssignWolfResponse(CurrentOperator, core.Wolves[CurrentOperator].Rhost, string(response))

			} else if getCommand[0] == "upload" {

				if getFileName != "" {
					core.LogData("Session " + strconv.Itoa(getActiveSession) + " returned the file:\n" + getFileName)

					if getFile == "" {
						core.LogData("Session " + strconv.Itoa(getActiveSession) + " file contents was empty, no file written for:\n" + getFileName)
						fmt.Println("\nSession " + strconv.Itoa(getActiveSession) + " file contents was empty, no file written for:\n" + getFileName)
					} else {

						if core.Sessions[getActiveSession].CommandQuery != "" {
							var cmdString = "upload"

							core.LogData(CurrentOperator + " executed on session " + strconv.Itoa(getActiveSession) + ": " + cmdString)

							_, err := core.ExecuteConnection(core.Sessions[getActiveSession].Rhost, core.Sessions[getActiveSession].Rport, core.Sessions[getActiveSession].Protocol, core.Sessions[getActiveSession].ShellPath, core.Sessions[getActiveSession].CommandQuery, cmdString, core.Sessions[getActiveSession].Query, core.Sessions[getActiveSession].RequestType, getFileName, getFile)
							if err != nil {
								errorString := "an error occurred executing the connection, is the shell still up?"
								core.LogData(errorString)
								returnErr := errors.New(errorString)
								ErrorHandler(returnErr)
							}

						} else {
							var cmdString = "upload " + getFileName + " " + getFile

							core.LogData(CurrentOperator + " executed on session " + strconv.Itoa(getActiveSession) + ": " + cmdString)

							core.QueueImplantCommand(getActiveSession, cmdString, CurrentOperator)
						}
					}
				}
			} else if getCommand[0] == "download" {

				if getFileName != "" {
					if core.Sessions[getActiveSession].CommandQuery != "" {

						var cmdString = "download"

						core.LogData("Session " + strconv.Itoa(getActiveSession) + " requested to download the file: " + getFileName)
						core.LogData(CurrentOperator + " executed on session " + strconv.Itoa(getActiveSession) + ": " + cmdString)

						data, err := core.ExecuteConnection(core.Sessions[getActiveSession].Rhost, core.Sessions[getActiveSession].Rport, core.Sessions[getActiveSession].Protocol, core.Sessions[getActiveSession].ShellPath, core.Sessions[getActiveSession].CommandQuery, cmdString, core.Sessions[getActiveSession].Query, core.Sessions[getActiveSession].RequestType, getFileName, "")
						if err != nil {
							data = "an error occurred executing the connection, is the shell still up?"
						}

						core.LogData("Session " + strconv.Itoa(getActiveSession) + " returned:\n" + data)

						currentWolf := core.Wolves[CurrentOperator]
						jsonData := `{"filename":"` + getFileName + `", "file":"` + data + `"}`
						core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
					} else {
						var cmdString = "download " + getFileName
						core.LogData("Session " + strconv.Itoa(getActiveSession) + " requested to download the file: " + getFileName)
						core.LogData(CurrentOperator + " executed on session " + strconv.Itoa(getActiveSession) + ": " + cmdString)
						core.QueueImplantCommand(getActiveSession, cmdString, CurrentOperator)
					}
				}

			}
		} else {
			WolfPackApp.RunCommand(getCommand)
		}

		currentWolf := core.Wolves[getUsername]

		if currentWolf.Response == "" {
			response := map[string]interface{}{
				"response": "",
			}
			json.NewEncoder(w).Encode(response)
		} else {
			response := map[string]interface{}{
				"response": currentWolf.Response,
			}

			json.NewEncoder(w).Encode(response)
			// Clear the response once returned
			core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, "")
		}

		IsWolfPackExec = false
	}

}

func wolfpackChat(data string, wolf core.Wolf, w http.ResponseWriter) {
	chatData := wolf.Username + ": " + data
	core.ChatLog(chatData)
	core.AssignWolfResponse(wolf.Username, wolf.Rhost, chatData)
}
