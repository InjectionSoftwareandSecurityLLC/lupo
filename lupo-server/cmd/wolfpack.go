package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/server"
	"github.com/desertbit/grumble"
)

// WolfPackServer - defines a wolfpack server structure composed of:
//
// lhost - the "listening" host address. This tells a listener what interface to listen on based on the address it is tied to.
//
// lport - the "listening" port. This tells a listener what port the lhost of the listener should open to receive connections on.
//
// status - the current status of the wolfpack server to determine if it is online or not.
//
// httpInstance - a pointer to an instance of the http.Server struct. This is used to reference the core HTTP Server itself when conducting operations such as starting/stopping a wolfpack server.
type WolfPackServer struct {
	lhost        string
	lport        int
	status       bool
	httpInstance *http.Server
}

var wolfPackServer WolfPackServer

var tlsCert string

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
			f.Int("p", "lport", 3074, "listening host port")
			f.String("k", "key", "lupo-server.key", "path to TLS private key")
			f.String("c", "cert", "lupo-server.crt", "path to TLS cert")
		},
		Run: func(c *grumble.Context) error {

			lhost := c.Flags.String("lhost")
			lport := c.Flags.Int("lport")
			listenString := lhost + ":" + strconv.Itoa(lport)

			tlsKey := c.Flags.String("key")
			tlsCert = c.Flags.String("cert")

			var operator string

			operator = "server"

			core.LogData(operator + " executed: wolfpack start -l " + lhost + " -p " + strconv.Itoa(lport) + " -k " + tlsKey + " -c " + tlsCert)

			app := App

			startWolfPackServer(listenerID, lhost, lport, listenString, psk, tlsKey, tlsCert, app)

			return nil
		},
	}
	wolfPackCmd.AddCommand(wolfPackStartCmd)

	wolfPackStopCmd := &grumble.Command{
		Name:     "stop",
		Help:     "stops the wolfpack server",
		LongHelp: "Stops the wolfpack teamserver",
		Run: func(c *grumble.Context) error {

			var operator string

			operator = "server"

			core.LogData(operator + " executed: wolfpack stop")

			wolfPackServer.httpInstance.Close()

			return nil
		},
	}
	wolfPackCmd.AddCommand(wolfPackStopCmd)

	wolfPackDeRegister := &grumble.Command{
		Name:     "deregister",
		Help:     "deregisters a wolfpack user",
		LongHelp: "Deregisters a wolfpack user removing their access to the wolfpack server",
		Args: func(a *grumble.Args) {
			a.String("user", "Username to deregister")
		},
		Run: func(c *grumble.Context) error {

			userName := c.Args.String("user")

			var operator string

			operator = "server"

			core.LogData(operator + " executed: wolfpack deregister " + userName)

			delete(core.Wolves, userName)

			core.LogData("User " + userName + " was deregistered")

			core.SuccessColorBold.Println("User Deregistered!")

			return nil
		},
	}
	wolfPackCmd.AddCommand(wolfPackDeRegister)

	wolfPackRegisterCmd := &grumble.Command{
		Name:     "register",
		Help:     "registers a wolfpack user",
		LongHelp: "Registers a wolfpack user and provides a client config for them connect via the Lupo Client",
		Flags: func(f *grumble.Flags) {
			f.String("o", "out", "wolfpack.json", "path/name to output the wolfpack client config")
		},
		Args: func(a *grumble.Args) {
			a.String("user", "A name for the user")
			a.String("wolfpsk", "Pre-Shared Key for wolfpack client", grumble.Default(core.GeneratePSK()))
		},
		Run: func(c *grumble.Context) error {

			outFile := c.Flags.String("out")

			userName := c.Args.String("user")

			psk := c.Args.String("wolfpsk")

			wolf := core.Wolf{
				WolfPSK:  psk,
				Username: userName,
				Rhost:    "",
				Response: "",
			}

			var operator string

			operator = "server"

			core.LogData(operator + " executed: wolfpack register -o " + outFile + " " + userName + " <redacted>")

			//Generate client config

			err := generateLupoClientConfig(wolf, outFile)

			if err != nil {
				core.ErrorColorBold.Println("User Was Not Registered!")
			} else {
				core.Wolves[userName] = wolf

				core.LogData("Registered User " + userName)

				core.SuccessColorBold.Println("User Registered!")
			}

			return nil
		},
	}
	wolfPackCmd.AddCommand(wolfPackRegisterCmd)

	wolfPackShow := &grumble.Command{
		Name:     "show",
		Help:     "show wolfpack server status and registered user information",
		LongHelp: "Displays the status of the wolfpack server along with information about registered wolfpack users",
		Run: func(c *grumble.Context) error {

			var operator string

			operator = "server"

			core.LogData(operator + " executed: wolfpack show")

			var status string
			var listenString string
			if wolfPackServer.status {
				status = core.GreenColorIns("ONLINE")
				listenString = wolfPackServer.lhost + ":" + strconv.Itoa(wolfPackServer.lport)
			} else {
				status = core.RedColorIns("OFFLINE")
				listenString = "None"
			}

			table := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
			fmt.Fprintf(table, "Listening On\tStatus\t\n")
			fmt.Fprintf(table, "%s\t%s\t\n",
				strings.Repeat("=", len("Listening On")),
				strings.Repeat("=", len("Status")))
			fmt.Fprintf(table, "%s\t%s\t\n",
				listenString,
				status)

			fmt.Fprintf(table, "\nUser\tRemote Host\tCheck In\t\n")
			fmt.Fprintf(table, "%s\t%s\t%s\t\n",
				strings.Repeat("=", len("User")),
				strings.Repeat("=", len("Remote Host")),
				strings.Repeat("=", len("Check In")))

			for user := range core.Wolves {
				fmt.Fprintf(table, "%s\t%s\t%s\t\n",
					core.Wolves[user].Username,
					core.Wolves[user].Rhost,
					core.Wolves[user].Checkin)
			}

			table.Flush()
			return nil
		},
	}
	wolfPackCmd.AddCommand(wolfPackShow)
}

