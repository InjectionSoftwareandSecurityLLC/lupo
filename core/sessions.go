package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

// Session - defines a session structure for Lupo session handling
type Session struct {
	ID         int
	Protocol   string
	Implant    Implant
	Rhost      string
	RawCheckin time.Time
	Checkin    string
	Status     string
}

var activeSession = -1

// Sessions - map of all sessions
var Sessions = make(map[int]Session)

// SessionID - Global SessionID counter
var SessionID int = 0

// SessionAppConfig - Primary session nested grumble CLI config construction
var SessionAppConfig = &grumble.Config{
	Name:                  "session",
	Description:           "Interactive Session CLI",
	HistoryFile:           "/tmp/lupo.log",
	Prompt:                "lupo session " + strconv.Itoa(activeSession) + " ☾ ",
	PromptColor:           color.New(color.FgGreen, color.Bold),
	HelpHeadlineColor:     color.New(color.FgWhite),
	HelpHeadlineUnderline: true,
	HelpSubCommands:       true,
}

// RegisterSession - Registers a session and adds it to the session map
func RegisterSession(sessionID int, protocol string, implant Implant, rhost string) {

	currentTime := time.Now()
	timeFormatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second())

	Sessions[sessionID] = Session{
		ID:         sessionID,
		Protocol:   protocol,
		Implant:    implant,
		Rhost:      rhost,
		RawCheckin: currentTime,
		Checkin:    timeFormatted,
		Status:     "ALIVE",
	}

	SessionID++
}

// SessionCheckIn - Updates the Last Check In anytime a verified session calls back
func SessionCheckIn(sessionID int) {
	currentTime := time.Now()
	timeFormatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second())

	var sessionUpdate = Sessions[sessionID]

	sessionUpdate.RawCheckin = currentTime
	sessionUpdate.Checkin = timeFormatted

	Sessions[sessionID] = sessionUpdate
}

// SessionStatusUpdate - Updates the current status of a session
func SessionStatusUpdate(sessionID int, status string) {

	var sessionUpdate = Sessions[sessionID]

	sessionUpdate.Status = status

	Sessions[sessionID] = sessionUpdate
}

// InitializeSessionCLI - Initialize the nested session CLI arguments
func InitializeSessionCLI(sessionApp *grumble.App, activeSession int) {

	backCmd := &grumble.Command{
		Name:     "back",
		Help:     "go back to core lupo cli (or use the exit command)",
		LongHelp: "Exit interactive session cli and return to lupo cli (The 'exit' command is an optional built-in to go back as well) ",
		Run: func(c *grumble.Context) error {
			activeSession = -1

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

			_, sessionExists := Sessions[activeSession]

			if !sessionExists {
				return errors.New("Session " + strconv.Itoa(activeSession) + " does not exist")
			}

			sessionApp.SetPrompt("lupo session " + strconv.Itoa(activeSession) + " ☾ ")

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

			QueueImplantCommand(activeSession, cmdString)

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

			delete(Sessions, id)

			WarningColorBold.Println("Session " + strconv.Itoa(id) + " has been terminated...")

			return nil
		},
	}

	sessionApp.AddCommand(sessionKillCmd)

	sessionLoadCmd := &grumble.Command{
		Name:     "load",
		Help:     "loads custom functions for a given implant",
		LongHelp: "Loads custom functions registered by an implant tied to the current session if any exist",
		Run: func(c *grumble.Context) error {
			for key, value := range Sessions[activeSession].Implant.Functions {

				command := key
				info := value.(string)

				implantFunction := &grumble.Command{
					Name: command,
					Help: info,
					Run: func(c *grumble.Context) error {

						QueueImplantCommand(activeSession, command)

						return nil
					},
				}

				sessionApp.AddCommand(implantFunction)

			}

			return nil
		},
	}

	sessionApp.AddCommand(sessionLoadCmd)

}
