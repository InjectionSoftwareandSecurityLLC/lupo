package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/core"
)

// PSK - Pre-shared key for implant authentication
var PSK string

// HTTPServerHandler - Handles all HTTPS/HTTPServer requests
func HTTPServerHandler(w http.ResponseWriter, r *http.Request) {
	// Setup webserver attributes like headers and response information
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		handleGetRequests(w, r)
	case "POST":
		handlePostRequests(w, r)
	default:
		// Invalid request type, stay silent don't respond to anything that isn't pre-defined
		return
	}

	return
}

func handleGetRequests(w http.ResponseWriter, r *http.Request) {

	// Construct variables for GET URL paramaters
	getParams := r.URL.Query()
	var getPSK string
	var getSessionID int
	var getUUID uuid.UUID
	var getImplantArch string
	var getUpdate float64
	var getData string
	var getAdditionalFunctions string
	var additionalFunctions map[string]interface{}
	var register bool
	var err error

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

	if len(getParams["register"]) > 0 {
		register, err = strconv.ParseBool(getParams["register"][0])
		if err != nil {
			returnErr := errors.New("http GET Request to register implant was not a valid Boolean, request ignored")
			ErrorHandler(returnErr)
			return
		}
	}

	if len(getParams["sessionID"]) > 0 {
		getSessionID, err = strconv.Atoi(getParams["sessionID"][0])
		if err != nil {
			returnErr := errors.New("http GET Request session ID was not a valid number, request ignored")
			ErrorHandler(returnErr)
			return
		}
	} else {
		getSessionID = -1
	}

	if len(getParams["UUID"]) > 0 {
		getUUID, err = uuid.Parse(getParams["UUID"][0])
		if err != nil {
			returnErr := errors.New("http GET Request UUID was not a UUID, request ignored")
			ErrorHandler(returnErr)
			return
		}
	} else {
		getUUID = core.ZeroedUUID
	}

	if len(getParams["arch"]) > 0 {
		getImplantArch = getParams["arch"][0]
	} else {
		getImplantArch = r.UserAgent()
	}

	if len(getParams["update"]) > 0 {
		getUpdate, err = strconv.ParseFloat(getParams["update"][0], 64)
		if err != nil {
			returnErr := errors.New("http GET Request update internval was not a valid number, request ignored")
			ErrorHandler(returnErr)
			return
		}
	} else {
		getUpdate = 0
	}

	if len(getParams["data"]) > 0 {
		getData = getParams["data"][0]
	} else {
		getData = ""
	}

	if len(getParams["functions"]) > 0 {
		getAdditionalFunctions = getParams["functions"][0]
		json.Unmarshal([]byte(getAdditionalFunctions), &additionalFunctions)
	} else {
		getAdditionalFunctions = ""
	}

	if getPSK == PSK {

		if register == true {

			implant := core.RegisterImplant(getImplantArch, getUpdate, additionalFunctions)

			core.RegisterSession(core.SessionID, "HTTP", implant, remoteAddr)

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			json.NewEncoder(w).Encode(response)

			core.SuccessColorBold.Println("\nNew implant registered successfully!")
			fmt.Println("Session: " + strconv.Itoa(newSession) + " opened")

			return

		}
	} else {
		returnErr := errors.New("http GET Request Invalid PSK, request ignored")
		ErrorHandler(returnErr)
		return
	}

	if core.Sessions[getSessionID].Implant.ID != getUUID || getUUID == core.ZeroedUUID {
		returnErr := errors.New("http GET Request Invalid UUID, request ignored")
		ErrorHandler(returnErr)
		return
	}

	if getData != "" {
		fmt.Println("\nSession " + strconv.Itoa(getSessionID) + " returned:\n" + getData)
	}

	var cmd string

	if core.Sessions[getSessionID].Implant.Commands != nil {
		cmd = core.Sessions[getSessionID].Implant.Commands[0]
	}

	response := map[string]interface{}{
		"cmd": cmd,
	}

	json.NewEncoder(w).Encode(response)

	core.UpdateImplant(getSessionID, getUpdate, additionalFunctions)

	core.SessionCheckIn(getSessionID)
}

