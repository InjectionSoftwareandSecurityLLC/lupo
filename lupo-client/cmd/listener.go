package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/desertbit/grumble"
)

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

			//psk := c.Flags.String("psk")
			//randPSK := c.Flags.Bool("rand")

			// Call out to server to generate new PSK

			/*
				if psk == "" {
					if randPSK {
						core.LogData(operator + " executed: listener manage -r true")
						psk = core.GeneratePSK()
						core.SuccessColorBold.Println("Your new random PSK is:")
						fmt.Println(psk)
						core.SuccessColorBold.Println("Embed the PSK into any implants to connect to any listeners in this instance.")
						fmt.Println("")
					} else {
						core.LogData(operator + " executed: listener manage")
						core.WarningColorBold.Println("Warning, you did not provide a PSK, this will keep the current PSK. You can ignore this if you did not want to update the PSK.")
						psk = core.DefaultPSK
					}
				} else {
					core.LogData(operator + " executed: listener manage -k <redacted>")
					core.SuccessColorBold.Println("Your new PSK is:")
					fmt.Println(psk)
					core.SuccessColorBold.Println("Embed the PSK into any implants to connect to any listeners in this instance.")
					fmt.Println("")
				}

				PSK = psk
			*/

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

			//lhost := c.Flags.String("lhost")
			//lport := c.Flags.Int("lport")
			//protocol := c.Flags.String("protocol")
			//listenString := lhost + ":" + strconv.Itoa(lport)

			// Call out to server to start a new listener, consider how to specify new certs whether we will send them upstream or require them to be on the server already

			/*
				var tlsKey string
				var tlsCert string
				if protocol == "HTTPS" {
					tlsKey = c.Flags.String("key")
					tlsCert = c.Flags.String("cert")
				} else {
					tlsKey = ""
					tlsCert = ""


				core.LogData(operator + " executed: listener start -l " + lhost + " -p " + strconv.Itoa(lport) + " -x " + protocol + " -k " + tlsKey + " -c " + tlsCert)

				if PSK == "" && !didDisplayPsk {
					core.SuccessColorBold.Println("Your randomly generated PSK is:")
					fmt.Println(core.DefaultPSK)
					core.SuccessColorBold.Println("Embed the PSK into any implants to connect to any listeners in this instance.")
					fmt.Println("")
					core.SuccessColorBold.Println("If you would like to set your own PSK, you can rotate the current key using the 'listener manage' sub command")
					didDisplayPsk = true
					PSK = core.DefaultPSK
				}

				startListener(listenerID, lhost, lport, protocol, listenString, tlsKey, tlsCert)

				listenerID++

			*/

			return nil
		},
	}
	listenCmd.AddCommand(listenStartCmd)

	listenShowCommand := &grumble.Command{
		Name:     "show",
		Help:     "show running listeners",
		LongHelp: "Display all running listeners",
		Run: func(c *grumble.Context) error {

			// Exec command to server to get list of listeners and output below

			table := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
			fmt.Fprintf(table, "ID\tHost\tPort\tProtocol\t\n")
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t\n",
				strings.Repeat("=", len("ID")),
				strings.Repeat("=", len("Host")),
				strings.Repeat("=", len("Port")),
				strings.Repeat("=", len("Protocol")))

			/*
				for i := range listeners {
					fmt.Fprintf(table, "%s\t%s\t%s\t%s\t\n",
						strconv.Itoa(listeners[i].id),
						listeners[i].lhost,
						strconv.Itoa(listeners[i].lport),
						listeners[i].protocol)
				}
			*/
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

			//killID := c.Args.Int("id")

			// Exec to server to get listeners list

			/*

				if _, ok := listeners[killID]; ok {
					if listeners[killID].protocol == "HTTP" || listeners[killID].protocol == "HTTPS" {
						httpServer := listeners[killID].httpInstance
						httpServer.Close()
					} else if listeners[killID].protocol == "TCP" {
						tcpServer := listeners[killID].tcpInstance
						tcpServer.Close()
					}
					delete(listeners, killID)
					responseMessage := "Killed listener: " + strconv.Itoa(killID)
					core.LogData(responseMessage)
					core.SuccessColorBold.Println(responseMessage)
					return nil
				} else {
					responseMessage := "Listener: " + strconv.Itoa(killID) + " does not exist"
					core.LogData(responseMessage)
					core.ErrorColorBold.Println(responseMessage)
					return nil
				}
			*/

			return nil

		},
	}
	listenCmd.AddCommand(listenKillCmd)
}
