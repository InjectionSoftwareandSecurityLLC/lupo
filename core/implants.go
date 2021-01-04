package core

import (
	"github.com/google/uuid"
)

// Implant - defines an implant structure composed of:
// id - unique identifier that is autoincremented on creation of a new implant
// Arch -  string for storing the Architecture of an implant's host system. This can be anything and is provided by the implant, but is expected to be something that identifies the host operating system and architecture.
// Commands - a slice of commands that is populated and used as a queue based on interaction from the session CLI.
// Update - an update interval in seconds that implants provide to tell the server how often it intends to check in. This value is used to determine if a session may have been killed.
// response - populated by a data payload, usually the output of execute commands on the implant. Once an implant executes a command it will be retrieved, usually through a data parameter, and populated into the implant structure it is associated with.
// Functions - a map of additional function names that can be registered to a given session. These contain a JSON string of {"name":"description"} that are loaded into the CLI if successfully registered via this property. Users can then execute these as unique session sub-commands. It is assumed that the implant has implemented these functions and will execute reserved actions once the registered keyword is received.
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

// UpdateImplant - function to update common implant fields on a given check in cycle such as the update interval, custom functions, and the command queue.
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

// QueueImplantCommand - inserts a command to the command queue to be executed by a specified implant on the next check in
func QueueImplantCommand(sessionID int, cmd string) {
	var sessionUpdate = Sessions[sessionID]

	sessionUpdate.Implant.Commands = append(sessionUpdate.Implant.Commands, cmd)

	Sessions[sessionID] = sessionUpdate
}
