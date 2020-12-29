package core

import (
	"github.com/google/uuid"
)

// Implant - implant structure for creating implants
type Implant struct {
	ID        uuid.UUID
	Arch      string
	Command   string
	response  string
	functions string
}

// ZeroedUUID - zeroed global used to clear UUIDs wherever applicable
var ZeroedUUID, _ = uuid.Parse("00000000-0000-0000-0000-000000000000")

// RegisterImplant - function to register a new implant and increment the ImplantID
func RegisterImplant(arch string, functions string) Implant {

	implantID := uuid.New()

	implant := Implant{
		ID:        implantID,
		Arch:      arch,
		Command:   "",
		response:  "",
		functions: functions,
	}
	return implant
}
