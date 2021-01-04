// cmd - the "cmd" package is the core packaged used to reference and manage all grumble integrated commands/features of the application.
//
// The "cmd" package houses all of the core "interface/application" code which is a mix of both user interface and logical functionality.
package cmd

import (
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

// lupoApp - Primary lupo grumble CLI construction
//
// This sets up the lupo prompt and color scheme, defines a history logfile, and toggles various grumble sepcific parameters for help command options.
var lupoApp = grumble.New(&grumble.Config{
	Name:                  "lupo",
	Description:           "Lupo Modular C2",
	HistoryFile:           "/tmp/lupo.log",
	Prompt:                "lupo ☾ ",
	PromptColor:           color.New(color.FgCyan, color.Bold),
	HelpHeadlineColor:     color.New(color.FgWhite),
	HelpHeadlineUnderline: true,
	HelpSubCommands:       true,
})

// App - Primary grumble CLI construction variable for switching nested app contexts
//
// On initialization this is set to the "lupo" grumble config, but is used to switch to nested/sub-shells throughout the application where necessary.
//
// The primary use case is switching between the "lupo" core shell and the nested "session" sub-shell that handles session management.
var App = lupoApp

// init - initializes the primary Lupo cli application
func init() {
	App.SetPrintASCIILogo(func(a *grumble.App) {
		a.Println("     _                  _")
		a.Println("    | '-.            .-' |")
		a.Println("    | -. '..\\\\,.//,.' .- |")
		a.Println("    |   \\  \\\\\\||///  /   | ")
		a.Println("   /|    )M\\/%%%%/\\/(  . |\\")
		a.Println("  (/\\  MM\\/%/\\||/%\\\\/MM  /\\)")
		a.Println("  (//M   \\%\\\\\\%%//%//   M\\\\)")
		a.Println("(// M________ /\\ ________M \\\\)")
		a.Println(" (// M\\ \\(',)|  |(',)/ /M \\\\) \\\\\\\\  ")
		a.Println("  (\\\\ M\\.  /,\\\\//,\\  ./M //)")
		a.Println("    / MMmm( \\\\||// )mmMM \\  \\\\\\")
		a.Println("     // MMM\\\\\\||///MMM \\\\ \\\\")
		a.Println("      \\//''\\)/||\\(/''\\\\/ \\\\")
		a.Println("      mrf\\\\( \\oo/ )\\\\\\/\\")
		a.Println("           \\'-..-'\\/\\\\")
		a.Println("              \\\\/ \\\\")
		a.Println("                      art by Morfina")
		a.Println()
	})

}
