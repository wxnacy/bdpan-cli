/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/cmd/initial"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/terminal"
	"github.com/wxnacy/go-bdpan"
	"github.com/wxnacy/go-tools"
)

var (
	globalReq = dto.NewGlobalReq()
	Log       = bdpan.GetLogger()
	startTime time.Time
)

func GetGlobalReq() *dto.GlobalReq {
	return globalReq
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func GetFileHandler() *handler.FileHandler {
	return handler.GetFileHandler()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "bdpan",
	Short:   "百度网盘命令行工具",
	Long:    ``,
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		startTime = time.Now()
		handler.GetRequest().GlobalReq = *globalReq
		// 初始化应用
		initial.InitApp()
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		duration := time.Since(startTime)
		logger.Infof("命令执行耗时: %v\n", duration)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 检查是否登录
		if tools.ArrayContainsString([]string{
			"login",
			"log",
		}, cmd.Use) {
			access, err := config.GetAccess()
			if err != nil {
				logger.Printf("请先登录: bdpan login")
				return
			}
			if err != nil || access.IsExpired() {
				logger.Printf("登录过期，请重新登录: bdpan login")
				return
			}
		}
		req := GetGlobalReq()
		path := req.Path
		if len(args) > 0 {
			path = args[0]
		}
		if req.IsVerbose {
			logger.SetLogLevel(logrus.DebugLevel)
		}
		handleCmdErr(terminal.NewTerminal(path).Run())
	},
}

var ErrQuit = errors.New("quit bdpan")

func handleCmdErr(err error) {
	if err != nil {
		if err.Error() == "^D" ||
			err.Error() == "^C" ||
			err == ErrQuit {
			fmt.Println("GoodBye")
			os.Exit(0)
		}
		logger.Printf("Error: %v", err)
		Log.Errorf("Error: %v", err)
		os.Exit(0)
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
	defaultConfig, _ := config.GetDefaultConfigPath()
	rootCmd.PersistentFlags().BoolVarP(&globalReq.IsVerbose, "verbose", "V", false, "打印赘余信息")
	rootCmd.PersistentFlags().StringVarP(&globalReq.Config, "config", "c", defaultConfig, "指定配置文件地址")

	// root 参数
	// rootCmd.PersistentFlags().StringVarP(&bdpanCommand.Path, "path", "p", "/", "直接查看文件")
	rootCmd.PersistentFlags().StringVarP(&globalReq.Path, "path", "p", "/", "网盘文件地址")
	// rootCmd.PersistentFlags().IntVarP(&rootCommand.Limit, "limit", "l", 10, "查询数目")
	// 运行前全局命令
	cobra.OnInitialize(func() {
		// 打印 debug 日志
		if globalReq.IsVerbose {
			bdpan.SetLogLevel(logrus.DebugLevel)
		}
	})
}
