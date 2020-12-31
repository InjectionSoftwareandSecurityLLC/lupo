package core

import (
	"github.com/google/uuid"
)

// Implant - implant structure for creating implants
type Implant struct {
	ID        uuid.UUID
	Arch      string
	Commands  []string
	Update    float64
	response  string
	Functions map[string]interface{}
}

// ZeroedUUID - zeroed global used to clear UUIDs wherever applicable
var ZeroedUUID, _ = uuid.Parse("00000000-0000-0000-0000-000000000000")

// RegisterImplant - function to register a new implant and increment the ImplantID
func RegisterImplant(arch string, updateInterval float64, functions map[string]interface{}) Implant {

	implantID := uuid.New()

	implant := Implant{
		ID:        implantID,
		Arch:      arch,
		Commands:  nil,
		Update:    updateInterval,
		response:  "",
		Functions: functions,
	}
	return implant
}

// UpdateImplant - function to update common implant fields on a given check in cycle
func UpdateImplant(sessionID int, updateInterval float64, functions map[string]interface{}) {
	var sessionUpdate = Sessions[sessionID]

	if updateInterval != 0 {
		sessionUpdate.Implant.Update = updateInterval
	}

	if functions != nil {
		sessionUpdate.Implant.Functions = functions
	}

	commandQueue := sessionUpdate.Implant.Commands

	if len(commandQueue) <= 1 {
		commandQueue = nil
	} else {
		commandQueue = commandQueue[1:]
	}

	sessionUpdate.Implant.Commands = commandQueue

	Sessions[sessionID] = sessionUpdate
}

// QueueImplantCommand - inserts a command to be executed by a specified implant on the next check in
func QueueImplantCommand(sessionID int, cmd string) {
	var sessionUpdate = Sessions[sessionID]

	sessionUpdate.Implant.Commands = append(sessionUpdate.Implant.Commands, cmd)

	Sessions[sessionID] = sessionUpdate
}
