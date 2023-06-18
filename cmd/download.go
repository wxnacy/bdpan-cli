/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
doc: https://pan.baidu.com/union/doc/pkuo3snyp
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/cli"
)

var (
	downloadCommand *cli.DownloadCommand
)

func NewDownloadCommand(c *cobra.Command) *cli.DownloadCommand {
	cmd := &cli.DownloadCommand{}

	c.Flags().StringVarP(&cmd.OutputDir, "output-dir", "d", "", "保存目录。默认为当前目录")
	c.Flags().StringVarP(&cmd.OutputPath, "output-path", "o", "", "保存地址。覆盖已存在文件，优先级比 --output-dir 高")

	c.Flags().BoolVar(&cmd.IsSync, "sync", false, "是否同步进行")
	c.Flags().BoolVarP(&cmd.IsRecursion, "recursion", "r", false, "是否递归下载")

	return cmd
}

func runDownload(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		downloadCommand.From = args[0]
	}
	downloadCommand.IsVerbose = globalArg.IsVerbose
	return downloadCommand.Run()
}

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "下载文件",
	Example: `  bdpan download /apps/video.mp4				下载文件
  bdpan download /apps/video.mp4 -d ~/Downloads			指定下载目录
  bdpan download /apps/video.mp4 -o ~/Downloads/1.mp4		指定下载地址
	`,
	DisableFlagsInUseLine: true,
	Long:                  ``,
	RunE:                  runDownload,
}

func init() {
	downloadCommand = NewDownloadCommand(downloadCmd)
	rootCmd.AddCommand(downloadCmd)
}
