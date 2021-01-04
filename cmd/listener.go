package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/core"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/server"
	"github.com/desertbit/grumble"
)

// psk - Global psk variable utilized by the listener for storing the Pre-Shared key established from the core "lupo" psk flag.
var psk string

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
	id           int
	lhost        string
	lport        int
	protocol     string
	httpInstance *http.Server
	tcpInstance  net.Listener
}

// listeners - a map of Listeners. This is used to manage listeners that are created by the user. The map structure makes it easy to search, add, modify, and delete a large amount of Listeners.
var listeners = make(map[int]Listener)

// listenerID - a global listener ID. Listener IDs are unique and auto-increment on creation. This value is kept track of throughout a Listener's life cycle so it can be incremented/decremented automatically wherever appropriate.
var listenerID int = 0

// init - Initializes the primary "listener" grumble command
//
// "listener" has no arguments and serves as a base for several subcommands.
//
//  "listener" subcommands include:
//
//  	"start" - Starts a listener. Can accept flags of lhost, lport, protocol, key (used in HTTPS), and cert (used in HTTPS) to specify how a listener will establish itself via the protocol given by the user (Defaults to HTTPS).
//
//  	"show" - Prints a table of all running listeners and their configuration details.
//
//  	"kill" - Accepts an argument of "id" that is used to de-register and shutdown a listener server.

func init() {

	listenCmd := &grumble.Command{
		Name:     "listener",
		Help:     "interact with and manage listeners",
		LongHelp: "Interact with and manage an HTTP/HTTPS or TCP Listener",
	}
	App.AddCommand(listenCmd)

	listenStartCmd := &grumble.Command{
		Name:     "start",
		Help:     "start a listener",
		LongHelp: "Starts an HTTP/HTTPS or TCP Listener",
		Flags: func(f *grumble.Flags) {
			f.String("l", "lhost", "127.0.0.1", "listening host IP/Domain")
			f.Int("p", "lport", 1337, "listening host port")
			f.String("x", "protocol", "HTTPS", "protocol to listen on (HTTP, HTTPS, or TCP)")
			f.String("k", "key", "lupo-server.key", "path to TLS private key")
			f.String("c", "cert", "lupo-server.crt", "path to TLS cert")
		},
		Run: func(c *grumble.Context) error {

			lhost := c.Flags.String("lhost")
			lport := c.Flags.Int("lport")
			protocol := c.Flags.String("protocol")
			listenString := lhost + ":" + strconv.Itoa(lport)

			var tlsKey string
			var tlsCert string
			if protocol == "HTTPS" {
				tlsKey = c.Flags.String("key")
				tlsCert = c.Flags.String("cert")
			} else {
				tlsKey = ""
				tlsCert = ""
			}

			psk := c.Flags.String("psk")

			startListener(listenerID, lhost, lport, protocol, listenString, psk, tlsKey, tlsCert)

			listenerID++

			return nil
		},
	}
	listenCmd.AddCommand(listenStartCmd)

	listenShowCommand := &grumble.Command{
		Name:     "show",
		Help:     "show running listeners",
		LongHelp: "Display all running listeners",
		Run: func(c *grumble.Context) error {
			table := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
			fmt.Fprintf(table, "ID\tHost\tPort\tProtocol\t\n")
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t\n",
				strings.Repeat("=", len("ID")),
				strings.Repeat("=", len("Host")),
				strings.Repeat("=", len("Port")),
				strings.Repeat("=", len("Protocol")))

			for i := range listeners {
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t\n",
					strconv.Itoa(listeners[i].id),
					listeners[i].lhost,
					strconv.Itoa(listeners[i].lport),
					listeners[i].protocol)
			}

			table.Flush()
			return nil
		},
	}
	listenCmd.AddCommand(listenShowCommand)

	listenKillCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kill a listener",
		LongHelp: "Kills an HTTP/HTTPS or TCP Listener",
		Args: func(a *grumble.Args) {
			a.Int("id", "Listener ID to kill")
		},
		Run: func(c *grumble.Context) error {

			killID := c.Args.Int("id")

			if _, ok := listeners[killID]; ok {
				if listeners[killID].protocol == "HTTP" || listeners[killID].protocol == "HTTPS" {
					httpServer := listeners[killID].httpInstance
					httpServer.Close()
				} else if listeners[killID].protocol == "TCP" {
					tcpServer := listeners[killID].tcpInstance
					tcpServer.Close()
				}
				delete(listeners, killID)
				core.SuccessColorBold.Println("Killing listener: " + strconv.Itoa(killID))
				return nil
			} else {
				core.ErrorColorBold.Println("Listener: " + strconv.Itoa(killID) + " does not exist")
				return nil
			}

		},
	}
	listenCmd.AddCommand(listenKillCmd)
}

// startListener - Creates a listener based on parameters generated via the "listener start" subcommand.
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
func startListener(id int, lhost string, lport int, protocol string, listenString string, psk string, tlsKey string, tlsCert string) {

	server.PSK = psk

	var newListener Listener

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

		core.SuccessColorBold.Println("Starting listener: " + strconv.Itoa(newListener.id))

		switch protocol {
		case "HTTP":
			go func(newListener Listener) {
				err := newServer.ListenAndServe()
				if err != nil {
					println("")
					core.ErrorColorBold.Println(err)
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
					core.ErrorColorBold.Println(err)
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
			log.Fatal(err)
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

		core.SuccessColorBold.Println("Starting listener: " + strconv.Itoa(newListener.id))

		go server.StartTCPServer(newServer)

	} else {
		core.ErrorColorUnderline.Println("Unsupported listener protocol specified: " + protocol + " is not implemented")
		return
	}

}