func handlePostRequests(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// Construct variables for POST Data paramaters
	postParams := r.Form
	var postPSK string
	var postSessionID int
	var postUUID uuid.UUID
	var postImplantArch string
	var postUpdate float64
	var postData string
	var postAdditionalFunctions string
	var additionalFunctions map[string]interface{}
	var register bool
	var err error

	// Get the Remote Address of the Implant from the request
	remoteAddr := r.RemoteAddr

	// Check POST Data parameters and handle errors
	if len(postParams["psk"]) > 0 {
		postPSK = postParams["psk"][0]
	} else {
		returnErr := errors.New("http POST Request did not provide PSK, request ignored")
		ErrorHandler(returnErr)
		return
	}

	if len(postParams["register"]) > 0 {
		register, err = strconv.ParseBool(postParams["register"][0])
		if err != nil {
			returnErr := errors.New("http POST Request to register implant was not a valid Boolean, request ignored")
			ErrorHandler(returnErr)
			return
		}
	}

	if len(postParams["sessionID"]) > 0 {
		postSessionID, err = strconv.Atoi(postParams["sessionID"][0])
		if err != nil {
			returnErr := errors.New("http POST Request session ID was not a valid number, request ignored")
			ErrorHandler(returnErr)
			return
		}
	} else {
		postSessionID = -1
	}

	if len(postParams["UUID"]) > 0 {
		postUUID, err = uuid.Parse(postParams["UUID"][0])
		if err != nil {
			returnErr := errors.New("http POST Request UUID was not a UUID, request ignored")
			ErrorHandler(returnErr)
			return
		}
	} else {
		postUUID = core.ZeroedUUID
	}

	if len(postParams["arch"]) > 0 {
		postImplantArch = postParams["arch"][0]
	} else {
		postImplantArch = r.UserAgent()
	}

	if len(postParams["update"]) > 0 {
		postUpdate, err = strconv.ParseFloat(postParams["update"][0], 64)
		if err != nil {
			returnErr := errors.New("http POST Request update internval was not a valid number, request ignored")
			ErrorHandler(returnErr)
			return
		}
	} else {
		postUpdate = 0
	}

	if len(postParams["data"]) > 0 {
		postData = postParams["data"][0]
	} else {
		postData = ""
	}

	if len(postParams["functions"]) > 0 {
		postAdditionalFunctions = postParams["functions"][0]
		json.Unmarshal([]byte(postAdditionalFunctions), &additionalFunctions)
	} else {
		postAdditionalFunctions = ""
	}

	if postPSK == PSK {

		if register == true {

			implant := core.RegisterImplant(postImplantArch, postUpdate, additionalFunctions)

			core.RegisterSession(core.SessionID, "HTTP", implant, remoteAddr)

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			json.NewEncoder(w).Encode(response)

			core.SuccessColorBold.Println("\nNew implant registered successfully!")
			fmt.Println("Session: " + strconv.Itoa(newSession) + " opened")

			return

		}
	} else {
		returnErr := errors.New("http POST Request Invalid PSK, request ignored")
		ErrorHandler(returnErr)
		return
	}

	if core.Sessions[postSessionID].Implant.ID != postUUID || postUUID == core.ZeroedUUID {
		returnErr := errors.New("http POST Request Invalid UUID, request ignored")
		ErrorHandler(returnErr)
		return
	}

	if postData != "" {
		fmt.Println("\nSession " + strconv.Itoa(postSessionID) + " returned:\n" + postData)
	}

	var cmd string

	if core.Sessions[postSessionID].Implant.Commands != nil {
		cmd = core.Sessions[postSessionID].Implant.Commands[0]
	}

	response := map[string]interface{}{
		"cmd": cmd,
	}

	json.NewEncoder(w).Encode(response)

	core.UpdateImplant(postSessionID, postUpdate, additionalFunctions)
	core.SessionCheckIn(postSessionID)

}
