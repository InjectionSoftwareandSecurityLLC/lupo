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
	WolfPSK   string
	Username  string
	Rhost     string
	Response  string
	Broadcast string
	Checkin   string
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

// AssignWolfResponse - this function takes in a username and rhost to keep track of the user being assigned the response.
// The response parameter is then updated and linked to a wolf which will be returned by the WolfPack Server.
func AssignWolfResponse(username string, rhost string, response string) {
	updateWolf := Wolves[username]

	updateWolf.Response = response

	Wolves[username] = updateWolf

	LogData("Wolf response for user: " + username + "@" + rhost + " was added")
}

// AssignWolfBroadcast - this function takes in a username and rhost to keep track of the user being assigned the broadcast message.
// The response parameter is then updated and linked to a wolf which will be returned by the WolfPack Server.
func AssignWolfBroadcast(username string, rhost string, response string) {
	updateWolf := Wolves[username]

	updateWolf.Broadcast = response

	Wolves[username] = updateWolf

	LogData("Wolf broadcast for user: " + username + "@" + rhost + " was added")
}

// BroadcastWolfPackChat - this function takes in a chat message response and broadcasts it to all wolves, they will only receive it if making a request from the Chat CLI.
// The response parameter is then updated and linked to each wolf which will be returned by the WolfPack Server.
func BroadcastWolfPackChat(response string) {

	for _, wolf := range Wolves {
		wolf.Broadcast = response
		Wolves[wolf.Username] = wolf
		LogData("Wolf Chat broadcast for user: " + wolf.Username + "@" + wolf.Rhost + " was added")
	}
}
