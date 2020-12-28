package core

// Implant - implant structure for creating implants
type Implant struct {
	id        int
	Arch      string
	Command   string
	response  string
	functions string
}

// ImplantID - global implant ID to keep track of implants
var ImplantID = 0

// RegisterImplant - function to register a new implant and increment the ImplantID
func RegisterImplant(arch string, functions string) Implant {
	implant := Implant{
		id:        ImplantID,
		Arch:      arch,
		Command:   "",
		response:  "",
		functions: functions,
	}
	ImplantID++
	return implant
}
