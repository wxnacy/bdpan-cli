package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"gopkg.in/yaml.v3"
)

// configCmd prints current config as YAML
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "以 YAML 格式打印当前配置",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		b, err := yaml.Marshal(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(string(b))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
