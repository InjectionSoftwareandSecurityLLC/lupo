package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

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
			fmt.Fprintf(table, "ID\tRemote Host\tArch\tProtocol\tLast Check In\tUpdate Interval\tStatus\t\n")
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
				strings.Repeat("=", len("ID")),
				strings.Repeat("=", len("Remote Host")),
				strings.Repeat("=", len("Arch")),
				strings.Repeat("=", len("Protocol")),
				strings.Repeat("=", len("Last Check In")),
				strings.Repeat("=", len("Update Interval")),
				strings.Repeat("=", len("Status")))

			if filterID != -1 {

				_, sessionExists := core.Sessions[filterID]

				if !sessionExists {
					return errors.New("cannot filter show on session " + strconv.Itoa(activeSession) + " because the session does not exist")
				}

				updateInterval := core.Sessions[filterID].Implant.Update
				lastCheckIn := core.Sessions[filterID].RawCheckin

				status, err := calcualateSessionStatus(updateInterval, lastCheckIn)

				var textStatus string

				if err != nil {
					textStatus = "UNKNOWN"
					core.SessionStatusUpdate(filterID, "UNKNOWN")
				} else if status {
					textStatus = core.GreenColorIns("ALIVE")
					core.SessionStatusUpdate(filterID, "ALIVE")
				} else if !status {
					textStatus = core.RedColorIns("DEAD")
					core.SessionStatusUpdate(filterID, "DEAD")
				} else {
					textStatus = core.ErrorColorBoldIns("ERROR")
					core.SessionStatusUpdate(filterID, "ERROR")
				}

				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%f\t%s\t\n",
					strconv.Itoa(core.Sessions[filterID].ID),
					core.Sessions[filterID].Rhost,
					core.Sessions[filterID].Implant.Arch,
					core.Sessions[filterID].Protocol,
					core.Sessions[filterID].Checkin,
					core.Sessions[filterID].Implant.Update,
					textStatus)

			} else {
				for i := range core.Sessions {

					updateInterval := core.Sessions[i].Implant.Update
					lastCheckIn := core.Sessions[i].RawCheckin

					status, err := calcualateSessionStatus(updateInterval, lastCheckIn)

					var textStatus string

					if err != nil {
						textStatus = "UNKNOWN"
						core.SessionStatusUpdate(i, "UNKNOWN")
					} else if status {
						textStatus = core.GreenColorIns("ALIVE")
						core.SessionStatusUpdate(i, "ALIVE")
					} else if !status {
						textStatus = core.RedColorIns("DEAD")
						core.SessionStatusUpdate(i, "DEAD")
					} else {
						textStatus = core.ErrorColorBoldIns("ERROR")
						core.SessionStatusUpdate(i, "ERROR")
					}

					fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%f\t%s\t\n",
						strconv.Itoa(core.Sessions[i].ID),
						core.Sessions[i].Rhost,
						core.Sessions[i].Implant.Arch,
						core.Sessions[i].Protocol,
						core.Sessions[i].Checkin,
						core.Sessions[i].Implant.Update,
						textStatus)
				}
			}

			table.Flush()

			return nil
		},
	}
	interactCmd.AddCommand(showCmd)

	killCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kills a specified session",
		LongHelp: "Kills a session with a specified ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Session ID to kill")
		},
		Run: func(c *grumble.Context) error {

			id := c.Args.Int("id")

			delete(core.Sessions, id)

			core.WarningColorBold.Println("Session " + strconv.Itoa(id) + " has been terminated...")

			return nil
		},
	}
	interactCmd.AddCommand(killCmd)

	cleanCmd := &grumble.Command{
		Name:     "clean",
		Help:     "cleans all sessions marked as DEAD",
		LongHelp: "Kills all sessions marked as DEAD to clear up the session list.",
		Run: func(c *grumble.Context) error {

			for i := range core.Sessions {

				sessionStatus := core.Sessions[i].Status

				if sessionStatus == "DEAD" {
					delete(core.Sessions, i)
					core.WarningColorBold.Println("Session " + strconv.Itoa(i) + " has been terminated...")
				}

			}

			return nil
		},
	}

	interactCmd.AddCommand(cleanCmd)

}

func calcualateSessionStatus(updateInterval float64, lastCheckIn time.Time) (bool, error) {

	if updateInterval == 0 {
		return true, errors.New("No update internal provided, could not be calculated")
	}

	currentTime := time.Now()

	delay := currentTime.Sub(lastCheckIn)

	floatDelay := float64(time.Duration(delay) / time.Second)

	if floatDelay > updateInterval+5 {
		return false, nil
	}

	return true, nil
}
