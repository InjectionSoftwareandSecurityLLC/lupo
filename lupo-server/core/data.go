package core

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TCPData - structure to contain parsable TCP initialization data received from TCP implants.
//
// Since TCP is a more raw network protocol it does not support a mechanism to collect parameterized data by default.
//
// As a result TCPServerHandler() function expects implant clients to transmit JSON data containing the relevant fields for conducting various operations.
//
// These fields are mapped via JSON to the TCPData struct to establish and interact with TCP based sessions.
//
// TCPData structures are composed of:
//
// PSK - the client Pre-Shared Key that the the implant will send to be compared for authentication to the server PSK
//
// SessionID - a unique Session ID that the implant sends to identify what session it is. This value is supplied to implants by the server after a successful registration.
//
// UUID - a unique UUID formatted identifier that the implant sends to identify what session it is. This value is supplied to implants by the server after a successful registration. The UUID is not the primary identifier but is a secondary validation to prevent id bruteforcing or id mis-matches during the registration/de-registration processes.
//
// ImplantArch - a string for storing the Architecture of an implant's host system. This can be anything and is provided by the implant, but is expected to be something that identifies the host operating system and architecture.
//
// Update - an update interval in seconds that implants provide to tell the server how often it intends to check in. This value is used to determine if a session may have been killed.
//
// Data - a data payload, usually the output of execute commands on the implant. Once an implant executes a command, it can send the output to the data parameter where it will be printed to the user in the Lupo CLI.
//
// AdditionalFunctions - additional function names that can be registered to a given session. These contain a JSON string of {"name":"description"} that is loaded into the CLI if successfully registered. Users can then execute these as unique session sub-commands. It is assumed that the implant has implemented these functions and will execute reserved actions once the registered keyword is received.
//
// Register - a boolean value that lets a listener know if an implant is attempting to register itself or not. If not provided registration is assumed to be false. If registration is attempted the listener will check for valid authentication via the PSK and attempt to register a new session.
//
// FileName - a string value provided by an implant that is the filename for a file being sent to download.
//
// File - a string value that is expected to be a base64 encoded string that is a file.
//
// CryptoPSK - a string value that is a preshared key for encrypting/decrypting raw TCP payloads.

type TCPData struct {
	PSK                 string
	SessionID           int
	UUID                uuid.UUID
	ImplantArch         string
	Update              float64
	Data                string
	AdditionalFunctions string
	Username            string
	Register            bool
	FileName            string
	File                string
}

// GeneratePSK - Generates a random 32 character string, encodes it with SHA256 as a PSK that is set by default on startup unless the user specifies a static PSK
func GeneratePSK() string {

	LogData("Generated new random Lupo C2 PSK")

	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"~`!@#$%^&*()\\/,.<>?+=")
	length := 12
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	hasher := sha256.New()
	hash := hex.EncodeToString(hasher.Sum([]byte(str)))
	return hash
}
