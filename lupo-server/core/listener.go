package core

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

// Listener - defines a listener structure composed of:
//
// id - unique identifier that is autoincremented on creation of a new listener
//
// lhost - the "listening" host address. This tells a listener what interface to listen on based on the address it is tied to.
//
// lport - the "listening" port. This tells a listener what port the lhost of the listener should open to receive connections on.
//
// protocol - the protocol to use when listening for incoming connections. Currenlty supports HTTP(S) and TCP.
//
// httpInstance - a pointer to an instance of the http.Server struct. This is used to reference the core HTTP Server itself when conducting operations such as starting/stopping a listener.
//
// tcpInstance - a copy of the net.Listener struct. This is used to interact with the core TCP Server itself when conducting operations such as starting/stopping a listener.
type Listener struct {
	ID           int
	Lhost        string
	Lport        int
	Protocol     string
	HTTPInstance *http.Server
	TCPInstance  net.Listener
	CryptoPSK    string
}

// ListenerStrings - more loose structure for handling listener data, primarily used to hand off as JSON to the lupo client.
// Contains all the same fields as a Listener structure but as string data types and omits the HTTP/TCPInstance values.
type ListenerStrings struct {
	ID       string
	Lhost    string
	Lport    string
	Protocol string
}

// PSK - global PSK for listeners to manage and set the server PSK
var PSK string

type ManageResponse struct {
	Response    string
	CurrentPSK  string
	Instruction string
}

type StartResponse struct {
	Response    string
	CurrentPSK  string
	Instruction string
	Help        string
	Status      string
}

// DidDisplayPsk - a boolean to check if the pre-generated PSK was already given to the user so it is not printed each time
var DidDisplayPsk = false

// Listeners - a map of Listeners. This is used to manage listeners that are created by the user. The map structure makes it easy to search, add, modify, and delete a large amount of Listeners.
var Listeners = make(map[int]Listener)

func ManagePSK(psk string, isRandom bool, operator string) (response string, currentPSK string, instruction string) {
	if psk == "" && !isRandom {
		LogData(operator + " executed: listener manage")
		response := "Warning, you did not provide a PSK, this will keep the current PSK. You can ignore this if you did not want to update the PSK."
		PSK = DefaultPSK
		currentPSK := ""
		instruction := ""
		return response, currentPSK, instruction
	}
	if isRandom {
		LogData(operator + " executed: listener manage -r true")
		PSK = GeneratePSK()
		response := "Your new random PSK is:"
		currentPSK := PSK
		instruction := "Embed the PSK into any implants to connect to any listeners in this instance."
		return response, currentPSK, instruction
	} else if psk != "" {
		LogData(operator + " executed: listener manage -k <redacted>")
		response := "Your new PSK is:"
		PSK = psk
		currentPSK := PSK
		instruction := "Embed the PSK into any implants to connect to any listeners in this instance."
		return response, currentPSK, instruction
	}

	LogData(operator + " executed: listener manage")
	response = "Warning, you did not provide a PSK, this will keep the current PSK. You can ignore this if you did not want to update the PSK."
	PSK = DefaultPSK
	currentPSK = ""
	instruction = ""
	return response, currentPSK, instruction
}

func GetFirstUsePSK() (response string, psk string, instructions string, help string) {
	if PSK == "" && !DidDisplayPsk {
		response = "Your randomly generated PSK is:"
		psk = DefaultPSK
		instructions = "Embed the PSK into any implants to connect to any listeners in this instance."
		fmt.Println("")
		help = "If you would like to set your own PSK, you can rotate the current key using the 'listener manage' sub command"
		DidDisplayPsk = true
		PSK = DefaultPSK

		return response, psk, instructions, help
	} else {
		return "", "", "", ""
	}
}