// startWolfPackServer - Creates a wolfpack server based on parameters generated via the "wolfpack start" subcommand.
func startWolfPackServer(id int, lhost string, lport int, listenString string, psk string, tlsKey string, tlsCert string, app *grumble.App) {

	core.ChatLog("Wolfpack Chat Server Started!")

	if wolfPackServer.status == true {
		var wolfPackInstanceError = "error: An instance of Wolfpack Server is already running, please run 'wolfpack stop' to kill the instance"
		core.LogData(wolfPackInstanceError)
		core.ErrorColorBold.Println(wolfPackInstanceError)
		return
	}
	server.WolfPackApp = app

	core.LogData("Starting new WolfPack server on " + listenString)

	newServer := &http.Server{Addr: listenString, Handler: http.HandlerFunc(server.WolfPackServerHandler)}

	wolfPackServer.lhost = lhost
	wolfPackServer.lport = lport
	wolfPackServer.httpInstance = newServer

	core.SuccessColorBold.Println("Starting WolfPack Server on: " + listenString)
	wolfPackServer.status = true
	go func() {
		err := newServer.ListenAndServeTLS(tlsCert, tlsKey)
		if err != nil {
			println("")
			core.LogData("error: failed to start WolfPack server")
			core.ErrorColorBold.Println(err)
			wolfPackServer.status = false
			return
		}
	}()

}

func generateLupoClientConfig(wolf core.Wolf, outFile string) error {

	certFile, err := os.Open(tlsCert)

	// if we os.Open returns an error then handle it
	if err != nil {
		core.ErrorColorUnderline.Println("could not read tlsCert file, did you start the Wolfpack server?")
		return err
	}

	certData, _ := ioutil.ReadAll(certFile)

	certFile.Close()

	certDataString := string(certData)

	re := regexp.MustCompile(`\n`)
	certDataString = re.ReplaceAllString(certDataString, "\\n")

	lupoClient := []byte(`
	{
		"protocol" : "https://",
		"rhost" :"` + wolfPackServer.lhost + `",
		"rport" :` + strconv.Itoa(wolfPackServer.lport) + `,
		"userName" :"` + wolf.Username + `",
		"psk" :"` + wolf.WolfPSK + `",
		"cert" :"` + certDataString + `"
	}
	`)

	lupoClientDir := "wolfpack_configs/"
	lupoClientSrcFile := outFile

	err = os.MkdirAll(lupoClientDir, 0755)
	if err != nil {
		core.ErrorColorBold.Println(err)
		return err

	}

	configFile, err := os.Create(lupoClientDir + lupoClientSrcFile)

	if err != nil {
		core.ErrorColorBold.Println(err)
		return err
	}

	configFile.WriteString(string(lupoClient))

	configFile.Close()

	core.LogData("Generated lupo client config for " + wolf.Username)
	core.SuccessColorBold.Println("Generated lupo client config for " + wolf.Username + " (" + outFile + ")")

	return nil

}
