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

// init - Initializes the primary "stager" grumble command for the lupo client.
//
// All commands proxy through the wolfpack server via authenticated HTTP GET requests.
//
// "stager" subcommands include:
//
//	"start" - Starts a file stager on the lupo server. Accepts flags for lhost, lport,
//	          protocol (HTTP/HTTPS), key, cert, and dir.
//
//	"show"  - Prints a table of all running stagers and their configuration details.
//
//	"kill"  - Accepts an argument of "id" to shut down and de-register a stager.
func init() {

	stagerCmd := &grumble.Command{
		Name:     "stager",
		Help:     "interact with and manage stagers",
		LongHelp: "Interact with and manage HTTP/HTTPS file stagers",
	}
	App.AddCommand(stagerCmd)

	stagerStartCmd := &grumble.Command{
		Name:     "start",
		Help:     "start a stager",
		LongHelp: "Starts an HTTP or HTTPS file stager that serves files from a local directory on the lupo server",
		Flags: func(f *grumble.Flags) {
			f.String("l", "lhost", "127.0.0.1", "listening host IP/Domain")
			f.Int("p", "lport", 8080, "listening host port")
			f.String("x", "protocol", "HTTP", "protocol to serve files on (HTTP or HTTPS)")
			f.String("k", "key", "tls-certs/lupo-server.key", "path to TLS private key (HTTPS only)")
			f.String("c", "cert", "tls-certs/lupo-server.crt", "path to TLS cert (HTTPS only)")
			f.String("d", "dir", "stager", "directory to serve files from (created if it does not exist)")
		},
		Run: func(c *grumble.Context) error {

			lhost := c.Flags.String("lhost")
			lport := c.Flags.Int("lport")
			protocol := c.Flags.String("protocol")
			dir := c.Flags.String("dir")

			commandString := "stager start -l " + lhost + " -p " + strconv.Itoa(lport) + " -x " + protocol + " -d " + dir

			// Only append TLS flags when using HTTPS to avoid empty tokens in the wolfpack command split
			if strings.ToUpper(protocol) == "HTTPS" {
				tlsKey := c.Flags.String("key")
				tlsCert := c.Flags.String("cert")
				commandString += " -k " + tlsKey + " -c " + tlsCert
			}

			reqString := core.AuthURL + "&command=" + url.QueryEscape(commandString)

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
			var coreResponse map[string]interface{}

			err = json.Unmarshal(jsonData, &coreResponse)
			if err != nil {
				return nil
			}

			err = json.Unmarshal([]byte(coreResponse["response"].(string)), &serverResponse)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			core.SuccessColorBold.Println(serverResponse.Status)

			return nil
		},
	}
	stagerCmd.AddCommand(stagerStartCmd)

	stagerShowCmd := &grumble.Command{
		Name:     "show",
		Help:     "show running stagers",
		LongHelp: "Display all running stagers on the lupo server",
		Run: func(c *grumble.Context) error {

			reqString := core.AuthURL + "&command=" + url.QueryEscape("stager show")

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

			var coreResponseInitial map[string]interface{}
			err = json.Unmarshal(jsonData, &coreResponseInitial)
			if err != nil {
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
			fmt.Fprintf(table, "ID\tHost\tPort\tProtocol\tDirectory\t\n")
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t\n",
				strings.Repeat("=", len("ID")),
				strings.Repeat("=", len("Host")),
				strings.Repeat("=", len("Port")),
				strings.Repeat("=", len("Protocol")),
				strings.Repeat("=", len("Directory")))

			for i := range coreResponse {
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t\n",
					coreResponse[i].(map[string]interface{})["ID"],
					coreResponse[i].(map[string]interface{})["Lhost"],
					coreResponse[i].(map[string]interface{})["Lport"],
					coreResponse[i].(map[string]interface{})["Protocol"],
					coreResponse[i].(map[string]interface{})["Dir"])
			}
			table.Flush()

			return nil
		},
	}
	stagerCmd.AddCommand(stagerShowCmd)

	stagerKillCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kill a stager",
		LongHelp: "Kills an HTTP/HTTPS file stager on the lupo server",
		Args: func(a *grumble.Args) {
			a.Int("id", "Stager ID to kill")
		},
		Run: func(c *grumble.Context) error {

			killID := c.Args.Int("id")

			reqString := core.AuthURL + "&command=" + url.QueryEscape("stager kill "+strconv.Itoa(killID))

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

			var coreResponse map[string]interface{}
			err = json.Unmarshal(jsonData, &coreResponse)
			if err != nil {
				return nil
			}

			if coreResponse["response"].(string) == "true" {
				core.SuccessColorBold.Println("Killed stager: " + strconv.Itoa(killID))
			} else {
				core.ErrorColorBold.Println("Stager: " + strconv.Itoa(killID) + " does not exist")
			}

			return nil
		},
	}
	stagerCmd.AddCommand(stagerKillCmd)
}
