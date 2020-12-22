package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/desertbit/grumble"
)

var listenerID int = 1
var killChannel chan int

func init() {

	listenCmd := &grumble.Command{
		Name:     "listen",
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
			f.String("s", "protocol", "HTTPS", "protocol to listen on (HTTP, HTTPS, or TCP)")
		},
		Run: func(c *grumble.Context) error {
			println(c.Flags.String("lhost"))
			println(c.Flags.Int("lport"))
			println(c.Flags.String("protocol"))

			listenString := c.Flags.String("lhost") + ":" + strconv.Itoa(c.Flags.Int("lport"))

			println(listenerID)
			Listener(listenerID, listenString)
			listenerID++

			return nil
		},
	}
	listenCmd.AddCommand(listenStartCmd)

	listenKillCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kill a listener",
		LongHelp: "Kills an HTTP/HTTPS or TCP Listener",
		Args: func(a *grumble.Args) {
			a.Int("id", "Listener ID to kill")
		},
		Run: func(c *grumble.Context) error {
			killChannel <- c.Args.Int("id")
			killCode := <-killChannel
			println(killCode)
			return nil
		},
	}
	listenCmd.AddCommand(listenKillCmd)
}

// Listener - Creates a listener
func Listener(id int, listenString string) {

	httpServer := &http.Server{Addr: listenString}
	killChannel = make(chan int)

	go func() {
		for {
			killID := <-killChannel
			switch killID {
			case id:
				httpServer.Close()
				return
			default:
				httpServer.ListenAndServe()

			}
		}

	}()
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
