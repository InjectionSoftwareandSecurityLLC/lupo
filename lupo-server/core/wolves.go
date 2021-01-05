package core

import (
	"fmt"
	"time"
)

// Wolf - defines a user structure known as a "wolf" composed of:
//
// WolfPSK - unique PSK randomly generated and seeded into the compilation of the wolfpack client binary on creation of a new user for authentication to the wolfpack server
//
// Username - a username to identify the user connecting to the wolfpack server
//
// Rhost - the "remote" host address. This contains a value of the external IP where a wolpack user is connecting from.
//
// Response - a response to transmit to the wolfpack user (may not be necessary if the server handler loop takes care of this once implemented)
type Wolf struct {
	WolfPSK  string
	Username string
	Rhost    string
	Response string
	Checkin  string
}

// Wolves - map of all operators (wolves). This is used to manage wolf pack server users that have been generated. The map structure makes it easy to search, add, modify, and delete a large amount of Wolves.
var Wolves = make(map[string]Wolf)

// UpdateWolf - updates the properties of an individual wolfpack user for processing elsewhere in the application. Updates the current command in the queue and the remote host connection value.
func UpdateWolf(username string, rhost string) {
	currentTime := time.Now()
	timeFormatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second())

	updateWolf := Wolves[username]

	updateWolf.Rhost = rhost

	updateWolf.Checkin = timeFormatted

	Wolves[username] = updateWolf

	LogData("Wolf structure for user: " + username + "@" + rhost + " was updated")

}
