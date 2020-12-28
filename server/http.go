package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

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

	//path := r.URL.Path[1:]

	getParams := r.URL.Query()
	var getPSK string
	var getSessionID int
	var getImplantArch string
	var getAdditionalFunctions string
	var regsiter bool
	var err error

	getRemoteAddr := r.RemoteAddr

	if len(getParams["psk"]) > 0 {
		getPSK = getParams["psk"][0]
	} else {
		errorColorBold.Println("GET Request: Implant Did Not Provide PSK")
		return
	}

	if len(getParams["register"]) > 0 {
		regsiter, err = strconv.ParseBool(getParams["register"][0])

		if err != nil {
			errorColorBold.Println("Register param passed, but type was not Boolean, ignored request")
			return
		}
	} else {
		errorColorBold.Println("Temp error - just means agent didn't request to be registered")
	}

	if len(getParams["sessionID"]) > 0 {
		getSessionID, err = strconv.Atoi(getParams["sessionID"][0])
		if err != nil {
			errorColorBold.Println("Session ID provided by agent was not a number")
			return
		}
	} else {
		getSessionID = -1
		errorColorBold.Println("Temp error - just means agent didn't provide session id with the request")
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
		errorColorBold.Println("Temp error - just means agent didn't provide additional functions with the request")
	}

	if getPSK == PSK {

		if regsiter == true {

			implant := core.RegisterImplant(getImplantArch, getAdditionalFunctions)

			core.RegisterSession(core.SessionID, "HTTP", implant, getRemoteAddr)

			fmt.Fprintf(w, "%s", strconv.Itoa(core.SessionID-1))
			return

		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}

		data := string(body)

		switch r.Method {
		case "GET":
			//log.Println("GET: " + path)
			fmt.Fprintf(w, "%s", core.Sessions[getSessionID].Implant.Command)

			var sessionUpdate = core.Sessions[getSessionID]

			sessionUpdate.Implant.Command = ""

			core.Sessions[getSessionID] = sessionUpdate

		case "POST":
			if data != "" {
				log.Println("POST: " + data)
				fmt.Fprintf(w, "%s", data)
			}
		default:
			fmt.Println("Invalid Request Type")
		}
	} else {
		errorColorBold.Println("Implant Failed PSK Check")
	}
}
