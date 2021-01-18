package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
// PSK - the client Pre-Shared Key that the the implant will send to be compared for authentication to the server PSK
//
// Username - a unique Username that is defined by the operator administering wolfpack server users. This is sent to identify what user is connecting.
//
// Command - a command issued by a user to be transmitted and executed by the Lupo server
//
// Polling - boolean status indicator to know if the incoming request was from the Lupo client polling functions or the user
//
// ActiveSession - the active session an operator is interacting with when executing commands, only applies to session sub-shell/nested shell commands

func handleWolfPackRequests(w http.ResponseWriter, r *http.Request) {

	// Construct variables for GET URL paramaters
	getParams := r.URL.Query()
	var getPSK string
	var getUsername string
	var getCommand []string
	var getPolling = false
	var getIsSessionShell = false
	var getActiveSession int

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

			fmt.Println(getParams["activeSession"][0])
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

	if getPSK != core.Wolves[getUsername].WolfPSK {
		errorString := "wolfpack GET Request Invalid PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if getPolling {
		currentWolf := core.Wolves[getUsername]

		if currentWolf.Broadcast != "" {

			fmt.Println(currentWolf.Broadcast)

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

				session, err := strconv.Atoi(getCommand[1])

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
					core.LogData(CurrentOperator + " executed: session " + getCommand[1])

				} else {
					core.AssignWolfResponse(CurrentOperator, core.Wolves[CurrentOperator].Rhost, "false")
					core.LogData(CurrentOperator + " executed: session " + getCommand[1])
				}

			} else if getCommand[0] == "cmd" {

				var cmdString string
				if len(getParams["cmdString"]) > 0 {
					cmdString = getParams["cmdString"][0]
				}
				core.LogData(CurrentOperator + " executed on session " + strconv.Itoa(getActiveSession) + ": cmd " + cmdString)
				core.QueueImplantCommand(getActiveSession, cmdString, CurrentOperator)

			}
		} else {
			WolfPackApp.RunCommand(getCommand)

		}

		currentWolf := core.Wolves[getUsername]

		fmt.Println(currentWolf.Response)

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
