package core

type Wolf struct {
	WolfPSK  string
	Username string
	Rhost    string
	Command  string
	Response string
}

// Wolves - map of all operators (wolves). This is used to manage wolf pack server users that have been generated. The map structure makes it easy to search, add, modify, and delete a large amount of Wolves.
var Wolves = make(map[string]Wolf)

func UpdateWolf(username string, command string, rhost string) {
	updateWolf := Wolves[username]

	updateWolf.Command = command

	updateWolf.Rhost = rhost

	Wolves[username] = updateWolf
}
