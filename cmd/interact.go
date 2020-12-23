package cmd

import (
	"github.com/InjectionSoftwareandSecurityLLC/lupo/server"
	"github.com/desertbit/grumble"
)

func init() {

	interactCmd := &grumble.Command{
		Name:     "interact",
		Help:     "interact with a session",
		LongHelp: "Interact with an available session by specifying the Session ID",
		Args: func(a *grumble.Args) {
			a.String("id", "Session ID to interact with")
			a.String("cmd", "temporary inline cmd execution for proof of concept until sessions are implemented")
		},
		Run: func(c *grumble.Context) error {
			println(c.Args.String("id"))

			server.CMD = c.Args.String("cmd")

			return nil
		},
	}
	App.AddCommand(interactCmd)
}
