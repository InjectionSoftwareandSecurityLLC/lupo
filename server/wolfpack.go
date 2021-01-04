package server

import (
	"errors"
	"net/http"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/core"
)

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

// handleGetRequests - handles any incoming GET requests received by the HTTP(S) listener. Once all values are handled various Implant data update/response routines are executed where relevant based on the provided parameters.
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

func handleWolfPackRequests(w http.ResponseWriter, r *http.Request) {

	// Construct variables for GET URL paramaters
	getParams := r.URL.Query()
	var getPSK string
	var getUsername string
	var getCommand string

	// Get the Remote Address of the Implant from the request
	remoteAddr := r.RemoteAddr

	// Check GET URL parameters and handle errors
	if len(getParams["psk"]) > 0 {
		getPSK = getParams["psk"][0]
	} else {
		returnErr := errors.New("http GET Request did not provide PSK, request ignored")
		ErrorHandler(returnErr)
		return
	}

	if len(getParams["user"]) > 0 {
		getUsername = getParams["user"][0]
	}

	if getPSK != core.Wolves[getUsername].WolfPSK {

		returnErr := errors.New("http GET Request Invalid PSK, request ignored")
		ErrorHandler(returnErr)
		return
	}

	// TODO: User Wolf update routine to execute command and return the response
	core.UpdateWolf(getUsername, getCommand, remoteAddr)

}
