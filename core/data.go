package core

import "github.com/google/uuid"

// TCPData - structure to contain parsable TCP initialization data received from TCP implants.
type TCPData struct {
	PSK                 string
	SessionID           int
	UUID                uuid.UUID
	ImplantArch         string
	Update              float64
	Data                string
	AdditionalFunctions string
	Register            bool
}
