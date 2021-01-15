// main - the "main" package of the application. defines the entry point of the application.
package main

import (
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-client/cmd"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-client/core"
	"github.com/desertbit/grumble"
)

// main - executes the primary grumble application defined in the "cmd" package
func main() {
	core.InitializeWolfPackRequests()
	grumble.Main(cmd.App)
}
