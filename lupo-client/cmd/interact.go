package cmd

import (
	"encoding/json"
	"errors"
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

// activeSession - Active session that is being interacted with by the user
//
// This data is supplied as a parameter when switching sessions with either the "interact" command or "session" sub-shell
var activeSession int

// init - Initializes the primary "interact" grumble command
//
// "interact" accepts an argument of "id" that is used to generate a new SessionApp with the SessionAppConfig
//
//  "interact" subcommands include:
//
//  	"show" - Shows all registered sessions. Accepts andargument of "id" that can be used to show a specific session based on the id.
//
//  	"kill" - Accepts an argument of "id" that is used to de-register a session.
//
//  	"clean" - De-registers all sessions marked as "DEAD" based on a pre-determined "Check-In" update interval.

func init() {

	interactCmd := &grumble.Command{
		Name:     "interact",
		Help:     "interact with a session",
		LongHelp: "Interact with an available session by specifying the Session ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Session ID to interact with")
		},
		Run: func(c *grumble.Context) error {

			activeSession = c.Args.Int("id")

			// Exec interact with server goes here to switch sessions

			reqString := "&command="
			commandString := "interact " + strconv.Itoa(activeSession)

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
				App = grumble.New(SessionAppConfig)
				App.SetPrompt("lupo session " + strconv.Itoa(activeSession) + " ☾ ")
				InitializeSessionCLI(App, activeSession)

				grumble.Main(App)

			} else {

				errorMessage := "Session " + strconv.Itoa(activeSession) + " does not exist"

				return errors.New(errorMessage)

			}
			return nil

		},
	}
	App.AddCommand(interactCmd)

	showCmd := &grumble.Command{
		Name:     "show",
		Help:     "show all sessions",
		LongHelp: "Show all available session information",
		Args: func(a *grumble.Args) {
			a.Int("id", "Filter on session id", grumble.Default(-1))
		},
		Run: func(c *grumble.Context) error {

			filterID := c.Args.Int("id")

			// Exec interact with server goes here to get a list of current sessions

			reqString := "&command="
			commandString := "interact show"

			if filterID != -1 {
				commandString += " " + strconv.Itoa(filterID)
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
				//fmt.Println(err)
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
				//fmt.Println(err)
				return nil
			}

			table := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
			fmt.Fprintf(table, "ID\tRemote Host\tArch\tProtocol\tLast Check In\tUpdate Interval\tStatus\t\n")
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
				strings.Repeat("=", len("ID")),
				strings.Repeat("=", len("Remote Host")),
				strings.Repeat("=", len("Arch")),
				strings.Repeat("=", len("Protocol")),
				strings.Repeat("=", len("Last Check In")),
				strings.Repeat("=", len("Update Interval")),
				strings.Repeat("=", len("Status")))

			// Populate this based on returned data

			if filterID != -1 {

				_, sessionExists := coreResponse[strconv.Itoa(filterID)]

				if !sessionExists {

					errorMessage := "cannot filter show on session " + strconv.Itoa(filterID) + " because the session does not exist"

					return errors.New(errorMessage)
				}

				var textStatus string

				if coreResponse[strconv.Itoa(filterID)].(map[string]interface{})["Status"] == "UNKNOWN" {
					textStatus = "UNKNOWN"
				} else if coreResponse[strconv.Itoa(filterID)].(map[string]interface{})["Status"] == "ALIVE" {
					textStatus = core.GreenColorIns("ALIVE")
				} else if coreResponse[strconv.Itoa(filterID)].(map[string]interface{})["Status"] == "DEAD" {
					textStatus = core.RedColorIns("DEAD")
				} else {
					textStatus = core.ErrorColorBoldIns("ERROR")
				}
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
					coreResponse[strconv.Itoa(filterID)].(map[string]interface{})["ID"],
					coreResponse[strconv.Itoa(filterID)].(map[string]interface{})["Rhost"],
					coreResponse[strconv.Itoa(filterID)].(map[string]interface{})["ImplantArch"],
					coreResponse[strconv.Itoa(filterID)].(map[string]interface{})["Protocol"],
					coreResponse[strconv.Itoa(filterID)].(map[string]interface{})["Checkin"],
					coreResponse[strconv.Itoa(filterID)].(map[string]interface{})["ImplantUpdate"],
					textStatus)

			} else {
				for i := range coreResponse {

					var textStatus string

					if coreResponse[i].(map[string]interface{})["Status"] == "UNKNOWN" {
						textStatus = "UNKNOWN"
					} else if coreResponse[i].(map[string]interface{})["Status"] == "ALIVE" {
						textStatus = core.GreenColorIns("ALIVE")
					} else if coreResponse[i].(map[string]interface{})["Status"] == "DEAD" {
						textStatus = core.RedColorIns("DEAD")
					} else {
						textStatus = core.ErrorColorBoldIns("ERROR")
					}

					fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
						coreResponse[i].(map[string]interface{})["ID"],
						coreResponse[i].(map[string]interface{})["Rhost"],
						coreResponse[i].(map[string]interface{})["ImplantArch"],
						coreResponse[i].(map[string]interface{})["Protocol"],
						coreResponse[i].(map[string]interface{})["Checkin"],
						coreResponse[i].(map[string]interface{})["ImplantUpdate"],
						textStatus)
				}

			}

			table.Flush()

			return nil
		},
	}
	interactCmd.AddCommand(showCmd)

	killCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kills a specified session",
		LongHelp: "Kills a session with a specified ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Session ID to kill")
		},
		Run: func(c *grumble.Context) error {

			id := c.Args.Int("id")

			// Exec command on server to return destroyed sessions

			reqString := "&command="
			commandString := "interact kill " + strconv.Itoa(id)

			reqString = core.AuthURL + reqString + url.QueryEscape(commandString)

			resp, err := core.WolfPackHTTP.Get(reqString)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			defer resp.Body.Close()

			jsonData, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				//fmt.Println(err)
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

			if err != nil {
				//fmt.Println(err)
				return nil
			}
			core.WarningColorBold.Println(coreResponseData)

			return nil
		},
	}
	interactCmd.AddCommand(killCmd)

	cleanCmd := &grumble.Command{
		Name:     "clean",
		Help:     "cleans all sessions marked as DEAD",
		LongHelp: "Kills all sessions marked as DEAD to clear up the session list.",
		Run: func(c *grumble.Context) error {

			// Exec to get cleaned sessions
			reqString := "&command="
			commandString := "interact clean"

			reqString = core.AuthURL + reqString + url.QueryEscape(commandString)

			resp, err := core.WolfPackHTTP.Get(reqString)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			defer resp.Body.Close()

			jsonData, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				//fmt.Println(err)
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

			if err != nil {
				//fmt.Println(err)
				return nil
			}
			core.WarningColorBold.Println(coreResponseData)

			return nil
		},
	}

	interactCmd.AddCommand(cleanCmd)

}
