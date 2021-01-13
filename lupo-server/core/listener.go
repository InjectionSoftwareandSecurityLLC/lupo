package core

import "fmt"

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