// ShowListeners - returns a string map  of Listeners and their details
func ShowListeners() map[string]ListenerStrings {

	var stringListeners = make(map[string]ListenerStrings)

	for i := range Listeners {
		tempListener := ListenerStrings{
			ID:       strconv.Itoa(Listeners[i].ID),
			Lhost:    Listeners[i].Lhost,
			Lport:    strconv.Itoa(Listeners[i].Lport),
			Protocol: Listeners[i].Protocol,
		}
		stringListeners[strconv.Itoa(i)] = tempListener
	}

	return stringListeners
}

// KillListener - kills a listener with the specified id and returns the response
func KillListener(id int) (responseSuccess string, responseFail string) {

	if _, ok := Listeners[id]; ok {
		if Listeners[id].Protocol == "HTTP" || Listeners[id].Protocol == "HTTPS" {
			httpServer := Listeners[id].HTTPInstance
			httpServer.Close()
		} else if Listeners[id].Protocol == "TCP" {
			tcpServer := Listeners[id].TCPInstance
			tcpServer.Close()
		}
		delete(Listeners, id)
		responseMessage := "Killed listener: " + strconv.Itoa(id)
		LogData(responseMessage)
		return responseMessage, ""
	} else {
		responseMessage := "Listener: " + strconv.Itoa(id) + " does not exist"
		LogData(responseMessage)
		return "", responseMessage
	}
}

/*
// StartListener - Creates a listener based on parameters generated via the "listener start" subcommand.
//
// Based on the parameters provided, this function will create a new Listener structure and save it to the listeners map.
//
// Each structure will contain either an HTTP(S) or TCP server instance which is used to start the actual listeners.
//
// HTTP Servers make use of an anonymous goroutine initially to start the listener, but all core handling functions are passed off to the HTTPServerHanlder() function.
//
// TCP Servers are started by executing a StartTCPServer function via goroutine. To maintain concurrency a subsequent goroutine is executed to handle the data for all TCP connections via TCPServerHandler() function.
//
// All listeners are concurrent and support multiple simultaneous connections.
func StartListener(id int, lhost string, lport int, protocol string, listenString string, tlsKey string, tlsCert string) {

	server.PSK = PSK

	var newListener Listener

	LogData("Starting new " + protocol + " listener on " + listenString)

	if protocol == "HTTP" || protocol == "HTTPS" {
		newServer := &http.Server{Addr: listenString, Handler: http.HandlerFunc(server.HTTPServerHandler)}
		newListener = Listener{
			id:           id,
			lhost:        lhost,
			lport:        lport,
			protocol:     protocol,
			httpInstance: newServer,
			tcpInstance:  nil,
		}

		listeners[id] = newListener

		switch protocol {
		case "HTTP":
			go func(newListener Listener) {
				err := newServer.ListenAndServe()
				if err != nil {
					println("")
					LogData("error: failed to start HTTP server")
					ErrorColorBold.Println(err)
					delete(listeners, newListener.id)
					listenerID--
					return
				}
			}(newListener)
		case "HTTPS":
			go func(newListener Listener) {
				err := newServer.ListenAndServeTLS(tlsCert, tlsKey)
				if err != nil {
					println("")
					LogData("error: failed to start HTTPS server")
					ErrorColorBold.Println(err)
					delete(listeners, newListener.id)
					listenerID--
					return
				}
			}(newListener)
		default:
			// Invalid request type, stay silent don't respond to anything that isn't pre-defined
			return
		}

	} else if protocol == "TCP" {

		newServer, err := net.Listen("tcp", listenString)
		if err != nil {
			LogData("error: failed to start TCP server")
			ErrorColorBold.Println(err)
			return
		}
		newListener = Listener{
			id:           id,
			lhost:        lhost,
			lport:        lport,
			protocol:     protocol,
			httpInstance: nil,
			tcpInstance:  newServer,
		}

		listeners[id] = newListener

		SuccessColorBold.Println("Starting listener: " + strconv.Itoa(newListener.id))

		go server.StartTCPServer(newServer)

	} else {
		errorString := "Unsupported listener protocol specified: " + protocol + " is not implemented"
		LogData("error: " + errorString)
		ErrorColorUnderline.Println(errorString)
		return
	}

}*/
