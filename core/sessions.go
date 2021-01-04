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

// Session - defines a session structure composed of:
// id - unique identifier that is autoincremented on creation of a new session
// protocol - the protocol to use when listening for incoming connections. Currenlty supports HTTP(S) and TCP.
// implant - an instance of an Implant that is tied to a session whenever an implant reaches out to register a new session.
// rhost - the "remote" host address. This contains a value of the external IP where an Implant is reaching out from.
// rawcheckin - the raw check in time structure that is calculated anytime an implant communicates successfully with a listener.
// checkin - a formatted version of the rawcheckin in time for easily displaying in print string output so it doesn't need to be converted each time.
// status - current activity status of the implant, can be ALIVE, DEAD, or UNKOWN. UNKOWN is defaulted to if no update interval is provided during implant communications.

type Session struct {
	ID         int
	Protocol   string
	Implant    Implant
	Rhost      string
	RawCheckin time.Time
	Checkin    string
	Status     string
}

// activeSession = global value to keep track of the current active session. Since session "0" is a valid session, this starts at "-1" to determine if no session is active.
var activeSession = -1

// Sessions - map of all sessions. This is used to manage sessions that are registered successfully by implants. The map structure makes it easy to search, add, modify, and delete a large amount of Sessions.
var Sessions = make(map[int]Session)

// SessionID - Global SessionID counter. Session IDs are unique and auto-increment on creation. This value is kept track of throughout a Session's life cycle so it can be incremented/decremented automatically wherever appropriate.
var SessionID int = 0

// SessionAppConfig - Primary session nested grumble CLI config construction
// This sets up the lupo "session" nested/sub-prompt and color scheme, defines a history logfile, and toggles various grumble sepcific parameters for help command options.

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

// RegisterSession - Registers a session and adds it to the session map and increments the global SessionID value
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
// "session" has no arguments and is not a grumble command in and of itself. It is a separate nested grumble application and contains all new base commands.
//  "session" base commands include:
//  	"back" - resets the current active session to "-1" and closes the nested session sub-shell.
//  	"session" - the actual "session" command which is used to switch sessions by specifying an argument of a session ID to switch to. This is identical to the interact command only it allows you to switch sessions while in the session sub-shell as "interact" is unavailable in the sub-shell.
//  	"cmd" - built in command directive to denote commands that are intended to be executed as a system command of a specified session. These commands are usually sent to the client as JSON in the format of {"cmd":"<some command"}. It supports multi-line/multi-arg commands.
//		"kill" - takes an argument of "id" which is used to de-register the specified session.
//		"load" - will load any additional functions that were registered by an implant. Must be ran each time you interact with a different session unless the implants of those sessions use the same additional functions.
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
