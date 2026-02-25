package core

import (
	"net/http"
	"strconv"
)

// Stager - defines a stager structure composed of:
//
// ID - unique identifier that is autoincremented on creation of a new stager
//
// Lhost - the "listening" host address. This tells the stager what interface to listen on.
//
// Lport - the "listening" port.
//
// Protocol - the protocol to use when serving files. Supports HTTP and HTTPS.
//
// Dir - the directory path to serve files from. Created on startup if it does not exist.
//
// HTTPInstance - a pointer to the underlying http.Server used to start/stop the stager.
type Stager struct {
	ID           int
	Lhost        string
	Lport        int
	Protocol     string
	Dir          string
	HTTPInstance *http.Server
}

// StagerStrings - more loose structure for handling stager data, primarily used to hand off as JSON to the lupo client.
// Contains all the same fields as a Stager structure but as string data types and omits the HTTPInstance value.
type StagerStrings struct {
	ID       string
	Lhost    string
	Lport    string
	Protocol string
	Dir      string
}

// Stagers - a map of Stagers. This is used to manage stagers that are created by the user.
var Stagers = make(map[int]Stager)

// ShowStagers - returns a string map of Stagers and their details
func ShowStagers() map[string]StagerStrings {
	var stringStagers = make(map[string]StagerStrings)

	for i := range Stagers {
		tempStager := StagerStrings{
			ID:       strconv.Itoa(Stagers[i].ID),
			Lhost:    Stagers[i].Lhost,
			Lport:    strconv.Itoa(Stagers[i].Lport),
			Protocol: Stagers[i].Protocol,
			Dir:      Stagers[i].Dir,
		}
		stringStagers[strconv.Itoa(i)] = tempStager
	}

	return stringStagers
}

// KillStager - kills a stager with the specified id and returns the response
func KillStager(id int) (responseSuccess string, responseFail string) {
	if _, ok := Stagers[id]; ok {
		httpServer := Stagers[id].HTTPInstance
		httpServer.Close()
		delete(Stagers, id)
		responseMessage := "Killed stager: " + strconv.Itoa(id)
		LogData(responseMessage)
		return responseMessage, ""
	} else {
		responseMessage := "Stager: " + strconv.Itoa(id) + " does not exist"
		LogData(responseMessage)
		return "", responseMessage
	}
}
