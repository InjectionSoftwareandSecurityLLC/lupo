package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-client/core"

	"github.com/desertbit/grumble"
)

// init - Initializes the primary "listener" grumble command
//
// "listener" has no arguments and serves as a base for several subcommands.
//
//  "listener" subcommands include:
//
//		"manage" - Manages global properties of listeners not related to actually establishing a listener, such as PSK rotation
//
//  	"start" - Starts a listener. Can accept flags of lhost, lport, protocol, key (used in HTTPS), and cert (used in HTTPS) to specify how a listener will establish itself via the protocol given by the user (Defaults to HTTPS).
//
//  	"show" - Prints a table of all running listeners and their configuration details.
//
//  	"kill" - Accepts an argument of "id" that is used to de-register and shutdown a listener server.

func init() {

	listenCmd := &grumble.Command{
		Name:     "listener",
		Help:     "interact with and manage listeners",
		LongHelp: "Interact with and manage an HTTP/HTTPS or TCP Listener",
	}
	App.AddCommand(listenCmd)

	listenManageCmd := &grumble.Command{
		Name:     "manage",
		Help:     "manages global listener attributes",
		LongHelp: "manages global listener attributes such as the PSK",
		Flags: func(f *grumble.Flags) {
			f.String("k", "psk", "", "sets the global PSK to something new to allow for PSK rotation (this will refuse future auth to any implants using the old PSK")
			f.Bool("r", "rand", false, "generates a new random psk when coupled with or omitting an empty psk flag")
		},
		Run: func(c *grumble.Context) error {

			psk := c.Flags.String("psk")
			randPSK := c.Flags.Bool("rand")

			// Call out to server to generate new PSK

			reqString := "&command="
			commandString := "listener manage"

			if psk != "" {
				commandString += " -k " + psk
			} else if randPSK {
				commandString += " -r"
			}

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
			}

			var serverResponse *Response

			// Parse the JSON response
			// We are expecting a JSON string with the key "response" by default, the value is a second JSON object that contains the specific fields needed to map to the expected listener manage Response struct
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

			if serverResponse.Instruction == "" {
				core.WarningColorBold.Println(serverResponse.Response)
				fmt.Println("")
			} else {
				core.SuccessColorBold.Println(serverResponse.Response)
				fmt.Println(serverResponse.CurrentPSK)
				core.SuccessColorBold.Println(serverResponse.Instruction)
				fmt.Println("")
			}

			return nil
		},
	}
	listenCmd.AddCommand(listenManageCmd)

	listenStartCmd := &grumble.Command{
		Name:     "start",
		Help:     "start a listener",
		LongHelp: "Starts an HTTP/HTTPS or TCP Listener",
		Flags: func(f *grumble.Flags) {
			f.String("l", "lhost", "127.0.0.1", "listening host IP/Domain")
			f.Int("p", "lport", 1337, "listening host port")
			f.String("x", "protocol", "HTTPS", "protocol to listen on (HTTP, HTTPS, or TCP)")
			f.String("k", "key", "lupo-server.key", "(ALPHA NOTICE: FILE MUST BE ON THE SERVER) path to TLS private key")
			f.String("c", "cert", "lupo-server.crt", "(ALPHA NOTICE: FILE MUST BE ON THE SERVER) path to TLS cert")
			f.String("e", "encrypt", "", "preshared encryption key for TCP only connections.")
		},
		Run: func(c *grumble.Context) error {

			lhost := c.Flags.String("lhost")
			lport := c.Flags.Int("lport")
			protocol := c.Flags.String("protocol")
			cryptoPSK := c.Flags.String("encrypt")

			// Call out to server to start a new listener, consider how to specify new certs whether we will send them upstream or require them to be on the server already

			var tlsKey string
			var tlsCert string
			if protocol == "HTTPS" {
				tlsKey = c.Flags.String("key")
				tlsCert = c.Flags.String("cert")
			} else {
				tlsKey = ""
				tlsCert = ""
			}
			// Call out to server to generate new PSK

			reqString := "&command="
			commandString := "listener start"
			commandString += " -l " + lhost + " -p " + strconv.Itoa(lport) + " -x " + protocol + " -k " + tlsKey + " -c " + tlsCert + " -e " + cryptoPSK

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
				fmt.Println(serverResponse.CurrentPSK)
				core.SuccessColorBold.Println(serverResponse.Instruction)
				fmt.Println("")
				core.SuccessColorBold.Println(serverResponse.Help)
				fmt.Println("")
			}

			core.SuccessColorBold.Println(serverResponse.Status)

			return nil
		},
	}
	listenCmd.AddCommand(listenStartCmd)

	listenShowCommand := &grumble.Command{
		Name:     "show",
		Help:     "show running listeners",
		LongHelp: "Display all running listeners",
		Run: func(c *grumble.Context) error {

			reqString := "&command="
			commandString := "listener show"

			reqString = core.AuthURL + reqString + url.QueryEscape(commandString)

			resp, err := core.WolfPackHTTP.Get(reqString)

			if err != nil {
				fmt.Println(reqString)
				fmt.Println(err)
				return nil
			}

			defer resp.Body.Close()

			jsonData, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			// Parse the JSON response
			// We are expecting a JSON string with the key "response" by default, the value is a second JSON object that contains the specific fields needed to reference for output below. Since this data is nested and mostly "complex" strings, we use the interface maps to parse the response to a secondary map of the same nature which is then used to access the core values. Keeps things dynamic so we only have to parse twice instead of several times via a loop.
			var coreResponseInitial map[string]interface{}
			err = json.Unmarshal(jsonData, &coreResponseInitial)

			if err != nil {
				//fmt.Println(err)
				return nil
			}
			coreResponseData := coreResponseInitial["response"].(string)

			coreResponse := make(map[string]interface{})

			err = json.Unmarshal([]byte(coreResponseData), &coreResponse)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			table := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
			fmt.Fprintf(table, "ID\tHost\tPort\tProtocol\t\n")
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t\n",
				strings.Repeat("=", len("ID")),
				strings.Repeat("=", len("Host")),
				strings.Repeat("=", len("Port")),
				strings.Repeat("=", len("Protocol")))

			for i := range coreResponse {

				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t\n",
					coreResponse[i].(map[string]interface{})["ID"],
					coreResponse[i].(map[string]interface{})["Lhost"],
					coreResponse[i].(map[string]interface{})["Lport"],
					coreResponse[i].(map[string]interface{})["Protocol"])
			}

			table.Flush()
			return nil
		},
	}
	listenCmd.AddCommand(listenShowCommand)

	listenKillCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kill a listener",
		LongHelp: "Kills an HTTP/HTTPS or TCP Listener",
		Args: func(a *grumble.Args) {
			a.Int("id", "Listener ID to kill")
		},
		Run: func(c *grumble.Context) error {

			killID := c.Args.Int("id")

			// Exec to server to get listeners list

			reqString := "&command="
			commandString := "listener kill " + strconv.Itoa(killID)

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

			// Parse the JSON response
			// We are expecting a JSON string with the key "response" by default, the value is just a raw string response that can be printed to the output
			var coreResponse map[string]interface{}
			err = json.Unmarshal(jsonData, &coreResponse)

			if err != nil {
				//fmt.Println(err)
				return nil
			}

			if coreResponse["response"].(string) == "true" {
				core.SuccessColorBold.Println("Killed listener: " + strconv.Itoa(killID))

			} else {
				core.ErrorColorBold.Println("Listener: " + strconv.Itoa(killID) + " does not exist")

			}
			return nil

		},
	}
	listenCmd.AddCommand(listenKillCmd)
}
