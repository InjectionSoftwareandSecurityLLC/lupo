package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-client/core"
	"github.com/desertbit/grumble"
)

// init - Registers the "proxy" command tree on the root Lupo client App.
// Each subcommand forwards to the lupo-server via the WolfPack HTTP API.
func init() {

	proxyCmd := &grumble.Command{
		Name:     "proxy",
		Help:     "manage Messenger proxy servers",
		LongHelp: "Start and manage Messenger tunneling proxy servers for SOCKS5 and TCP port forwarding",
	}
	App.AddCommand(proxyCmd)

	// ------------------------------------------------------------------
	// proxy start
	// ------------------------------------------------------------------
	proxyStartCmd := &grumble.Command{
		Name:  "start",
		Help:  "start a new Messenger proxy server",
		Flags: func(f *grumble.Flags) {
			f.Int("p", "server-port", 8080, "port for the Messenger server to listen on")
			f.String("l", "lhost", "0.0.0.0", "interface to bind the Messenger server on")
		},
		Run: func(c *grumble.Context) error {
			cmdString := "proxy start -p " + strconv.Itoa(c.Flags.Int("server-port")) +
				" -l " + c.Flags.String("lhost")
			return wolfPackRequest(cmdString)
		},
	}
	proxyCmd.AddCommand(proxyStartCmd)

	// ------------------------------------------------------------------
	// proxy show
	// ------------------------------------------------------------------
	proxyShowCmd := &grumble.Command{
		Name: "show",
		Help: "show all proxy servers, SOCKS5 listeners, and TCP forwarders",
		Run: func(c *grumble.Context) error {
			return wolfPackRequest("proxy show")
		},
	}
	proxyCmd.AddCommand(proxyShowCmd)

	// ------------------------------------------------------------------
	// proxy kill <id>
	// ------------------------------------------------------------------
	proxyKillCmd := &grumble.Command{
		Name: "kill",
		Help: "kill a proxy server by ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Proxy ID to kill")
		},
		Run: func(c *grumble.Context) error {
			return wolfPackRequest("proxy kill " + strconv.Itoa(c.Args.Int("id")))
		},
	}
	proxyCmd.AddCommand(proxyKillCmd)

	// ------------------------------------------------------------------
	// proxy socks  [proxy socks kill <id>]
	// ------------------------------------------------------------------
	proxySOCKSCmd := &grumble.Command{
		Name: "socks",
		Help: "open a SOCKS5 listener on a proxy tunnel",
		Flags: func(f *grumble.Flags) {
			f.Int("p", "port", 1080, "local port on lupo-server to bind the SOCKS5 listener on")
			f.Int("i", "proxy", -1, "proxy ID to attach to (auto-selected when only one proxy is running)")
		},
		Run: func(c *grumble.Context) error {
			cmdString := "proxy socks -p " + strconv.Itoa(c.Flags.Int("port"))
			if id := c.Flags.Int("proxy"); id > 0 {
				cmdString += " -i " + strconv.Itoa(id)
			}
			return wolfPackRequest(cmdString)
		},
	}
	proxyCmd.AddCommand(proxySOCKSCmd)

	proxySOCKSCmd.AddCommand(&grumble.Command{
		Name: "kill",
		Help: "stop a SOCKS5 listener by forwarder ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Forwarder ID to stop")
		},
		Run: func(c *grumble.Context) error {
			return wolfPackRequest("proxy socks kill " + strconv.Itoa(c.Args.Int("id")))
		},
	})

	// ------------------------------------------------------------------
	// proxy forward  [proxy forward kill <id>]
	// ------------------------------------------------------------------
	proxyForwardCmd := &grumble.Command{
		Name: "forward",
		Help: "open a TCP port forward on a proxy tunnel",
		Flags: func(f *grumble.Flags) {
			f.String("c", "config", "", "forward config: lhost:lport:dhost:dport")
			f.Int("i", "proxy", -1, "proxy ID to attach to (auto-selected when only one proxy is running)")
		},
		Run: func(c *grumble.Context) error {
			config := c.Flags.String("config")
			if config == "" {
				fmt.Println("[!] --config is required (format: lhost:lport:dhost:dport)")
				return nil
			}
			cmdString := "proxy forward -c " + config
			if id := c.Flags.Int("proxy"); id > 0 {
				cmdString += " -i " + strconv.Itoa(id)
			}
			return wolfPackRequest(cmdString)
		},
	}
	proxyCmd.AddCommand(proxyForwardCmd)

	proxyForwardCmd.AddCommand(&grumble.Command{
		Name: "kill",
		Help: "stop a TCP forwarder by forwarder ID",
		Args: func(a *grumble.Args) {
			a.Int("id", "Forwarder ID to stop")
		},
		Run: func(c *grumble.Context) error {
			return wolfPackRequest("proxy forward kill " + strconv.Itoa(c.Args.Int("id")))
		},
	})
}

// wolfPackRequest sends a proxy command to the lupo-server and prints the response.
func wolfPackRequest(cmdString string) error {
	reqString := core.AuthURL + "&command=" + url.QueryEscape(cmdString)
	resp, err := core.WolfPackHTTP.Get(reqString)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	printWolfPackResponse(body)
	return nil
}

// printWolfPackResponse unmarshals and prints the "response" field from a
// WolfPack JSON response body.
func printWolfPackResponse(body []byte) {
	var r map[string]interface{}
	if err := json.Unmarshal(body, &r); err != nil {
		fmt.Println(string(body))
		return
	}
	if msg, ok := r["response"].(string); ok && msg != "" {
		fmt.Println(msg)
	}
}
