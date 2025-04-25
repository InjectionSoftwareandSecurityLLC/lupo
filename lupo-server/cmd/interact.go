package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/server"
	"github.com/desertbit/grumble"
)

// activeSession - Active session that is being interacted with by the user
//
// This data is supplied as a parameter when switching sessions with either the "interact" command or "session" sub-shell
var activeSession int

// init - Initializes the primary "interact" grumble command
//
// "interact" accepts an argument of "id" that is used to generate a new SessionApp with the SessionAppConfig
//
//  "interact" subcommands include:
//
//  	"show" - Shows all registered sessions. Accepts andargument of "id" that can be used to show a specific session based on the id.
//
//  	"kill" - Accepts an argument of "id" that is used to de-register a session.
//
//  	"clean" - De-registers all sessions marked as "DEAD" based on a pre-determined "Check-In" update interval.

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

			var operator string

			operator = "server"

			_, sessionExists := core.Sessions[activeSession]

			if server.IsWolfPackExec {
				operator = server.CurrentOperator

				core.LogData(operator + " executed: interact " + strconv.Itoa(activeSession))

				currentWolf := core.Wolves[operator]

				if sessionExists {
					core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, "true")
				} else {
					core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, "false")
				}

			} else {
				core.LogData(operator + " executed: interact " + strconv.Itoa(activeSession))

				if !sessionExists {

					errorMessage := "Session " + strconv.Itoa(activeSession) + " does not exist"

					core.LogData("error: " + errorMessage)

					return errors.New(errorMessage)

				}

				App = grumble.New(SessionAppConfig)
				App.SetPrompt("lupo session " + strconv.Itoa(activeSession) + " ☾ ")
				InitializeSessionCLI(App, activeSession)

				grumble.Main(App)

			}

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

			// Pre-caculate session statuses

			filterID := c.Args.Int("id")

			if filterID != -1 {

				_, sessionExists := core.Sessions[filterID]

				if !sessionExists {

					errorMessage := "cannot filter show on session " + strconv.Itoa(filterID) + " because the session does not exist"

					core.LogData("error: " + errorMessage)

					return errors.New(errorMessage)
				}

				updateInterval := core.Sessions[filterID].Implant.Update
				lastCheckIn := core.Sessions[filterID].RawCheckin

				var status bool
				var err error

				session := core.Sessions[filterID]

				connectString := session.Rhost + ":" + strconv.Itoa(session.Rport) + "/" + session.ShellPath

				if session.CommandQuery != "" {
					status, err = core.WebShellStatus(filterID, session.Rhost, session.Rport, session.Protocol, session.RequestType, session.CommandQuery, session.Query, connectString, session.ShellPath)
				} else {
					status, err = calculateSessionStatus(updateInterval, lastCheckIn)
				}

				if err != nil {
					core.SessionStatusUpdate(filterID, "UNKNOWN")
				} else if status {
					core.SessionStatusUpdate(filterID, "ALIVE")
				} else if !status {
					core.SessionStatusUpdate(filterID, "DEAD")
				} else {
					core.SessionStatusUpdate(filterID, "ERROR")
				}

			} else {
				for i := range core.Sessions {

					updateInterval := core.Sessions[i].Implant.Update
					lastCheckIn := core.Sessions[i].RawCheckin

					var status bool
					var err error

					session := core.Sessions[i]

					connectString := session.Rhost + "/" + session.ShellPath

					if session.CommandQuery != "" {
						status, err = core.WebShellStatus(filterID, session.Rhost, session.Rport, session.Protocol, session.RequestType, session.CommandQuery, session.Query, connectString, session.ShellPath)
					} else {
						status, err = calculateSessionStatus(updateInterval, lastCheckIn)
					}

					if err != nil {
						core.SessionStatusUpdate(i, "UNKNOWN")
					} else if status {
						core.SessionStatusUpdate(i, "ALIVE")
					} else if !status {
						core.SessionStatusUpdate(i, "DEAD")
					} else {
						core.SessionStatusUpdate(i, "ERROR")
					}
				}
			}
			var operator string

			operator = "server"

			if server.IsWolfPackExec {
				operator = server.CurrentOperator

				if filterID != -1 {
					core.LogData(operator + " executed: listener show " + strconv.Itoa(filterID))
				} else {
					core.LogData(operator + " executed: listener show")

				}

				currentWolf := core.Wolves[operator]

				sessionMap := core.ShowSessions()

				jsonResp, err := json.Marshal(sessionMap)

				if err != nil {
					return errors.New("could not create JSON response")
				}

				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, string(jsonResp))
			} else {

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

					core.LogData(operator + " executed: interact show" + strconv.Itoa(filterID))

					var textStatus string

					if core.Sessions[filterID].Status == "UNKNOWN" {
						textStatus = "UNKNOWN"
					} else if core.Sessions[filterID].Status == "ALIVE" {
						textStatus = core.GreenColorIns("ALIVE")
					} else if core.Sessions[filterID].Status == "DEAD" {
						textStatus = core.RedColorIns("DEAD")
					} else {
						textStatus = core.ErrorColorBoldIns("ERROR")
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
					core.LogData(operator + " executed: interact show")

					for i := range core.Sessions {

						var textStatus string

						if core.Sessions[i].Status == "UNKNOWN" {
							textStatus = "UNKNOWN"
						} else if core.Sessions[i].Status == "ALIVE" {
							textStatus = core.GreenColorIns("ALIVE")
						} else if core.Sessions[i].Status == "DEAD" {
							textStatus = core.RedColorIns("DEAD")
						} else {
							textStatus = core.ErrorColorBoldIns("ERROR")
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
			}

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

			var operator string

			if server.IsWolfPackExec {
				operator = server.CurrentOperator

				core.LogData(operator + " executed: interact kill" + strconv.Itoa(id))

				currentWolf := core.Wolves[operator]

				_, sessionExists := core.Sessions[id]

				var response string
				if sessionExists {
					delete(core.Sessions, id)
					response = "Session " + strconv.Itoa(id) + " has been terminated..."
				} else {
					response = "Session " + strconv.Itoa(id) + " does not exist..."
				}

				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, response)
			} else {

				operator = "server"

				core.LogData(operator + " executed: interact kill" + strconv.Itoa(id))

				_, sessionExists := core.Sessions[id]

				if sessionExists {
					delete(core.Sessions, id)
					core.WarningColorBold.Println("Session " + strconv.Itoa(id) + " has been terminated...")
				} else {
					core.WarningColorBold.Println("Session " + strconv.Itoa(id) + " does not exist...")

				}

			}
			return nil
		},
	}
	interactCmd.AddCommand(killCmd)

	cleanCmd := &grumble.Command{
		Name:     "clean",
		Help:     "cleans all sessions marked as DEAD",
		LongHelp: "Kills all sessions marked as DEAD to clear up the session list.",
		Run: func(c *grumble.Context) error {

			var operator string

			if server.IsWolfPackExec {
				operator = server.CurrentOperator

				core.LogData(operator + " executed: interact clean")

				currentWolf := core.Wolves[operator]

				var response string
				var isFirstIteration = true

				for i := range core.Sessions {

					sessionStatus := core.Sessions[i].Status

					if sessionStatus == "DEAD" {
						delete(core.Sessions, i)
						if isFirstIteration {
							response += "Session " + strconv.Itoa(i) + " has been terminated..."
						} else {
							response += "\nSession " + strconv.Itoa(i) + " has been terminated..."
						}
						isFirstIteration = false
					}

				}

				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, response)

			} else {
				operator = "server"

				core.LogData(operator + " executed: interact clean")

				for i := range core.Sessions {

					sessionStatus := core.Sessions[i].Status

					if sessionStatus == "DEAD" {
						delete(core.Sessions, i)
						core.WarningColorBold.Println("Session " + strconv.Itoa(i) + " has been terminated...")
					}

				}
			}

			return nil
		},
	}

	interactCmd.AddCommand(cleanCmd)

}

// calculateSessionStatus - Uses an update interval in seconds that is registered by an implant.
//
// The update interval is then compared to the difference in the last "Check-In" time and the current time.
//
// The result of this comparison + a 5 second buffer is checked. If the difference exceeds the expected update interval + 5 the function returns false.
func calculateSessionStatus(updateInterval float64, lastCheckIn time.Time) (bool, error) {

	if updateInterval == 0 {
		return true, errors.New("No update interval provided, could not be calculated")
	}

	currentTime := time.Now()

	delay := currentTime.Sub(lastCheckIn)

	floatDelay := float64(time.Duration(delay) / time.Second)

	if floatDelay > updateInterval+5 {
		return false, nil
	}

	return true, nil
}
