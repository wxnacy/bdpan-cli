package main

import (
	"log"

	"github.com/wxnacy/bdpan-cli/cmd"
)

func main() {
	// We need to get the root command from the cmd package.
	// As rootCmd is not exported, we need a way to access it.
	// Let's add a function in cmd package to get the root command.

	rootCmd := cmd.GetRootCmd()
	rootCmd.Use = "bdpan"

	err := rootCmd.GenZshCompletionFile("scripts/completion/bdpan.zsh")
	if err != nil {
		log.Fatal(err)
	}
}
