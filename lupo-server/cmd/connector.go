package cmd

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/server"
	"github.com/desertbit/grumble"
)

// connectorID - a global connector ID. connector IDs are unique and auto-increment on creation. This value is kept track of throughout a connector's life cycle so it can be incremented/decremented automatically wherever appropriate.
var connectorID int = 0

// init - Initializes the primary "connector" grumble command
//
// "connector" has no arguments and serves as a base for several subcommands.
//
//  "connector" subcommands include:
//
//		"manage" - Manages global properties of connectors not related to actually establishing a connector, such as PSK rotation
//
//  	"start" - Starts a connector. Can accept flags of lhost, lport, protocol, key (used in HTTPS), and cert (used in HTTPS) to specify how a connector will establish itself via the protocol given by the user (Defaults to HTTPS).
//
//  	"show" - Prints a table of all running connectors and their configuration details.
//
//  	"kill" - Accepts an argument of "id" that is used to de-register and shutdown a connector server.

func init() {

	connectCmd := &grumble.Command{
		Name:     "connector",
		Help:     "interact with and manage connectors",
		LongHelp: "Interact with and manage an HTTP/HTTPS or TCP connector",
	}
	App.AddCommand(connectCmd)

	connectStartCmd := &grumble.Command{
		Name:     "start",
		Help:     "start a connector",
		LongHelp: "Starts an HTTP/HTTPS connector",
		Flags: func(f *grumble.Flags) {
			f.String("r", "rhost", "127.0.0.1", "connecting host IP/Domain")
			f.Int("p", "rport", 1337, "connecting host port")
			f.String("x", "protocol", "HTTPS", "protocol to connect on (HTTP, HTTPS, or TCP)")
			f.String("c", "command", "cmd=", "primary command parameter to tell the connector how to execute commands on the remote shell")
			f.String("q", "query", "", "additional query parameters to pass to the connection string, can be multiple prefilled in standard URI parameter notation (psk=example&action=thing)")
			f.String("t", "type", "GET", "the request type, GET or POST")
			f.String("d", "path", "index.php", "path and file name of your web shell")
		},
		Run: func(c *grumble.Context) error {

			rhost := c.Flags.String("rhost")
			rport := c.Flags.Int("rport")
			protocol := c.Flags.String("protocol")
			command := c.Flags.String("command")
			query := c.Flags.String("query")
			requestType := c.Flags.String("type")
			path := c.Flags.String("path")
			connectString := rhost + ":" + strconv.Itoa(rport) + "/" + path

			var operator string

			if server.IsWolfPackExec {

				operator = server.CurrentOperator

				core.LogData(operator + " executed: connector start -r " + rhost + " -p " + strconv.Itoa(rport) + " -x " + protocol)

				response, err := core.StartConnector(connectorID, rhost, rport, protocol, requestType, command, query, connectString, path)

				var resp core.StartResponse

				if err != nil {
					if response != "" {
						resp = core.StartResponse{
							Response: "",
							Status:   response,
						}
					}
				} else {
					if response != "" {
						resp = core.StartResponse{
							Response: response,
						}
					}
				}

				currentWolf := core.Wolves[operator]

				jsonResp, err := json.Marshal(resp)

				if err != nil {
					return errors.New("could not creat JSON response")
				}

				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, string(jsonResp))

			} else {
				operator = "server"

				core.LogData(operator + " executed: connector start -r " + rhost + " -p " + strconv.Itoa(rport) + " -x " + protocol)

				response, err := core.StartConnector(connectorID, rhost, rport, protocol, requestType, command, query, connectString, path)

				if err != nil {
					return errors.New(response)
				}

				core.SuccessColorBold.Println(response)

			}

			return nil
		},
	}
	connectCmd.AddCommand(connectStartCmd)

}
