package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-client/core"
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

			// Call out to server to generate new PSK

			reqString := "&command="
			commandString := "connector start"
			commandString += " -r " + rhost + " -p " + strconv.Itoa(rport) + " -x " + protocol + " -c " + command + " -q " + query + " -t " + requestType + " -d " + path

			reqString = core.AuthURL + reqString + url.QueryEscape(commandString)

			resp, err := core.WolfPackHTTP.Get(reqString)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			defer resp.Body.Close()

			jsonData, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			type Response struct {
				Response    string
				CurrentPSK  string
				Instruction string
				Help        string
				Status      string
			}

			var serverResponse *Response

			// Parse the JSON response
			// We are expecting a JSON string with the key "response" by default, the value is a second JSON object that contains the specific fields needed to map to the expected listener start Response struct
			var coreResponse map[string]interface{}
			err = json.Unmarshal(jsonData, &coreResponse)

			if err != nil {
				//fmt.Println(err)
				return nil
			}

			err = json.Unmarshal([]byte(coreResponse["response"].(string)), &serverResponse)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			if serverResponse.Response != "" {
				core.SuccessColorBold.Println(serverResponse.Response)
			} else {
				return errors.New(serverResponse.Status)
			}

			return nil

		},
	}
	connectCmd.AddCommand(connectStartCmd)

}
