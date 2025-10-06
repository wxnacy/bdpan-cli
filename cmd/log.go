package cmd

import (
	"fmt"

	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/config"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Tail log file",
	Run: func(cmd *cobra.Command, args []string) {
		logFile := config.Get().Logger.LogFileConfig.Filename
		if logFile == "" {
			fmt.Println("Log file not configured")
			return
		}

		t, err := tail.TailFile(logFile, tail.Config{Follow: true, ReOpen: true})
		if err != nil {
			fmt.Printf("Failed to tail log file: %s\n", err)
			return
		}

		for line := range t.Lines {
			fmt.Println(line.Text)
		}
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}