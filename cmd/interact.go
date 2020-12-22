package cmd

import (
	"github.com/desertbit/grumble"
)

func init() {

	interactCmd := &grumble.Command{
		Name:     "interact",
		Help:     "interact with a session",
		LongHelp: "Interact with an available session by specifying the Session ID",
		Args: func(a *grumble.Args) {
			a.String("id", "Session ID to interact with")
		},
		Run: func(c *grumble.Context) error {
			println(c.Args.String("id"))
			return nil
		},
	}
	App.AddCommand(interactCmd)
}
