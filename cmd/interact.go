package cmd

import (
	"strconv"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/core"
	"github.com/desertbit/grumble"
)

// activeSession - Active session that is being interacted with by the user
var activeSession int

func init() {

	interactCmd := &grumble.Command{
		Name:     "interact",
		Help:     "interact with a session",
		LongHelp: "Interact with an available session by specifying the Session ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Session ID to interact with")
		},
		Run: func(c *grumble.Context) error {

			activeSession = c.Args.Int("id")

			App = grumble.New(core.SessionAppConfig)
			App.SetPrompt("lupo session " + strconv.Itoa(activeSession) + " ☾ ")
			core.InitializeSession(App, activeSession)

			grumble.Main(App)

			return nil
		},
	}
	App.AddCommand(interactCmd)

	showCmd := &grumble.Command{
		Name:     "show",
		Help:     "show all sessions",
		LongHelp: "Show all available session information",
		Args: func(a *grumble.Args) {
			a.Int("id", "Filter on session id", grumble.Default(-1))
		},
		Run: func(c *grumble.Context) error {

			filterID := c.Args.Int("id")

			if filterID != -1 {
				successColorBold.Println("Filtered show executed...")
			} else {
				successColorBold.Println("Unfiltered show executed...")
			}

			return nil
		},
	}
	interactCmd.AddCommand(showCmd)

}
