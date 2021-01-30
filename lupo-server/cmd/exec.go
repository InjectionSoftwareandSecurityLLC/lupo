package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/server"
	"github.com/desertbit/grumble"
	"github.com/mattn/go-shellwords"
)

// init - Initializes the primary "exec" grumble command
//
// "exec" accepts an argument of "cmd" that will execute any local commands to around shell interaction with the host system without leaving the Lupo cli
//

func init() {

	execCmd := &grumble.Command{
		Name:     "exec",
		Help:     "execute a local shell command",
		LongHelp: "Executes a local shell command to allow interaction with the local system while remaining in the Lupo CLI",
		Args: func(a *grumble.Args) {
			a.StringList("cmd", "Command to execute on the local system")
		},
		Run: func(c *grumble.Context) error {

			cmd := strings.Join(c.Args.StringList("cmd"), " ")

			var operator string

			operator = "server"

			if server.IsWolfPackExec {
				operator = server.CurrentOperator

				core.LogData(operator + " executed: exec " + cmd)

			} else {
				core.LogData(operator + " executed: exec " + cmd)

				parsedCmd, err := shellwords.Parse(cmd)

				// Get the root command
				cmd := parsedCmd[0]

				// Cut off the root command and extract any args if they exist
				argS := parsedCmd[1:]

				if err != nil {
					return nil
				}

				// Check if it is a command with our without args and execute appropriately
				if cmd != "" && len(argS) > 0 {
					// Maintain directory context if cd is issued
					if cmd == "cd" {
						os.Chdir(strings.Join(argS, " "))
					} else {
						data, err := exec.Command(cmd, argS...).Output()

						if err != nil {
							return nil
						}

						fmt.Println(string(data))

					}

				} else if cmd != "" {

					data, err := exec.Command(cmd).Output()

					if err != nil {
						return nil
					}

					fmt.Println(string(data))

				}

			}

			return nil
		},
	}
	App.AddCommand(execCmd)

}
