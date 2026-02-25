package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/server"
	"github.com/desertbit/grumble"
)

// stagerID - a global stager ID. Stager IDs are unique and auto-increment on creation.
// This is kept completely separate from listenerID so the two do not interfere.
var stagerID int = 0

// init - Initializes the primary "stager" grumble command.
//
// "stager" subcommands include:
//
//	"start" - Starts a file stager. Accepts flags for lhost, lport, protocol (HTTP/HTTPS),
//	          key, cert, and dir (the directory to serve, created if absent).
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
		LongHelp: "Starts an HTTP or HTTPS file stager that serves files from a local directory",
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
			protocol := strings.ToUpper(c.Flags.String("protocol"))
			dir := c.Flags.String("dir")

			var tlsKey, tlsCert string
			if protocol == "HTTPS" {
				tlsKey = c.Flags.String("key")
				tlsCert = c.Flags.String("cert")
			}

			var operator string

			if server.IsWolfPackExec {

				operator = server.CurrentOperator
				core.LogData(operator + " executed: stager start -l " + lhost + " -p " + strconv.Itoa(lport) + " -x " + protocol + " -d " + dir)

				status, err := startStager(stagerID, lhost, lport, protocol, dir, tlsKey, tlsCert)

				var resp core.StartResponse
				if err != nil {
					resp = core.StartResponse{Status: "error: could not start stager: " + err.Error()}
				} else {
					resp = core.StartResponse{Status: status}
				}

				currentWolf := core.Wolves[operator]
				jsonResp, marshalErr := json.Marshal(resp)
				if marshalErr != nil {
					return errors.New("could not create JSON response")
				}
				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, string(jsonResp))

			} else {

				operator = "server"
				core.LogData(operator + " executed: stager start -l " + lhost + " -p " + strconv.Itoa(lport) + " -x " + protocol + " -d " + dir)

				status, err := startStager(stagerID, lhost, lport, protocol, dir, tlsKey, tlsCert)
				if err != nil {
					return err
				}
				core.SuccessColorBold.Println(status)

			}

			return nil
		},
	}
	stagerCmd.AddCommand(stagerStartCmd)

	stagerShowCmd := &grumble.Command{
		Name:     "show",
		Help:     "show running stagers",
		LongHelp: "Display all running stagers",
		Run: func(c *grumble.Context) error {

			var operator string
			operator = "server"

			if server.IsWolfPackExec {

				operator = server.CurrentOperator
				core.LogData(operator + " executed: stager show")

				currentWolf := core.Wolves[operator]
				stagerMap := core.ShowStagers()
				jsonResp, err := json.Marshal(stagerMap)
				if err != nil {
					return errors.New("could not create JSON response")
				}
				core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, string(jsonResp))

			} else {

				core.LogData(operator + " executed: stager show")

				table := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
				fmt.Fprintf(table, "ID\tHost\tPort\tProtocol\tDirectory\t\n")
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t\n",
					strings.Repeat("=", len("ID")),
					strings.Repeat("=", len("Host")),
					strings.Repeat("=", len("Port")),
					strings.Repeat("=", len("Protocol")),
					strings.Repeat("=", len("Directory")))

				for i := range core.Stagers {
					fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t\n",
						strconv.Itoa(core.Stagers[i].ID),
						core.Stagers[i].Lhost,
						strconv.Itoa(core.Stagers[i].Lport),
						core.Stagers[i].Protocol,
						core.Stagers[i].Dir)
				}
				table.Flush()

			}

			return nil
		},
	}
	stagerCmd.AddCommand(stagerShowCmd)

	stagerKillCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kill a stager",
		LongHelp: "Kills an HTTP/HTTPS file stager",
		Args: func(a *grumble.Args) {
			a.Int("id", "Stager ID to kill")
		},
		Run: func(c *grumble.Context) error {

			killID := c.Args.Int("id")

			var operator string
			operator = "server"

			if server.IsWolfPackExec {

				operator = server.CurrentOperator
				core.LogData(operator + " executed: stager kill " + strconv.Itoa(killID))

				currentWolf := core.Wolves[operator]
				success, fail := core.KillStager(killID)

				if success != "" {
					core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, "true")
				} else if fail != "" {
					core.AssignWolfResponse(currentWolf.Username, currentWolf.Rhost, "false")
				}

			} else {

				core.LogData(operator + " executed: stager kill " + strconv.Itoa(killID))

				success, fail := core.KillStager(killID)

				if success != "" {
					core.SuccessColorBold.Println(success)
				} else if fail != "" {
					core.ErrorColorBold.Println(fail)
				}

			}

			return nil
		},
	}
	stagerCmd.AddCommand(stagerKillCmd)
}

// startStager - Creates and starts a file stager based on the provided parameters.
//
// Creates the serving directory if it does not already exist.
// Starts an HTTP or HTTPS server whose sole handler is a static file server rooted at dir.
// The stager is stored in core.Stagers and stagerID is incremented on success.
func startStager(id int, lhost string, lport int, protocol string, dir string, tlsKey string, tlsCert string) (string, error) {

	listenString := lhost + ":" + strconv.Itoa(lport)

	// Create the serving directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		core.LogData("error: failed to create stager directory: " + dir)
		return "", errors.New("failed to create stager directory: " + err.Error())
	}

	core.LogData("Starting new " + protocol + " stager on " + listenString + " serving: " + dir)

	newServer := &http.Server{
		Addr:    listenString,
		Handler: server.StagerServerHandler(dir),
	}

	newStager := core.Stager{
		ID:           id,
		Lhost:        lhost,
		Lport:        lport,
		Protocol:     protocol,
		Dir:          dir,
		HTTPInstance: newServer,
	}

	core.Stagers[id] = newStager

	switch protocol {
	case "HTTP":
		go func(s core.Stager) {
			err := newServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				println("")
				core.LogData("error: failed to start HTTP stager")
				core.ErrorColorBold.Println(err)
				delete(core.Stagers, s.ID)
				stagerID--
			}
		}(newStager)
	case "HTTPS":
		go func(s core.Stager) {
			err := newServer.ListenAndServeTLS(tlsCert, tlsKey)
			if err != nil && err != http.ErrServerClosed {
				println("")
				core.LogData("error: failed to start HTTPS stager")
				core.ErrorColorBold.Println(err)
				delete(core.Stagers, s.ID)
				stagerID--
			}
		}(newStager)
	default:
		delete(core.Stagers, id)
		return "", errors.New("unsupported stager protocol: " + protocol + " (use HTTP or HTTPS)")
	}

	stagerID++
	return "Starting stager: " + strconv.Itoa(newStager.ID), nil
}
