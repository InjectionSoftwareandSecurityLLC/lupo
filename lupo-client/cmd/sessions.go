package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-client/core"
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

// SessionAppConfig - Primary session nested grumble CLI config construction
// This sets up the lupo "session" nested/sub-prompt and color scheme, defines a history logfile, and toggles various grumble sepcific parameters for help command options.
var SessionAppConfig = &grumble.Config{
	Name:                  "session",
	Description:           "Interactive Session CLI",
	HistoryFile:           ".lupo.history",
	Prompt:                "lupo session " + strconv.Itoa(0) + " ☾ ", // placeholder, will get this value from the server
	PromptColor:           color.New(color.FgGreen, color.Bold),
	HelpHeadlineColor:     color.New(color.FgWhite),
	HelpHeadlineUnderline: true,
	HelpSubCommands:       true,
}

// InitializeSessionCLI - Initialize the nested session CLI arguments
//
// "session" has no arguments and is not a grumble command in and of itself. It is a separate nested grumble application and contains all new base commands.
//
// "session" base commands include:
//
// 	"back" - resets the current active session to "-1" and closes the nested session sub-shell.
//
// 	"session" - the actual "session" command which is used to switch sessions by specifying an argument of a session ID to switch to. This is identical to the interact command only it allows you to switch sessions while in the session sub-shell as "interact" is unavailable in the sub-shell.
//
// 	"cmd" - built in command directive to denote commands that are intended to be executed as a system command of a specified session. These commands are usually sent to the client as JSON in the format of {"cmd":"<some command"}. It supports multi-line/multi-arg commands.
//
// 	"kill" - takes an argument of "id" which is used to de-register the specified session.
//
// 	"load" - will load any additional functions that were registered by an implant. Must be ran each time you interact with a different session unless the implants of those sessions use the same additional functions.
func InitializeSessionCLI(sessionApp *grumble.App, activeSession int) {

	// Send log to server
	//core.LogData(operator + " started interaction with session: " + strconv.Itoa(activeSession))

	backCmd := &grumble.Command{
		Name:     "back",
		Help:     "go back to core lupo cli (or use the exit command)",
		LongHelp: "Exit interactive session cli and return to lupo cli (The 'exit' command is an optional built-in to go back as well) ",
		Run: func(c *grumble.Context) error {
			activeSession = -1

			// Exec to server to send log

			// Exec to server to get listeners list

			reqString := "&isSessionShell=true&command="
			commandString := "back"

			reqString = core.AuthURL + reqString + url.QueryEscape(commandString)

			_, err := core.WolfPackHTTP.Get(reqString)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			sessionApp.Close()

			return nil
		},
	}
	sessionApp.AddCommand(backCmd)

	sessionSwitchCmd := &grumble.Command{
		Name:     "session",
		Help:     "switch to session id",
		LongHelp: "Interact with a different available session by specifying the Session ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Session ID to interact with")
		},
		Run: func(c *grumble.Context) error {
			ActiveSession = c.Args.Int("id")

			// Exec on server to get sessions

			reqString := "&isSessionShell=true&command="
			commandString := "session " + strconv.Itoa(ActiveSession)

			reqString = core.AuthURL + reqString + url.QueryEscape(commandString)

			resp, err := core.WolfPackHTTP.Get(reqString)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			defer resp.Body.Close()

			jsonData, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			// Parse the JSON response
			// We are expecting a JSON string with the key "response" by default, the value is just a raw string response that can be printed to the output
			var coreResponse map[string]interface{}
			err = json.Unmarshal(jsonData, &coreResponse)

			if err != nil {
				//fmt.Println(err)
				return nil
			}

			if coreResponse["response"].(string) == "true" {
				// Close to unload any session specific functions
				sessionApp.Close()

				App = grumble.New(SessionAppConfig)
				App.SetPrompt("lupo session " + strconv.Itoa(ActiveSession) + " ☾ ")
				InitializeSessionCLI(App, ActiveSession)

				grumble.Main(App)

			} else {
				return errors.New("Session " + strconv.Itoa(ActiveSession) + " does not exist")

			}
			return nil
		},
	}

	sessionApp.AddCommand(sessionSwitchCmd)

	sessionCMDCmd := &grumble.Command{
		Name:     "cmd",
		Help:     "execute command on session",
		LongHelp: "Executes a standard OS command that the implant for the current session will execute.",
		Args: func(a *grumble.Args) {
			a.StringList("cmd", "OS Command to be executed by the target session")
		},
		Run: func(c *grumble.Context) error {

			cmd := c.Args.StringList("cmd")

			cmdString := strings.Join(cmd, " ")

			// Exec on server and send command
			reqString := "&isSessionShell=true&command=cmd&activeSession=" + strconv.Itoa(ActiveSession)
			commandString := "&cmdString=" + url.QueryEscape(cmdString)

			reqString = core.AuthURL + reqString + commandString

			_, err := core.WolfPackHTTP.Get(reqString)

			if err != nil {
				fmt.Println(err)
				return nil
			}
			return nil
		},
	}

	sessionApp.AddCommand(sessionCMDCmd)

	sessionKillCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kills a specified session",
		LongHelp: "Kills a session with a specified ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Session ID to kill")
		},
		Run: func(c *grumble.Context) error {

			//id := c.Args.Int("id")

			// Exec on server and get sessions
			/*
				core.LogData(operator + " executed: kill " + strconv.Itoa(id))

				delete(core.Sessions, id)

				warningString := "Session " + strconv.Itoa(id) + " has been terminated..."

				core.LogData(warningString)

				core.WarningColorBold.Println(warningString)
			*/

			return nil
		},
	}

	sessionApp.AddCommand(sessionKillCmd)

	sessionLoadCmd := &grumble.Command{
		Name:     "load",
		Help:     "loads custom functions for a given implant",
		LongHelp: "Loads custom functions registered by an implant tied to the current session if any exist",
		Run: func(c *grumble.Context) error {

			// Exec on server and get sessions

			/*
				for key, value := range core.Sessions[activeSession].Implant.Functions {

					command := key
					info := value.(string)

					implantFunction := &grumble.Command{
						Name: command,
						Help: info,
						Run: func(c *grumble.Context) error {

							core.QueueImplantCommand(activeSession, command)

							return nil
						},
					}

					sessionApp.AddCommand(implantFunction)
					core.LogData("Session " + strconv.Itoa(activeSession) + " loaded extended function: " + command)

				}
			*/

			return nil
		},
	}

	sessionApp.AddCommand(sessionLoadCmd)

}
