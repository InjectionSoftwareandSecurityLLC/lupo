package cmd

import (
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

// lupoApp - Primary lupo grumble CLI construction
var lupoApp = grumble.New(&grumble.Config{
	Name:                  "lupo",
	Description:           "Lupo Modular C2",
	HistoryFile:           "/tmp/lupo.log",
	Prompt:                "lupo ☾ ",
	PromptColor:           color.New(color.FgCyan, color.Bold),
	HelpHeadlineColor:     color.New(color.FgWhite),
	HelpHeadlineUnderline: true,
	HelpSubCommands:       true,
	Flags: func(f *grumble.Flags) {
		f.String("k", "psk", "wolfpack", "Pre-Shared Key for implant authentication")
	},
})

// App - Primary grumble CLI construction variable for switching nested app contexts
var App = lupoApp

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
