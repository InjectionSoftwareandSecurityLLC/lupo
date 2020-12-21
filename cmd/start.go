/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts an interactive lupo C2 instance.",
	Long: `Starts an interactive lupo C2 instance. This is the primary command to begin C2 operations. 
	
	Once started you will be dropped into an interactive lupo prompt that enables further interaction with the service.`,
	Run: func(cmd *cobra.Command, args []string) {
		for {
			var cmdInput string

			fmt.Print("lupo > ")
			fmt.Scanln(&cmdInput)

			if cmdInput == "listen" {
				cmd = listenCmd
				cmd.Run(cmd, nil)
			}
			cmd = nil
			cmdInput = ""
			fmt.Println("start called")
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
