package cmd

import (
	"github.com/desertbit/grumble"
)

// init - Initializes the primary "chat" grumble command
//
// "chat" enables the ActiveChat boolean and drops the user into the chat sub-shell CLI.
//
//  "chat" subcommands include:
//
// 	"back" - resets the current active chat to "false" and closes the nested chat sub-shell.
//
// 	"message" - takes a message that will be sent to the chat API and broadcasts to all users in the chat.
func init() {

	chatCmd := &grumble.Command{
		Name:     "chat",
		Help:     "interact with the chat and send messages",
		LongHelp: "Interact with a different available chat by specifying the Chat ID",
		Run: func(c *grumble.Context) error {

			App = grumble.New(ChatAppConfig)
			App.SetPrompt("lupo chat ☾ ")
			InitializeChatCLI(App)

			grumble.Main(App)

			return nil

		},
	}
	App.AddCommand(chatCmd)
}
