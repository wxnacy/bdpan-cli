/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/cli"
)

var (
	globalArg = &GlobalArg{}
	Log       = bdpan.GetLogger()
)

type GlobalArg struct {
	IsVerbose bool
	AppId     string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "bdpan",
	Short:   "百度网盘命令行工具",
	Long:    ``,
	Version: "0.3.0",
	Run: func(cmd *cobra.Command, args []string) {
		handleCmdErr(bdpanCommand.Exec(args))
	},
}

func handleCmdErr(err error) {
	if err != nil {
		if err.Error() == "^D" ||
			err.Error() == "^C" ||
			err == cli.ErrQuit {
			fmt.Println("GoodBye")
			return
		}
		Log.Errorf("Error: %v", err)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 全局参数
	rootCmd.PersistentFlags().StringVar(&globalArg.AppId, "app-id", "", "指定添加 App Id")
	rootCmd.PersistentFlags().BoolVarP(&globalArg.IsVerbose, "verbose", "v", false, "打印赘余信息")

	// root 参数
	rootCmd.PersistentFlags().StringVarP(&bdpanCommand.Path, "path", "p", "/", "直接查看文件")
	// rootCmd.PersistentFlags().IntVarP(&rootCommand.Limit, "limit", "l", 10, "查询数目")
	// 运行前全局命令
	cobra.OnInitialize(func() {
		// 打印 debug 日志
		if globalArg.IsVerbose {
			bdpan.SetLogLevel(logrus.DebugLevel)
		}
	})
}
