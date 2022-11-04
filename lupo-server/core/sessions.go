package core

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/desertbit/grumble"
)

// Session - defines a session structure composed of:
//
// id - unique identifier that is autoincremented on creation of a new session
//
// protocol - the protocol to use when listening for incoming connections. Currenlty supports HTTP(S) and TCP.
//
// implant - an instance of an Implant that is tied to a session whenever an implant reaches out to register a new session.
//
// rhost - the "remote" host address. This contains a value of the external IP where an Implant is reaching out from.
//
// rawcheckin - the raw check in time structure that is calculated anytime an implant communicates successfully with a listener.
//
// checkin - a formatted version of the rawcheckin in time for easily displaying in print string output so it doesn't need to be converted each time.
//
// status - current activity status of the implant, can be ALIVE, DEAD, or UNKNOWN. UNKNOWN is defaulted to if no update interval is provided during implant communications.

type Session struct {
	ID           int
	Protocol     string
	Implant      Implant
	Rhost        string
	RawCheckin   time.Time
	Checkin      string
	Status       string
	Rport        int
	CommandQuery string
	Query        string
	RequestType  string
	ShellPath    string
}

// SessionStrings - more loose structure for handling session data, primarily used to hand off as JSON to the lupo client.
// Contains all the same fields as a Session structure but as string data types and omits the HTTP/TCPInstance values.
type SessionStrings struct {
	ID            string
	Protocol      string
	ImplantArch   string
	ImplantUpdate string
	Rhost         string
	RawCheckin    string
	Checkin       string
	Status        string
	Rport         string
	CommandQuery  string
	Query         string
	RequestType   string
	ShellPath     string
}

// ActiveSession = global value to keep track of the current active session. Since session "0" is a valid session, this starts at "-1" to determine if no session is active.
var ActiveSession = -1

// Sessions - map of all sessions. This is used to manage sessions that are registered successfully by implants. The map structure makes it easy to search, add, modify, and delete a large amount of Sessions.
var Sessions = make(map[int]Session)

// SessionID - Global SessionID counter. Session IDs are unique and auto-increment on creation. This value is kept track of throughout a Session's life cycle so it can be incremented/decremented automatically wherever appropriate.
var SessionID int = 0

// RegisterSession - Registers a session and adds it to the session map and increments the global SessionID value
func RegisterSession(sessionID int, protocol string, implant Implant, rhost string, rport int, command string, query string, requestType string, shellpath string) {

	currentTime := time.Now()
	timeFormatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second())

	if rport != 0 {
		Sessions[sessionID] = Session{
			ID:           sessionID,
			Protocol:     protocol,
			Implant:      implant,
			Rhost:        rhost + ":" + strconv.Itoa(rport),
			RawCheckin:   currentTime,
			Checkin:      timeFormatted,
			Status:       "ALIVE",
			Rport:        rport,
			CommandQuery: command,
			Query:        query,
			RequestType:  requestType,
			ShellPath:    shellpath,
		}
	} else {
		Sessions[sessionID] = Session{
			ID:           sessionID,
			Protocol:     protocol,
			Implant:      implant,
			Rhost:        rhost,
			RawCheckin:   currentTime,
			Checkin:      timeFormatted,
			Status:       "ALIVE",
			Rport:        rport,
			CommandQuery: command,
			Query:        query,
			RequestType:  requestType,
			ShellPath:    shellpath,
		}
	}

	SessionID++

	LogData("Registered new session with ID: " + strconv.Itoa(sessionID))
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

	LogData("Session " + strconv.Itoa(sessionID) + " checked in")
}

// SessionStatusUpdate - Updates the current status of a session
func SessionStatusUpdate(sessionID int, status string) {

	var sessionUpdate = Sessions[sessionID]

	sessionUpdate.Status = status

	Sessions[sessionID] = sessionUpdate

	LogData("Updated status of session " + strconv.Itoa(sessionID) + " session is: " + status)
}

// BroadcastSession - Broadcast a message that a new session has been established
func BroadcastSession(session string) {

	successMessage := "New implant registered successfully!"
	message := "Session: " + session + " established"
	SuccessColorBold.Println("\n" + successMessage)
	LogData(message)
	fmt.Println(message)

	for key := range Wolves {
		broadcast := `{"successMessage":"` + successMessage + `","message":"` + message + `"}`
		AssignWolfBroadcast(Wolves[key].Username, Wolves[key].Rhost, broadcast)
	}
}

// ShowSessions - returns a map of Sessions and their details
func ShowSessions() map[string]SessionStrings {
	var stringSessions = make(map[string]SessionStrings)

	for i := range Sessions {
		tempSession := SessionStrings{
			ID:            strconv.Itoa(Sessions[i].ID),
			Protocol:      Sessions[i].Protocol,
			ImplantArch:   Sessions[i].Implant.Arch,
			ImplantUpdate: strconv.FormatFloat(Sessions[i].Implant.Update, 'f', -1, 64),
			Rhost:         Sessions[i].Rhost,
			RawCheckin:    Sessions[i].RawCheckin.String(),
			Checkin:       Sessions[i].Checkin,
			Status:        Sessions[i].Status,
		}
		stringSessions[strconv.Itoa(i)] = tempSession
	}

	return stringSessions
}

// SessionExists - returns if a session exists or not
func SessionExists(session int) bool {

	_, sessionExists := Sessions[session]

	return sessionExists
}

// LoadExtendedFunctions - Loads the functions registered by an implant
func LoadExtendedFunctions(sessionApp *grumble.App, activeSession int) {
	for key, value := range Sessions[activeSession].Implant.Functions {

		command := key
		info := value.(string)

		implantFunction := &grumble.Command{
			Name: command,
			Help: info,
			Run: func(c *grumble.Context) error {

				QueueImplantCommand(activeSession, command, "server")

				return nil
			},
		}

		sessionApp.AddCommand(implantFunction)
		LogData("Session " + strconv.Itoa(activeSession) + " loaded extended function: " + command)

	}
}

// ClientLoadExtendedFunctions - Loads the functions registered by an implant and returns those functions for the lupo client to load
func ClientLoadExtendedFunctions(activeSession int) []byte {

	sessionFunctions, err := json.Marshal(Sessions[activeSession].Implant.Functions)

	if err != nil {
		ErrorColorBold.Println("Error: could not parse session function JSON")
		return nil
	}

	for key := range Sessions[activeSession].Implant.Functions {

		command := key

		LogData("Session " + strconv.Itoa(activeSession) + " loaded extended function: " + command)

	}

	return sessionFunctions

}
