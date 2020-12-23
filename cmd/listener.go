package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

// Listener - defines a listener structure
type Listener struct {
	id       int
	lhost    string
	lport    int
	protocol string
	instance *http.Server
}

var listeners []Listener
var listenerID int = 0

var errorColorUnderline = color.New(color.FgRed).Add(color.Underline)
var errorColorBold = color.New(color.FgRed).Add(color.Bold)
var successColorBold = color.New(color.FgGreen).Add(color.Bold)

// RemoveListener - Simple slice management function to help remove items from slices
func RemoveListener(l []Listener, index int) []Listener {
	return append(l[:index], l[index:]...)
}

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

			startListener(listenerID, lhost, lport, protocol, listenString)

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
				strings.Repeat("=", len("tProtocol")))

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

			for i := range listeners {
				if listeners[i].id == killID {
					httpServer := listeners[i].instance
					httpServer.Close()
					listeners = append(listeners[:i], listeners[i+1:]...)
					successColorBold.Println("Killing listener: " + strconv.Itoa(killID))
					return nil
				}
			}
			return nil
		},
	}
	listenCmd.AddCommand(listenKillCmd)
}

// startListener - Creates a listener
func startListener(id int, lhost string, lport int, protocol string, listenString string) {

	httpServer := &http.Server{Addr: listenString}

	newListener := Listener{
		id:       id,
		lhost:    lhost,
		lport:    lport,
		protocol: protocol,
		instance: httpServer,
	}

	listeners = append(listeners, newListener)
	successColorBold.Println("Starting listener: " + strconv.Itoa(newListener.id))

	go func(newListener Listener) {
		err := httpServer.ListenAndServe()
		if err != nil {
			println("")
			errorColorBold.Println(err)
			for i := range listeners {
				if listeners[i].id == newListener.id {
					listeners = append(listeners[:i], listeners[i+1:]...)
					listenerID--
					return
				}
			}
			return
		}
	}(newListener)

}

// HTTPServer - Handles HTTPServer requests
func HTTPServer(w http.ResponseWriter, req *http.Request) {

	path := req.URL.Path[1:]

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
	}

	data := string(body)

	switch req.Method {
	case "GET":
		if path != "" {
			log.Println("GET: " + path)
			fmt.Fprintf(w, "%s", path)
		}
	case "POST":
		if data != "" {
			log.Println("POST: " + data)
			fmt.Fprintf(w, "%s", data)
		}
	default:
		fmt.Println("Invalid Request Type")
	}
}
