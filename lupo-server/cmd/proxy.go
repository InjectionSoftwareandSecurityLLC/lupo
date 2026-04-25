package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/server"
	"github.com/desertbit/grumble"
)

// init - Registers the "proxy" command tree on the root Lupo App.
//
// The Messenger proxy server is a server-level resource, not tied to any
// individual session. Any Messenger-compatible client — including custom implants
// or standalone Messenger clients — can connect to it once started.
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
		Name:     "start",
		Help:     "start a new Messenger proxy server",
		LongHelp: "Starts the Messenger HTTP/WS server. Any Messenger-compatible client can connect once running.",
		Flags: func(f *grumble.Flags) {
			f.Int("p", "server-port", 8080, "port for the Messenger server to listen on")
			f.String("l", "lhost", "0.0.0.0", "interface to bind the Messenger server on")
		},
		Run: func(c *grumble.Context) error {
			serverPort := c.Flags.Int("server-port")
			lhost := c.Flags.String("lhost")
			operator := resolveOperator()
			core.LogData(operator + " executed: proxy start")

			if server.IsWolfPackExec {
				wolf := core.Wolves[operator]
				go func() {
					result, err := proxyStartLogic(lhost, serverPort)
					if err != nil {
						core.AssignWolfBroadcast(wolf.Username, wolf.Rhost,
							jsonError("proxy start failed: "+err.Error()))
						return
					}
					core.AssignWolfBroadcast(wolf.Username, wolf.Rhost, result)
				}()
				core.AssignWolfResponse(wolf.Username, wolf.Rhost,
					`{"response":"[*] Proxy starting — watch for broadcast with connection details."}`)
				return nil
			}

			fmt.Println("[*] Starting Messenger proxy server...")
			result, err := proxyStartLogic(lhost, serverPort)
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil
		},
	}
	proxyCmd.AddCommand(proxyStartCmd)

	// ------------------------------------------------------------------
	// proxy show
	// ------------------------------------------------------------------
	proxyShowCmd := &grumble.Command{
		Name:     "show",
		Help:     "show all proxy servers, SOCKS5 listeners, and TCP forwarders",
		LongHelp: "Displays tables of all running Messenger proxy servers along with their active SOCKS5 listeners and TCP forwarders",
		Run: func(c *grumble.Context) error {
			operator := resolveOperator()
			core.LogData(operator + " executed: proxy show")

			out := formatProxyShow()
			if server.IsWolfPackExec {
				wolf := core.Wolves[operator]
				b, _ := json.Marshal(map[string]string{"response": out})
				core.AssignWolfResponse(wolf.Username, wolf.Rhost, string(b))
				return nil
			}

			fmt.Print(out)
			return nil
		},
	}
	proxyCmd.AddCommand(proxyShowCmd)

	// ------------------------------------------------------------------
	// proxy kill <id>
	// ------------------------------------------------------------------
	proxyKillCmd := &grumble.Command{
		Name:     "kill",
		Help:     "kill a proxy server by ID",
		LongHelp: "Terminates the Messenger proxy server with the given ID and tears down all its active forwarders",
		Args: func(a *grumble.Args) {
			a.Int("id", "Proxy ID to kill")
		},
		Run: func(c *grumble.Context) error {
			id := c.Args.Int("id")
			operator := resolveOperator()
			core.LogData(operator + " executed: proxy kill " + strconv.Itoa(id))

			if server.IsWolfPackExec {
				wolf := core.Wolves[operator]
				if err := core.StopProxy(id); err != nil {
					core.AssignWolfResponse(wolf.Username, wolf.Rhost, jsonError(err.Error()))
					return nil
				}
				b, _ := json.Marshal(map[string]string{"response": fmt.Sprintf("[+] Proxy %d stopped.", id)})
				core.AssignWolfResponse(wolf.Username, wolf.Rhost, string(b))
				return nil
			}

			if err := core.StopProxy(id); err != nil {
				return err
			}
			fmt.Printf("[+] Proxy %d stopped.\n", id)
			return nil
		},
	}
	proxyCmd.AddCommand(proxyKillCmd)

	// ------------------------------------------------------------------
	// proxy socks  [proxy socks kill <id>]
	// ------------------------------------------------------------------
	proxySOCKSCmd := &grumble.Command{
		Name:     "socks",
		Help:     "open a SOCKS5 listener on a proxy tunnel",
		LongHelp: "Opens a SOCKS5 proxy on the lupo-server machine via a Messenger tunnel. Use 'proxy socks kill <id>' to stop.",
		Flags: func(f *grumble.Flags) {
			f.Int("p", "port", 1080, "local port to bind the SOCKS5 listener on")
			f.Int("i", "proxy", -1, "proxy ID to attach to (auto-selected when only one proxy is running)")
		},
		Run: func(c *grumble.Context) error {
			port := c.Flags.Int("port")
			operator := resolveOperator()
			state, err := resolveProxyID(c.Flags.Int("proxy"))
			if err != nil {
				if server.IsWolfPackExec {
					wolf := core.Wolves[operator]
					core.AssignWolfResponse(wolf.Username, wolf.Rhost, jsonError(err.Error()))
					return nil
				}
				return err
			}
			core.LogData(fmt.Sprintf("%s executed: proxy socks -p %d -i %d", operator, port, state.ID))

			if server.IsWolfPackExec {
				wolf := core.Wolves[operator]
				if err := core.ProxySOCKS(state.ID, port); err != nil {
					core.AssignWolfResponse(wolf.Username, wolf.Rhost, jsonError(err.Error()))
					return nil
				}
				msg := fmt.Sprintf("[+] SOCKS5 listening on 127.0.0.1:%d (proxy %d)\n"+
					"    WolfPack relay: ssh -L %d:127.0.0.1:%d <lupo-server>",
					port, state.ID, port, port)
				b, _ := json.Marshal(map[string]string{"response": msg})
				core.AssignWolfResponse(wolf.Username, wolf.Rhost, string(b))
				return nil
			}

			if err := core.ProxySOCKS(state.ID, port); err != nil {
				return err
			}
			fmt.Printf("[+] SOCKS5 proxy listening on 127.0.0.1:%d (proxy %d)\n"+
				"    Route tools through this port via proxychains or similar.\n", port, state.ID)
			return nil
		},
	}
	proxyCmd.AddCommand(proxySOCKSCmd)

	proxySOCKSKillCmd := &grumble.Command{
		Name:     "kill",
		Help:     "stop a SOCKS5 listener by forwarder ID",
		LongHelp: "Stops the SOCKS5 listener with the given ID (shown in 'proxy show')",
		Args: func(a *grumble.Args) {
			a.Int("id", "Forwarder ID to stop")
		},
		Run: func(c *grumble.Context) error {
			id := c.Args.Int("id")
			operator := resolveOperator()
			core.LogData(operator + " executed: proxy socks kill " + strconv.Itoa(id))

			if server.IsWolfPackExec {
				wolf := core.Wolves[operator]
				if err := core.ProxyKillForwarder(id); err != nil {
					core.AssignWolfResponse(wolf.Username, wolf.Rhost, jsonError(err.Error()))
					return nil
				}
				b, _ := json.Marshal(map[string]string{"response": fmt.Sprintf("[+] SOCKS5 listener %d stopped.", id)})
				core.AssignWolfResponse(wolf.Username, wolf.Rhost, string(b))
				return nil
			}

			if err := core.ProxyKillForwarder(id); err != nil {
				return err
			}
			fmt.Printf("[+] SOCKS5 listener %d stopped.\n", id)
			return nil
		},
	}
	proxySOCKSCmd.AddCommand(proxySOCKSKillCmd)

	// ------------------------------------------------------------------
	// proxy forward  [proxy forward kill <id>]
	// ------------------------------------------------------------------
	proxyForwardCmd := &grumble.Command{
		Name:     "forward",
		Help:     "open a TCP port forward on a proxy tunnel",
		LongHelp: "Opens a TCP local port forward on the lupo-server machine via a Messenger tunnel. Use 'proxy forward kill <id>' to stop.",
		Flags: func(f *grumble.Flags) {
			f.String("c", "config", "", "forward config: lhost:lport:dhost:dport")
			f.Int("i", "proxy", -1, "proxy ID to attach to (auto-selected when only one proxy is running)")
		},
		Run: func(c *grumble.Context) error {
			config := c.Flags.String("config")
			if config == "" {
				return fmt.Errorf("--config is required (format: lhost:lport:dhost:dport)")
			}
			operator := resolveOperator()
			state, err := resolveProxyID(c.Flags.Int("proxy"))
			if err != nil {
				if server.IsWolfPackExec {
					wolf := core.Wolves[operator]
					core.AssignWolfResponse(wolf.Username, wolf.Rhost, jsonError(err.Error()))
					return nil
				}
				return err
			}
			core.LogData(fmt.Sprintf("%s executed: proxy forward -c %s -i %d", operator, config, state.ID))

			if server.IsWolfPackExec {
				wolf := core.Wolves[operator]
				if err := core.ProxyForward(state.ID, config); err != nil {
					core.AssignWolfResponse(wolf.Username, wolf.Rhost, jsonError(err.Error()))
					return nil
				}
				b, _ := json.Marshal(map[string]string{"response": "[+] TCP forward active: " + config})
				core.AssignWolfResponse(wolf.Username, wolf.Rhost, string(b))
				return nil
			}

			if err := core.ProxyForward(state.ID, config); err != nil {
				return err
			}
			fmt.Println("[+] TCP forward active:", config)
			return nil
		},
	}
	proxyCmd.AddCommand(proxyForwardCmd)

	proxyForwardKillCmd := &grumble.Command{
		Name:     "kill",
		Help:     "stop a TCP forwarder by forwarder ID",
		LongHelp: "Stops the TCP port forwarder with the given ID (shown in 'proxy show')",
		Args: func(a *grumble.Args) {
			a.Int("id", "Forwarder ID to stop")
		},
		Run: func(c *grumble.Context) error {
			id := c.Args.Int("id")
			operator := resolveOperator()
			core.LogData(operator + " executed: proxy forward kill " + strconv.Itoa(id))

			if server.IsWolfPackExec {
				wolf := core.Wolves[operator]
				if err := core.ProxyKillForwarder(id); err != nil {
					core.AssignWolfResponse(wolf.Username, wolf.Rhost, jsonError(err.Error()))
					return nil
				}
				b, _ := json.Marshal(map[string]string{"response": fmt.Sprintf("[+] Forwarder %d stopped.", id)})
				core.AssignWolfResponse(wolf.Username, wolf.Rhost, string(b))
				return nil
			}

			if err := core.ProxyKillForwarder(id); err != nil {
				return err
			}
			fmt.Printf("[+] Forwarder %d stopped.\n", id)
			return nil
		},
	}
	proxyForwardCmd.AddCommand(proxyForwardKillCmd)
}

