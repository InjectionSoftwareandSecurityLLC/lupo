package core

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

// Define custom colors for text output
var errorColorUnderline = color.New(color.FgRed).Add(color.Underline)
var errorColorBold = color.New(color.FgRed).Add(color.Bold)
var successColorBold = color.New(color.FgGreen).Add(color.Bold)

// Session - defines a session structure for Lupo session handling
type Session struct {
	ID       int
	Protocol string
	Implant  Implant
	Rhost    string
	Checkin  string
	Status   string
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

	t := time.Now()
	timeFormatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	Sessions[sessionID] = Session{
		ID:       sessionID,
		Protocol: protocol,
		Implant:  implant,
		Rhost:    rhost,
		Checkin:  timeFormatted,
		Status:   "TEMP",
	}

	SessionID++
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
			a.String("cmd", "OS Command to be executed by the target session")
		},
		Run: func(c *grumble.Context) error {

			cmd := c.Args.String("cmd")

			var sessionUpdate = Sessions[activeSession]

			sessionUpdate.Implant.Command = cmd

			Sessions[activeSession] = sessionUpdate

			return nil
		},
	}

	sessionApp.AddCommand(sessionCMDCmd)
}
