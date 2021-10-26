package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-client/core"
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

// ChatAppConfig - Primary chat nested grumble CLI config construction
// This sets up the lupo "chat" nested/sub-prompt and color scheme, defines a history logfile, and toggles various grumble sepcific parameters for help command options.
var ChatAppConfig = &grumble.Config{
	Name:                  "chat",
	Description:           "Interactive Chat CLI",
	HistoryFile:           ".lupo.history",
	Prompt:                "lupo chat ☾ ", // placeholder, will get this value from the server
	PromptColor:           color.New(color.FgGreen, color.Bold),
	HelpHeadlineColor:     color.New(color.FgWhite),
	HelpHeadlineUnderline: true,
	HelpSubCommands:       true,
	Flags: func(f *grumble.Flags) {
		f.String("c", "config", "wolfpack.json", "config file for lupo client, expects default filename to exist if not specified")
		f.String("r", "resource", "", "resource file for lupo server, all commands in this file will be executed on startup, expects default filename to exist if not specified")
	},
}

// InitializeChatCLI - Initialize the nested chat CLI arguments
//
// "chat" has no arguments and is not a grumble command in and of itself. It is a separate nested grumble application and contains all new base commands.
//
// "chat" base commands include:
//
// 	"back" - resets the current active chat to "false" and closes the nested chat sub-shell.
//
// 	"message" - takes a message that will be sent to the chat API and broadcasts to all users in the chat.
func InitializeChatCLI(chatApp *grumble.App) {

	// Initialize a new polling thread specific to this shell CLI so we still receive broadcasts messages
	go core.ChatRelay()
	go GetChatLog()

	core.ActiveChat = true

	// Send log to server
	//core.LogData(operator + " started interaction with chat: " + strconv.Itoa(activeChat))

	backCmd := &grumble.Command{
		Name:     "back",
		Help:     "go back to core lupo cli (or use the exit command)",
		LongHelp: "Exit interactive chat cli and return to lupo cli (The 'exit' command is an optional built-in to go back as well) ",
		Run: func(c *grumble.Context) error {

			// Exec to server to get listeners list

			reqString := "&isChatShell=true&command="
			commandString := "back"

			reqString = core.AuthURL + reqString + url.QueryEscape(commandString)

			_, err := core.WolfPackHTTP.Get(reqString)

			if err != nil {
				fmt.Println(err)
				return nil
			}
			core.ActiveChat = false
			chatApp.Close()

			return nil
		},
	}
	chatApp.AddCommand(backCmd)

	chatMessageCmd := &grumble.Command{
		Name:     "message",
		Help:     "interact with the chat and send messages",
		LongHelp: "Interact with a different available chat by specifying the Chat ID",
		Args: func(a *grumble.Args) {
			a.StringList("message", "message to send to the chat")
		},
		Run: func(c *grumble.Context) error {
			message := strings.Join(c.Args.StringList("message"), " ")

			// Exec on server to get chats

			reqString := "&isChatShell=true&message="

			reqString = core.AuthURL + reqString + url.QueryEscape(message)

			resp, err := core.WolfPackHTTP.Get(reqString)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			defer resp.Body.Close()

			jsonData, err := ioutil.ReadAll(resp.Body)

			// Clean up new lines from potentially JSON breaking output
			re := regexp.MustCompile(`\n`)
			jsonDataString := re.ReplaceAllString(string(jsonData), "\\n")

			jsonCleanData := []byte(jsonDataString)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			// Parse the JSON response
			// We are expecting a JSON string with the key "response" by default, the value is just a raw string response that can be printed to the output
			var coreResponse map[string]interface{}
			err = json.Unmarshal(jsonCleanData, &coreResponse)

			if err != nil {
				//fmt.Println(err)
				return nil
			}

			if coreResponse["chatData"].(string) != "" {
				fmt.Println(coreResponse["chatData"].(string))
			}
			return nil
		},
	}

	chatApp.AddCommand(chatMessageCmd)

}

func GetChatLog() {
	// Exec on server to get chats

	reqString := "&isChatShell=true&getChatLog=true"

	reqString = core.AuthURL + reqString

	resp, err := core.WolfPackHTTP.Get(reqString)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	jsonData, err := ioutil.ReadAll(resp.Body)

	// Clean up new lines from potentially JSON breaking output
	re := regexp.MustCompile(`\n`)
	jsonDataString := re.ReplaceAllString(string(jsonData), "\\n")

	jsonCleanData := []byte(jsonDataString)

	if err != nil {
		fmt.Println(err)
	}

	// Parse the JSON response
	// We are expecting a JSON string with the key "response" by default, the value is just a raw string response that can be printed to the output
	var coreResponse map[string]interface{}
	err = json.Unmarshal(jsonCleanData, &coreResponse)

	fmt.Println(coreResponse["chatData"].(string))

	if err != nil {
		fmt.Println(err)
	}

	if coreResponse["chatData"].(string) != "" {
		fmt.Println(coreResponse["chatData"].(string))
	}
}