// ---------------------------------------------------------------------------
// Shared logic
// ---------------------------------------------------------------------------

func proxyStartLogic(lhost string, serverPort int) (string, error) {
	state, err := core.StartProxy(lhost, serverPort)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"[+] Proxy %d active\n"+
			"    Bind           : %s:%d\n"+
			"    Encryption key : %s\n"+
			"    Server URL     : %s",
		state.ID, lhost, serverPort, state.EncKey, state.ServerURL,
	), nil
}

// formatProxyShow builds a multi-table string of all running proxies, their
// SOCKS5 listeners, and their TCP forwarders.
func formatProxyShow() string {
	var buf bytes.Buffer
	proxies := core.GetAllProxies()

	// ---- Proxy Servers ----
	buf.WriteString("\nProxy Servers\n")
	buf.WriteString("=============\n")
	if len(proxies) == 0 {
		buf.WriteString("(none)\n")
	} else {
		tw := tabwriter.NewWriter(&buf, 0, 2, 2, ' ', 0)
		fmt.Fprintf(tw, "ID\tBind\tPort\tStatus\tClient\n")
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			strings.Repeat("=", 2), strings.Repeat("=", 4),
			strings.Repeat("=", 4), strings.Repeat("=", 6),
			strings.Repeat("=", 6))
		for _, p := range proxies {
			client := "(none)"
			if p.MessengerID != "" {
				client = p.MessengerID
			}
			fmt.Fprintf(tw, "%d\t%s\t%d\t%s\t%s\n",
				p.ID, p.Lhost, p.ServerPort, p.Status, client)
		}
		tw.Flush()
	}

	// Partition forwarders across all proxies
	type fwdRow struct {
		proxyID int
		fwd     core.ProxyForwarder
	}
	var socks, forwards []fwdRow
	for _, p := range proxies {
		for _, f := range p.GetForwarders() {
			if f.Type == "socks" {
				socks = append(socks, fwdRow{p.ID, f})
			} else {
				forwards = append(forwards, fwdRow{p.ID, f})
			}
		}
	}

	// ---- SOCKS5 Listeners ----
	buf.WriteString("\nSOCKS5 Listeners\n")
	buf.WriteString("================\n")
	if len(socks) == 0 {
		buf.WriteString("(none)\n")
	} else {
		tw := tabwriter.NewWriter(&buf, 0, 2, 2, ' ', 0)
		fmt.Fprintf(tw, "ID\tProxy\tAddress\n")
		fmt.Fprintf(tw, "%s\t%s\t%s\n",
			strings.Repeat("=", 2), strings.Repeat("=", 5), strings.Repeat("=", 7))
		for _, r := range socks {
			fmt.Fprintf(tw, "%d\t%d\t%s:%s\n",
				r.fwd.ID, r.proxyID, r.fwd.ListeningHost, r.fwd.ListeningPort)
		}
		tw.Flush()
	}

	// ---- TCP Forwarders ----
	buf.WriteString("\nTCP Forwarders\n")
	buf.WriteString("==============\n")
	if len(forwards) == 0 {
		buf.WriteString("(none)\n")
	} else {
		tw := tabwriter.NewWriter(&buf, 0, 2, 2, ' ', 0)
		fmt.Fprintf(tw, "ID\tProxy\tListening\tDestination\n")
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			strings.Repeat("=", 2), strings.Repeat("=", 5),
			strings.Repeat("=", 9), strings.Repeat("=", 11))
		for _, r := range forwards {
			fmt.Fprintf(tw, "%d\t%d\t%s:%s\t%s:%s\n",
				r.fwd.ID, r.proxyID,
				r.fwd.ListeningHost, r.fwd.ListeningPort,
				r.fwd.DestinationHost, r.fwd.DestinationPort)
		}
		tw.Flush()
	}

	return buf.String()
}

// resolveProxyID returns the ProxyState for flagID. If flagID <= 0 and exactly
// one proxy is running it auto-selects it; otherwise it returns an error.
func resolveProxyID(flagID int) (*core.ProxyState, error) {
	if flagID > 0 {
		s, ok := core.GetProxy(flagID)
		if !ok {
			return nil, fmt.Errorf("no proxy with id %d", flagID)
		}
		return s, nil
	}
	all := core.GetAllProxies()
	if len(all) == 0 {
		return nil, fmt.Errorf("no active proxies — run 'proxy start' first")
	}
	if len(all) > 1 {
		return nil, fmt.Errorf("multiple proxies running — specify one with --proxy <id>")
	}
	return all[0], nil
}

// ---------------------------------------------------------------------------
// WolfPack / local response helpers
// ---------------------------------------------------------------------------

func resolveOperator() string {
	if server.IsWolfPackExec {
		return server.CurrentOperator
	}
	return "server"
}

func jsonError(msg string) string {
	b, _ := json.Marshal(map[string]string{"response": "[!] " + msg})
	return string(b)
}
