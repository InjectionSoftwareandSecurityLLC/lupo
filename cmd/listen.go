package cmd

import (
	"github.com/desertbit/grumble"
)

func init() {

	listenCmd := &grumble.Command{
		Name:     "listen",
		Help:     "start a listener",
		LongHelp: "Starts an HTTP/HTTPS or TCP Listener",
		Args: func(a *grumble.Args) {
			a.String("--lhost", "listening host IP/Domain", grumble.Default("127.0.0.1"))
			a.String("--lport", "listening host port", grumble.Default("1337"))
			a.String("--protocol", "protocol to listen on (HTTP, HTTPS, or TCP)", grumble.Default("HTTPS"))
		},
		Run: func(c *grumble.Context) error {
			println(c.Args.String("lhost"))
			println(c.Args.String("lport"))
			println(c.Args.String("protocol"))
			return nil
		},
	}
	App.AddCommand(listenCmd)
}
