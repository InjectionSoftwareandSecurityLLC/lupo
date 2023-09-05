package core

import (
	"strconv"

	"github.com/google/uuid"
)

// Implant - defines an implant structure composed of:
//
// id - unique identifier that is autoincremented on creation of a new implant
//
// Arch -  string for storing the Architecture of an implant's host system. This can be anything and is provided by the implant, but is expected to be something that identifies the host operating system and architecture.
//
// Commands - a slice of commands that is populated and used as a queue based on interaction from the session CLI.
//
// Update - an update interval in seconds that implants provide to tell the server how often it intends to check in. This value is used to determine if a session may have been killed.
//
// response - populated by a data payload, usually the output of execute commands on the implant. Once an implant executes a command it will be retrieved, usually through a data parameter, and populated into the implant structure it is associated with.
//
// Functions - a map of additional function names that can be registered to a given session. These contain a JSON string of {"name":"description"} that are loaded into the CLI if successfully registered via this property. Users can then execute these as unique session sub-commands. It is assumed that the implant has implemented these functions and will execute reserved actions once the registered keyword is received.
type Implant struct {
	ID        uuid.UUID
	Arch      string
	Commands  []Commands
	Update    float64
	response  string
	Functions map[string]interface{}
}

// Commands - defines the structure of Commands
//
// # Command - the actual command to be executed
//
// Operator - an operator or "wolf" that is executing the specific command
type Commands struct {
	Command  string
	Operator string
}

// ZeroedUUID - zeroed global used to clear UUIDs wherever applicable
var ZeroedUUID, _ = uuid.Parse("00000000-0000-0000-0000-000000000000")

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
		newUUIDString := "Registered old implant with UUID: " + oldUUID + "using new UUID: " + implantID.String()
		LogData(newUUIDString)
		SuccessColorBold.Println(newUUIDString)
	} else {
		LogData("Registered new implant with UUID: " + implantID.String())
	}

	return implant
}

// UpdateImplant - function to update common implant fields on a given check in cycle such as the update interval, custom functions, and the command queue.
func UpdateImplant(sessionID int, updateInterval float64, arch string, functions map[string]interface{}) {

	mutex.Lock()
	var sessionUpdate = Sessions[sessionID]
	mutex.Unlock()
	if updateInterval != 0 {
		mutex.Lock()
		sessionUpdate.Implant.Update = updateInterval
		mutex.Unlock()
	}

	if functions != nil {
		mutex.Lock()
		sessionUpdate.Implant.Functions = functions
		mutex.Unlock()
	}

	if arch != "" {
		mutex.Lock()
		sessionUpdate.Implant.Arch = arch
		mutex.Unlock()
	}

	mutex.Lock()
	commandQueue := sessionUpdate.Implant.Commands
	mutex.Unlock()

	if len(commandQueue) <= 1 {
		mutex.Lock()
		commandQueue = nil
		mutex.Unlock()
	} else {
		mutex.Lock()
		commandQueue = commandQueue[1:]
		mutex.Unlock()
	}
	mutex.Lock()
	sessionUpdate.Implant.Commands = commandQueue
	mutex.Unlock()

	mutex.Lock()
	Sessions[sessionID] = sessionUpdate
	mutex.Unlock()

	LogData("Updated implant with Session ID: " + strconv.Itoa(sessionID))
}

// QueueImplantCommand - inserts a command to the command queue to be executed by a specified implant on the next check in
func QueueImplantCommand(sessionID int, cmd string, operator string) {
	var sessionUpdate = Sessions[sessionID]

	newCommand := Commands{
		Command:  cmd,
		Operator: operator,
	}

	sessionUpdate.Implant.Commands = append(sessionUpdate.Implant.Commands, newCommand)

	Sessions[sessionID] = sessionUpdate

	LogData("Queued command on implant with Session ID " + strconv.Itoa(sessionID) + ": " + cmd)
}
