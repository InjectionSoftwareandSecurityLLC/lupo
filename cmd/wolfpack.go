package cmd

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/core"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/server"
	"github.com/desertbit/grumble"
)

func init() {

	wolfPackCmd := &grumble.Command{
		Name:     "wolfpack",
		Help:     "manage the wolfpack server",
		LongHelp: "Manage the wolfpack team server",
	}
	App.AddCommand(wolfPackCmd)

	wolfPackStartCmd := &grumble.Command{
		Name:     "start",
		Help:     "starts the wolfpack server",
		LongHelp: "Starts the wolfpack teamserver",
		Flags: func(f *grumble.Flags) {
			f.String("l", "lhost", "127.0.0.1", "listening host IP/Domain")
			f.Int("p", "lport", 4074, "listening host port")
			f.String("k", "key", "lupo-server.key", "path to TLS private key")
			f.String("c", "cert", "lupo-server.crt", "path to TLS cert")
		},
		Run: func(c *grumble.Context) error {

			lhost := c.Flags.String("lhost")
			lport := c.Flags.Int("lport")
			listenString := lhost + ":" + strconv.Itoa(lport)

			tlsKey := c.Flags.String("key")
			tlsCert := c.Flags.String("cert")

			psk := c.Flags.String("psk")

			if core.DefaultPSK == c.Flags.String("psk") && !didDisplayPsk {
				core.SuccessColorBold.Println("Your randomly generated PSK is:")
				fmt.Println(core.DefaultPSK)
				core.SuccessColorBold.Println("Embed the PSK into any implants to connect to listeners in this instance.")
				fmt.Println("")
				didDisplayPsk = true
			}

			startWolfPackServer(listenerID, lhost, lport, listenString, psk, tlsKey, tlsCert)

			return nil
		},
	}
	wolfPackCmd.AddCommand(wolfPackStartCmd)

	wolfPackRegisterCmd := &grumble.Command{
		Name:     "register",
		Help:     "registers a wolfpack user",
		LongHelp: "Registers a wolfpack user and provides a client binary for them connect",
		Flags: func(f *grumble.Flags) {
			f.String("o", "out", "wolfpack-client", "path/name to output the wolfpack client binary")
			f.String("a", "arch", "linux", "architecture to compile the wolfpack client binary for linux (DEFAULT), darwin, or windows")
		},
		Args: func(a *grumble.Args) {
			a.String("user", "A name for the user")
			a.String("wolfpsk", "Pre-Shared Key for wolfpack client", grumble.Default(core.GeneratePSK()))
		},
		Run: func(c *grumble.Context) error {

			outFile := c.Flags.String("out")
			arch := c.Flags.String("arch")

			userName := c.Args.String("user")

			psk := c.Args.String("wolfpsk")

			wolf := core.Wolf{
				WolfPSK:  psk,
				Username: userName,
				Rhost:    "",
				Command:  "",
				Response: "",
			}

			core.Wolves[userName] = wolf

			// TODO: Generate client binary

			fmt.Println(outFile + arch)

			return nil
		},
	}
	wolfPackCmd.AddCommand(wolfPackStartCmd)
}

// startWolfPackServer - Creates a wolfpack server based on parameters generated via the "wolfpack start" subcommand.
func startWolfPackServer(id int, lhost string, lport int, listenString string, psk string, tlsKey string, tlsCert string) {

	server.PSK = psk

	newServer := &http.Server{Addr: listenString, Handler: http.HandlerFunc(server.WolfPackServerHandler)}

	core.SuccessColorBold.Println("Starting listener wolfpack server")

	go func() {
		err := newServer.ListenAndServeTLS(tlsCert, tlsKey)
		if err != nil {
			println("")
			core.ErrorColorBold.Println(err)
			return
		}
	}()

}
