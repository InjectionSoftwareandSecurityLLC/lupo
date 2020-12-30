package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/server"
	"github.com/desertbit/grumble"
)

var psk string

// Listener - defines a listener structure
type Listener struct {
	id       int
	lhost    string
	lport    int
	protocol string
	instance *http.Server
}

var listeners = make(map[int]Listener)
var listenerID int = 0

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
			f.String("x", "protocol", "HTTP", "protocol to listen on (HTTP, HTTPS, or TCP)") //Temporarily default to HTTP change to HTTPS once implemented
		},
		Run: func(c *grumble.Context) error {

			lhost := c.Flags.String("lhost")
			lport := c.Flags.Int("lport")
			protocol := c.Flags.String("protocol")
			listenString := lhost + ":" + strconv.Itoa(lport)
			psk := c.Flags.String("psk")

			startListener(listenerID, lhost, lport, protocol, listenString, psk)

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
			httpServer := listeners[killID].instance
			httpServer.Close()
			delete(listeners, killID)
			successColorBold.Println("Killing listener: " + strconv.Itoa(killID))
			return nil
		},
	}
	listenCmd.AddCommand(listenKillCmd)
}

// startListener - Creates a listener
func startListener(id int, lhost string, lport int, protocol string, listenString string, psk string) {

	server.PSK = psk

	httpServer := &http.Server{Addr: listenString, Handler: http.HandlerFunc(server.HTTPServerHandler)}

	newListener := Listener{
		id:       id,
		lhost:    lhost,
		lport:    lport,
		protocol: protocol,
		instance: httpServer,
	}

	listeners[id] = newListener

	successColorBold.Println("Starting listener: " + strconv.Itoa(newListener.id))

	go func(newListener Listener) {
		err := httpServer.ListenAndServe()
		if err != nil {
			println("")
			errorColorBold.Println(err)
			delete(listeners, newListener.id)
			listenerID--
			return
		}
	}(newListener)

}
