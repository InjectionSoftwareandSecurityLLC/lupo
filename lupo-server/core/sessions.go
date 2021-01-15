package core

import (
	"fmt"
	"strconv"
	"time"
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
	ID         int
	Protocol   string
	Implant    Implant
	Rhost      string
	RawCheckin time.Time
	Checkin    string
	Status     string
}

// ActiveSession = global value to keep track of the current active session. Since session "0" is a valid session, this starts at "-1" to determine if no session is active.
var ActiveSession = -1

// Sessions - map of all sessions. This is used to manage sessions that are registered successfully by implants. The map structure makes it easy to search, add, modify, and delete a large amount of Sessions.
var Sessions = make(map[int]Session)

// SessionID - Global SessionID counter. Session IDs are unique and auto-increment on creation. This value is kept track of throughout a Session's life cycle so it can be incremented/decremented automatically wherever appropriate.
var SessionID int = 0

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
