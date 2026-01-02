package core

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
	"strings"

	"github.com/desertbit/grumble"
)

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
	SubDomainFragments 	[]string
	
}

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
	SubDomainFragments 	string
}

var ActiveSession = -1
var Sessions sync.Map
var SessionID int = 0
var sessionsMu sync.RWMutex

func RegisterSession(sessionID int, protocol string, implant Implant, rhost string, rport int, command string, query string, requestType string, shellpath string) {
	currentTime := time.Now()
	timeFormatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second())

	session := Session{
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

	if rport != 0 {
		session.Rhost = rhost + ":" + strconv.Itoa(rport)
	}

	sessionsMu.Lock()
	Sessions.Store(sessionID, session)
	SessionID++
	sessionsMu.Unlock()

	LogData("Registered new session with ID: " + strconv.Itoa(sessionID))
}

func SessionCheckIn(sessionID int, protocol string) {
	currentTime := time.Now()
	timeFormatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second())

	sessionsMu.Lock()
	if s, ok := Sessions.Load(sessionID); ok {
		session := s.(Session)
		session.RawCheckin = currentTime
		session.Checkin = timeFormatted
		session.Protocol = protocol
		Sessions.Store(sessionID, session)
	}
	sessionsMu.Unlock()

	LogData("Session " + strconv.Itoa(sessionID) + " checked in")
}

func SessionStatusUpdate(sessionID int, status string) {
	sessionsMu.Lock()
	if s, ok := Sessions.Load(sessionID); ok {
		session := s.(Session)
		session.Status = status
		Sessions.Store(sessionID, session)
	}
	sessionsMu.Unlock()

	LogData("Updated status of session " + strconv.Itoa(sessionID) + " session is: " + status)
}

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

func ShowSessions() map[string]SessionStrings {
	stringSessions := make(map[string]SessionStrings)

	Sessions.Range(func(key, value any) bool {
		i := key.(int)
		s := value.(Session)

		tempSession := SessionStrings{
			ID:            strconv.Itoa(s.ID),
			Protocol:      s.Protocol,
			ImplantArch:   s.Implant.Arch,
			ImplantUpdate: strconv.FormatFloat(s.Implant.Update, 'f', -1, 64),
			Rhost:         s.Rhost,
			RawCheckin:    s.RawCheckin.String(),
			Checkin:       s.Checkin,
			Status:        s.Status,
			SubDomainFragments: strings.Join(s.SubDomainFragments, " "),
		}

		stringSessions[strconv.Itoa(i)] = tempSession
		return true
	})

	return stringSessions
}

func SessionExists(session int) bool {
	_, ok := Sessions.Load(session)
	return ok
}

func LoadExtendedFunctions(sessionApp *grumble.App, activeSession int) {
	if s, ok := Sessions.Load(activeSession); ok {
		for key, value := range s.(Session).Implant.Functions {
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
}

func ClientLoadExtendedFunctions(activeSession int) []byte {
	if s, ok := Sessions.Load(activeSession); ok {
		funcMap := s.(Session).Implant.Functions
		sessionFunctions, err := json.Marshal(funcMap)
		if err != nil {
			ErrorColorBold.Println("Error: could not parse session function JSON")
			return nil
		}

		for key := range funcMap {
			LogData("Session " + strconv.Itoa(activeSession) + " loaded extended function: " + key)
		}
		return sessionFunctions
	}
	return nil
}
