package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/core"
)

// StartTCPServer - starts a tcp server with given parameters
func StartTCPServer(tcpServer net.Listener) {
	defer tcpServer.Close()

	for {
		conn, err := tcpServer.Accept()
		if err != nil {
			if errType, ok := err.(*net.OpError); ok && errType.Op == "accept" {
				break
			}
			log.Println(err)
			continue
			// Print the error using a log.Fatal would exit the server
		}
		// Using a go routine to handle the connection
		go TCPServerHandler(conn)
	}
}

// TCPServerHandler - handles incoming TCP connections
func TCPServerHandler(conn net.Conn) {

	defer conn.Close()

	var tcpParams core.TCPData
	var remoteAddr string
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		remoteAddr = addr.IP.String()
	}
	var additionalFunctions map[string]interface{}

	netData, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		core.ErrorColorBold.Println("Error reading TCP connection from implant claiming to be session: " + strconv.Itoa(tcpParams.SessionID))
		fmt.Println(err)
		return
	}

	err = json.Unmarshal([]byte(netData), &tcpParams)

	if err != nil {
		core.ErrorColorBold.Println("There was an error with parsing input from a TCP based implant, check the error below:")
		fmt.Println(err)
	}

	if tcpParams.PSK == "" {
		returnErr := errors.New("TCP Request did not provide PSK, request ignored")
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
				core.ErrorColorBold.Println("Error converting TCP response to JSON")
			}

			conn.Write([]byte(jsonResp))

			core.SuccessColorBold.Println("\nNew implant registered successfully!")
			fmt.Println("Session: " + strconv.Itoa(newSession) + " established")

			return

		}
	} else {
		returnErr := errors.New("TCP Request Invalid PSK, request ignored")
		ErrorHandler(returnErr)
		return
	}

	if core.Sessions[tcpParams.SessionID].Implant.ID != tcpParams.UUID || tcpParams.UUID == core.ZeroedUUID {
		returnErr := errors.New("TCP Request Invalid UUID, request ignored")
		ErrorHandler(returnErr)
		return
	}

	if tcpParams.Data != "" {
		fmt.Println("\nSession " + strconv.Itoa(tcpParams.SessionID) + " returned:\n" + tcpParams.Data)
	}

	var cmd string

	if core.Sessions[tcpParams.SessionID].Implant.Commands != nil {
		cmd = core.Sessions[tcpParams.SessionID].Implant.Commands[0]
	}

	response := map[string]interface{}{
		"cmd": cmd,
	}

	jsonResp, err := json.Marshal(response)

	if err != nil {
		core.ErrorColorBold.Println("Error converting TCP cmd to JSON")
	}

	core.UpdateImplant(tcpParams.SessionID, tcpParams.Update, additionalFunctions)
	core.SessionCheckIn(tcpParams.SessionID)

	conn.Write([]byte(jsonResp))

}
