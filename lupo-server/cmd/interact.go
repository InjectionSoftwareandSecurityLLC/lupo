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

			operator := "server"

			_, sessionExists := core.Sessions.Load(activeSession)

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
			filterID := c.Args.Int("id")

			if filterID != -1 {
				sessionVal, sessionExists := core.Sessions.Load(filterID)

				if !sessionExists {
					errorMessage := "cannot filter show on session " + strconv.Itoa(filterID) + " because the session does not exist"
					core.LogData("error: " + errorMessage)
					return errors.New(errorMessage)
				}

				session := sessionVal.(core.Session)
				updateInterval := session.Implant.Update
				lastCheckIn := session.RawCheckin

				var status bool
				var err error

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
				} else {
					core.SessionStatusUpdate(filterID, "DEAD")
				}

			} else {
				core.Sessions.Range(func(key, value any) bool {
					sessionID := key.(int)
					session := value.(core.Session)

					updateInterval := session.Implant.Update
					lastCheckIn := session.RawCheckin

					var status bool
					var err error

					connectString := session.Rhost + "/" + session.ShellPath

					if session.CommandQuery != "" {
						status, err = core.WebShellStatus(sessionID, session.Rhost, session.Rport, session.Protocol, session.RequestType, session.CommandQuery, session.Query, connectString, session.ShellPath)
					} else {
						status, err = calculateSessionStatus(updateInterval, lastCheckIn)
					}

					if err != nil {
						core.SessionStatusUpdate(sessionID, "UNKNOWN")
					} else if status {
						core.SessionStatusUpdate(sessionID, "ALIVE")
					} else {
						core.SessionStatusUpdate(sessionID, "DEAD")
					}

					return true
				})
			}

			operator := "server"

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
					sessionVal, _ := core.Sessions.Load(filterID)
					session := sessionVal.(core.Session)

					core.LogData(operator + " executed: interact show " + strconv.Itoa(filterID))

					var textStatus string

					switch session.Status {
					case "UNKNOWN":
						textStatus = "UNKNOWN"
					case "ALIVE":
						textStatus = core.GreenColorIns("ALIVE")
					case "DEAD":
						textStatus = core.RedColorIns("DEAD")
					default:
						textStatus = core.ErrorColorBoldIns("ERROR")
					}

					fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%f\t%s\t\n",
						strconv.Itoa(session.ID),
						session.Rhost,
						session.Implant.Arch,
						session.Protocol,
						session.Checkin,
						session.Implant.Update,
						textStatus)

				} else {
					core.LogData(operator + " executed: interact show")

					core.Sessions.Range(func(key, value any) bool {
						session := value.(core.Session)

						var textStatus string

						switch session.Status {
						case "UNKNOWN":
							textStatus = "UNKNOWN"
						case "ALIVE":
							textStatus = core.GreenColorIns("ALIVE")
						case "DEAD":
							textStatus = core.RedColorIns("DEAD")
						default:
							textStatus = core.ErrorColorBoldIns("ERROR")
						}

						fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%f\t%s\t\n",
							strconv.Itoa(session.ID),
							session.Rhost,
							session.Implant.Arch,
							session.Protocol,
							session.Checkin,
							session.Implant.Update,
							textStatus)

						return true
					})
				}

				table.Flush()
			}

			return nil
		},
	}
	interactCmd.AddCommand(showCmd)

	commandCmd := &grumble.Command{
		Name:     "cmd",
		Help:     "run a command across one or more sessions",
		LongHelp: "Executes a command across one or more sessions based on the filter provided.",
		Args: func(a *grumble.Args) {
			a.String("filter", "Filter sessions by ID, range (e.g., '1-5'), comma-separated IDs, 'all', or Arch (e.g., 'Linux')", grumble.Default("-1"))
			a.StringList("cmd", "OS Command to be executed by the target session")
		},
		Run: func(c *grumble.Context) error {

			filterArg := c.Args.String("filter")
			cmd := c.Args.StringList("cmd")
			cmdString := strings.Join(cmd, " ")

			operator := "server"

			if server.IsWolfPackExec {

				operator = server.CurrentOperator
			}

			var ids []int

			// Helper to add session ID if arch matches
			addSessionIfArchMatches := func(sessionID int, session core.Session) {
				if strings.EqualFold(filterArg, session.Implant.Arch) || filterArg == "all" {
					ids = append(ids, sessionID)
				}
			}
			
			// Determine type of filter
			if filterArg == "-1" {
				return fmt.Errorf("no session filter provided")
			} else if filterArg == "all" || !isNumericFilter(filterArg) {
				// All sessions or Arch string
				core.Sessions.Range(func(key, value any) bool {
					sessionID := key.(int)
					session := value.(core.Session)
					addSessionIfArchMatches(sessionID, session)
					return true
				})
			} else if strings.Contains(filterArg, "-") && isNumericRange(filterArg) {
				// Numeric range, e.g., "5-10"
				parts := strings.Split(filterArg, "-")
				start, _ := strconv.Atoi(parts[0])
				end, _ := strconv.Atoi(parts[1])
				for i := start; i <= end; i++ {
					core.Sessions.Range(func(key, value any) bool {
						if key.(int) == i {
							ids = append(ids, i)
							return false
						}
						return true
					})
				}
			} else {
				// Comma-separated list of numbers
				parts := strings.Split(filterArg, ",")
				for _, p := range parts {
					id, err := strconv.Atoi(strings.TrimSpace(p))
					if err != nil {
						return fmt.Errorf("invalid session filter: %s", p)
					}
					core.Sessions.Range(func(key, value any) bool {
						if key.(int) == id {
							ids = append(ids, id)
							return false
						}
						return true
					})
				}
			}
			// Execute the command for each session
			for _, id := range ids {
				core.LogData(operator + " executed on session " + strconv.Itoa(id) + ": cmd " + cmdString)
				core.CmdExec(id, cmdString, operator)
			}
	
			return nil
		},
	}

	interactCmd.AddCommand(commandCmd)

	killCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kills a specified session",
		LongHelp: "Kills a session with a specified ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Session ID to kill")
		},
		Run: func(c *grumble.Context) error {
			id := c.Args.Int("id")

			operator := "server"

			if server.IsWolfPackExec {
				operator = server.CurrentOperator

				core.LogData(operator + " executed: interact kill " + strconv.Itoa(id))

				currentWolf := core.Wolves[operator]

				_, sessionExists := core.Sessions.Load(id)

				var response string
				if sessionExists {
					core.Sessions.Delete(id)
					response = "Session " + strconv.Itoa(id) + " has been terminated..."
				} else {
					response = "Session " + strconv.Itoa(id) + " does not exist..."
				}

				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, response)
			} else {
				core.LogData(operator + " executed: interact kill " + strconv.Itoa(id))

				_, sessionExists := core.Sessions.Load(id)

				if sessionExists {
					core.Sessions.Delete(id)
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
			operator := "server"

			if server.IsWolfPackExec {
				operator = server.CurrentOperator

				core.LogData(operator + " executed: interact clean")

				currentWolf := core.Wolves[operator]

				var response string
				isFirstIteration := true

				core.Sessions.Range(func(key, value any) bool {
					sessionID := key.(int)
					session := value.(core.Session)

					if session.Status == "DEAD" {
						core.Sessions.Delete(sessionID)
						if isFirstIteration {
							response += "Session " + strconv.Itoa(sessionID) + " has been terminated..."
							isFirstIteration = false
						} else {
							response += "\nSession " + strconv.Itoa(sessionID) + " has been terminated..."
						}
					}
					return true
				})

				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, response)

			} else {
				core.LogData(operator + " executed: interact clean")

				core.Sessions.Range(func(key, value any) bool {
					sessionID := key.(int)
					session := value.(core.Session)

					if session.Status == "DEAD" {
						core.Sessions.Delete(sessionID)
						core.WarningColorBold.Println("Session " + strconv.Itoa(sessionID) + " has been terminated...")
					}
					return true
				})
			}

			return nil
		},
	}

	interactCmd.AddCommand(cleanCmd)
}

// calculateSessionStatus - Uses an update interval in seconds that is registered by an implant.
// The update interval is then compared to the difference in the last "Check-In" time and the current time.
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

// Checks if the string is purely numeric (digits, commas, optional single dash)
func isNumericFilter(s string) bool {
	for _, c := range s {
		if !(c >= '0' && c <= '9') && c != ',' && c != '-' {
			return false
		}
	}
	return true
}

// Checks if the string is a numeric range like "5-10"
func isNumericRange(s string) bool {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return false
	}
	_, err1 := strconv.Atoi(parts[0])
	_, err2 := strconv.Atoi(parts[1])
	return err1 == nil && err2 == nil
}
