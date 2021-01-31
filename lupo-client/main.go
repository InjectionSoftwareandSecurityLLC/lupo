// main - the "main" package of the application. defines the entry point of the application.
package main

import (
	"flag"
	"fmt"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-client/cmd"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-client/core"
	"github.com/desertbit/grumble"
)

// main - executes the primary grumble application defined in the "cmd" package
func main() {

	var configFile = flag.String("c", "wolfpack.json", "config file for lupo client, expects default filename to exist if not specified")

	flag.Parse()

	err := core.InitializeWolfPackRequests(configFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	go core.Poll()

	grumble.Main(cmd.App)
}
