package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

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

			_, sessionExists := core.Sessions[activeSession]

			if !sessionExists {
				return errors.New("Session " + strconv.Itoa(activeSession) + " does not exist")
			}

			App = grumble.New(core.SessionAppConfig)
			App.SetPrompt("lupo session " + strconv.Itoa(activeSession) + " ☾ ")
			core.InitializeSessionCLI(App, activeSession)

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

			table := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
			fmt.Fprintf(table, "ID\tRemote Host\tArch\tProtocol\tLast Check In\tStatus\t\n")
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t\n",
				strings.Repeat("=", len("ID")),
				strings.Repeat("=", len("Remote Host")),
				strings.Repeat("=", len("Arch")),
				strings.Repeat("=", len("Protocol")),
				strings.Repeat("=", len("Last Check In")),
				strings.Repeat("=", len("Status")))

			if filterID != -1 {
				successColorBold.Println("Filtered show executed...")

				_, sessionExists := core.Sessions[filterID]

				if !sessionExists {
					return errors.New("Cannot filter show on session " + strconv.Itoa(activeSession) + " because the session does not exist")
				}

				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t\n",
					strconv.Itoa(core.Sessions[filterID].ID),
					core.Sessions[filterID].Rhost,
					core.Sessions[filterID].Implant.Arch,
					core.Sessions[filterID].Protocol,
					core.Sessions[filterID].Checkin,
					core.Sessions[filterID].Status)

			} else {
				successColorBold.Println("Unfiltered show executed...")
				for i := range core.Sessions {
					fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t\n",
						strconv.Itoa(core.Sessions[i].ID),
						core.Sessions[i].Rhost,
						core.Sessions[i].Implant.Arch,
						core.Sessions[i].Protocol,
						core.Sessions[i].Checkin,
						core.Sessions[i].Status)
				}
			}

			table.Flush()

			return nil
		},
	}
	interactCmd.AddCommand(showCmd)

}
