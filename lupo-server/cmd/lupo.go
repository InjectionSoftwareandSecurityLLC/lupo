// Package cmd - the "cmd" package is the core packaged used to reference and manage all grumble integrated commands/features of the application.
//
// The "cmd" package houses all of the core "interface/application" code which is a mix of both user interface and logical functionality.
package cmd

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

// lupoApp - Primary lupo grumble CLI construction
//
// This sets up the lupo prompt and color scheme, defines a history logfile, and toggles various grumble specific parameters for help command options.
var lupoApp = grumble.New(&grumble.Config{
	Name:                  "lupo",
	Description:           "Lupo Modular C2",
	HistoryFile:           ".lupo.history",
	Prompt:                "lupo ☾ ",
	PromptColor:           color.New(color.FgCyan, color.Bold),
	HelpHeadlineColor:     color.New(color.FgWhite),
	HelpHeadlineUnderline: true,
	HelpSubCommands:       true,
	Flags: func(f *grumble.Flags) {
		f.String("r", "resource", "", "resource file for lupo server, all commands in this file will be executed on startup, expects default filename to exist if not specified")
	},
})

// App - Primary grumble CLI construction variable for switching nested app contexts
//
// On initialization this is set to the "lupo" grumble config, but is used to switch to nested/sub-shells throughout the application where necessary.
//
// The primary use case is switching between the "lupo" core shell and the nested "session" sub-shell that handles session management.
var App = lupoApp

// init - initializes the primary Lupo cli application
func init() {

	core.LogData("Lupo C2 started!")

	App.SetPrintASCIILogo(func(a *grumble.App) {
		a.Println("               -.`                   `..                     ")
		a.Println("             `. -..`               `..  .                   		")
		a.Println("             .  .`.``            ```.-.  .                  		")
		a.Println("            ``  .. ```.........-.``` .-  .                  		")
		a.Println("            .`  `-` ``.`   `   `..` `..  .                  		")
		a.Println("            .    `-.`      .     `-.`.`  .`                 		")
		a.Println("            .  `.-`        .       `.-.  `.                 		")
		a.Println("            -`-.`          :          `-.`.                 		")
		a.Println("            -``        `   -   `        `..                 		")
		a.Println("           -.      ````         ````      .-                		")
		a.Println("          `-    ````+++- `   ` -/++````   `-                		")
		a.Println("         `-`  ``` .``:+o``   ``o+:` . ```  `-               		")
		a.Println("         -` ``  `` ``  `       `  `` `.  `` ..              		")
		a.Println("         .`     ``     ` `   ` `     `.    `..              		")
		a.Println("         -.     ``    `  `   `  `    `.   `.-.              		")
		a.Println("         .-.`    `.`  ` ` ``` ` `  `.`    `.-`              		")
		a.Println("          ..`      ``.  `yhddy`  .``    ``.-`               		")
		a.Println("           `...      /. `mMMMm` ./`     `-.`                		")
		a.Println("             `..      .` -oyo- `.`   ``..`                  		")
		a.Println("                `..`    `.---.`     `.`                     		")
		a.Println("                   ..``          `.``                       		")
		a.Println("                     `..       ..`                          		")
		a.Println("                       `...``..`                            		")
		a.Println("                          ...                                ")
		a.Println()
		a.Println("v1.0.8")
		a.Println()
	})

}

// ExecuteResourceFile - executes a provided set of lupo commands from a specified file.
func ExecuteResourceFile(resourceFile string) {

	var rcFile *os.File
	var err error

	if resourceFile != "" {
		rcFile, err = os.Open(resourceFile)
	} else {
		return
	}

	// if we os.Open returns an error then handle it
	if err != nil {
		return
	}
	time.Sleep(1 * time.Second)
	core.LogData("Executing resource file: " + resourceFile)
	core.SuccessColorBold.Println("Executing resource file: " + resourceFile)
	time.Sleep(2 * time.Second)

	// Create a new Scanner for the file.
	scanner := bufio.NewScanner(rcFile)
	// Loop over all lines in the file and execute them.
	for scanner.Scan() {
		line := scanner.Text()
		cmd := strings.Fields(line)
		App.RunCommand(cmd)
	}
	return
}
