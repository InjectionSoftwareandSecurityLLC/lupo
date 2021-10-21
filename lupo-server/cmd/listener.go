package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/server"
	"github.com/desertbit/grumble"
)

// psk - Global psk variable utilized by the listener for storing the Pre-Shared key established from the core "lupo" psk flag.
var psk string

// listenerID - a global listener ID. Listener IDs are unique and auto-increment on creation. This value is kept track of throughout a Listener's life cycle so it can be incremented/decremented automatically wherever appropriate.
var listenerID int = 0

// init - Initializes the primary "listener" grumble command
//
// "listener" has no arguments and serves as a base for several subcommands.
//
//  "listener" subcommands include:
//
//		"manage" - Manages global properties of listeners not related to actually establishing a listener, such as PSK rotation
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

	listenManageCmd := &grumble.Command{
		Name:     "manage",
		Help:     "manages global listener attributes",
		LongHelp: "manages global listener attributes such as the PSK",
		Flags: func(f *grumble.Flags) {
			f.String("k", "psk", "", "sets the global PSK to something new to allow for PSK rotation (this will refuse future auth to any implants using the old PSK")
			f.Bool("r", "rand", false, "generates a new random psk when coupled with or omitting an empty psk flag")
		},
		Run: func(c *grumble.Context) error {

			psk := c.Flags.String("psk")
			randPSK := c.Flags.Bool("rand")

			var operator string

			if server.IsWolfPackExec {

				operator = server.CurrentOperator

				response, currentPSK, instruction := core.ManagePSK(psk, randPSK, operator)

				currentWolf := core.Wolves[operator]

				resp := core.ManageResponse{
					Response:    response,
					CurrentPSK:  currentPSK,
					Instruction: instruction,
				}

				jsonResp, err := json.Marshal(resp)

				if err != nil {
					return errors.New("could not creat JSON response")
				}

				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, string(jsonResp))

			} else {
				operator = "server"
				response, currentPSK, instruction := core.ManagePSK(psk, randPSK, operator)

				if instruction == "" {
					core.WarningColorBold.Println(response)
					fmt.Println("")
				} else {
					core.SuccessColorBold.Println(response)
					fmt.Println(currentPSK)
					core.SuccessColorBold.Println(instruction)
					fmt.Println("")
				}
			}

			return nil
		},
	}
	listenCmd.AddCommand(listenManageCmd)

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
			var operator string

			if server.IsWolfPackExec {

				operator = server.CurrentOperator

				core.LogData(operator + " executed: listener start -l " + lhost + " -p " + strconv.Itoa(lport) + " -x " + protocol + " -k " + tlsKey + " -c " + tlsCert)

				response, psk, instructions, help := core.GetFirstUsePSK()

				var resp core.StartResponse

				if response != "" {
					resp = core.StartResponse{
						Response:    response,
						CurrentPSK:  psk,
						Instruction: instructions,
						Help:        help,
						Status:      "",
					}
				}

				status, err := startListener(listenerID, lhost, lport, protocol, listenString, tlsKey, tlsCert)

				if err != nil {
					resp.Status = "error: could not start listener"
				} else {
					resp.Status = status
				}

				currentWolf := core.Wolves[operator]

				jsonResp, err := json.Marshal(resp)

				if err != nil {
					return errors.New("could not creat JSON response")
				}

				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, string(jsonResp))

			} else {
				operator = "server"

				core.LogData(operator + " executed: listener start -l " + lhost + " -p " + strconv.Itoa(lport) + " -x " + protocol + " -k " + tlsKey + " -c " + tlsCert)

				response, psk, instructions, help := core.GetFirstUsePSK()

				if response != "" {
					core.SuccessColorBold.Println(response)
					fmt.Println(psk)
					core.SuccessColorBold.Println(instructions)
					fmt.Println("")
					core.SuccessColorBold.Println(help)

				}
				status, err := startListener(listenerID, lhost, lport, protocol, listenString, tlsKey, tlsCert)

				if err != nil {
					return err
				} else {
					core.SuccessColorBold.Println(status)
				}

			}

			return nil
		},
	}
	listenCmd.AddCommand(listenStartCmd)

	listenShowCommand := &grumble.Command{
		Name:     "show",
		Help:     "show running listeners",
		LongHelp: "Display all running listeners",
		Run: func(c *grumble.Context) error {

			var operator string

			operator = "server"

			if server.IsWolfPackExec {
				operator = server.CurrentOperator

				core.LogData(operator + " executed: listener show")

				currentWolf := core.Wolves[operator]

				listenerMap := core.ShowListeners()

				jsonResp, err := json.Marshal(listenerMap)

				if err != nil {
					return errors.New("could not create JSON response")
				}

				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, string(jsonResp))

			} else {

				core.LogData(operator + " executed: listener show")

				table := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
				fmt.Fprintf(table, "ID\tHost\tPort\tProtocol\t\n")
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t\n",
					strings.Repeat("=", len("ID")),
					strings.Repeat("=", len("Host")),
					strings.Repeat("=", len("Port")),
					strings.Repeat("=", len("Protocol")))

				for i := range core.Listeners {
					fmt.Fprintf(table, "%s\t%s\t%s\t%s\t\n",
						strconv.Itoa(core.Listeners[i].ID),
						core.Listeners[i].Lhost,
						strconv.Itoa(core.Listeners[i].Lport),
						core.Listeners[i].Protocol)
				}
				table.Flush()

			}

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

			var operator string

			operator = "server"

			if server.IsWolfPackExec {
				operator = server.CurrentOperator

				core.LogData(operator + " executed: listener kill " + strconv.Itoa(killID))

				currentWolf := core.Wolves[operator]

				success, fail := core.KillListener(killID)

				if success != "" {
					core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, "true")
				} else if fail != "" {
					core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, "false")
				} else {
					return nil
				}

			} else {
				core.LogData(operator + " executed: listener kill " + strconv.Itoa(killID))

				success, fail := core.KillListener(killID)

				if success != "" {
					core.SuccessColorBold.Println(success)
				} else if fail != "" {
					core.ErrorColorBold.Println(fail)
				} else {
					return nil
				}

			}

			return nil

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
func startListener(id int, lhost string, lport int, protocol string, listenString string, tlsKey string, tlsCert string) (string, error) {

	server.PSK = core.PSK

	var newListener core.Listener

	core.LogData("Starting new " + protocol + " listener on " + listenString)

	if protocol == "HTTP" || protocol == "HTTPS" {
		newServer := &http.Server{Addr: listenString, Handler: http.HandlerFunc(server.HTTPServerHandler)}
		newListener = core.Listener{
			ID:           id,
			Lhost:        lhost,
			Lport:        lport,
			Protocol:     protocol,
			HTTPInstance: newServer,
			TCPInstance:  nil,
		}

		core.Listeners[id] = newListener

		switch protocol {
		case "HTTP":
			go func(newListener core.Listener) {
				err := newServer.ListenAndServe()
				if err != nil {
					println("")
					core.LogData("error: failed to start HTTP server")
					core.ErrorColorBold.Println(err)
					delete(core.Listeners, newListener.ID)
					listenerID--
					return
				}
			}(newListener)
		case "HTTPS":
			go func(newListener core.Listener) {
				err := newServer.ListenAndServeTLS(tlsCert, tlsKey)
				if err != nil {
					println("")
					core.LogData("error: failed to start HTTPS server")
					core.ErrorColorBold.Println(err)
					delete(core.Listeners, newListener.ID)
					listenerID--
					return
				}
			}(newListener)
		default:
			return "", errors.New("failed to start listener due to invalid request type")
		}

	} else if protocol == "TCP" {

		newServer, err := net.Listen("tcp", listenString)
		if err != nil {
			core.LogData("error: failed to start TCP server")
			core.ErrorColorBold.Println(err)
			return "", err
		}
		newListener = core.Listener{
			ID:           id,
			Lhost:        lhost,
			Lport:        lport,
			Protocol:     protocol,
			HTTPInstance: nil,
			TCPInstance:  newServer,
		}

		core.Listeners[id] = newListener

		//core.SuccessColorBold.Println("Starting listener: " + strconv.Itoa(newListener.id))

		go server.StartTCPServer(newServer)

	} else {
		errorString := "Unsupported listener protocol specified: " + protocol + " is not implemented"
		//core.LogData("error: " + errorString)
		core.ErrorColorUnderline.Println(errorString)
		return "", errors.New(errorString)
	}

	listenerID++

	return "Starting listener: " + strconv.Itoa(newListener.ID), nil

}
