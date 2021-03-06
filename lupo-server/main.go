// main - the "main" package of the application. defines the entry point of the application.
package main

import (
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/cmd"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/desertbit/grumble"
)

// main - executes the primary grumble application defined in the "cmd" package
func main() {

	grumble.Main(cmd.App)
	core.LogData("Lupo C2 stopped!")
}
