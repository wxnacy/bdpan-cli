/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
doc: https://pan.baidu.com/union/doc/pkuo3snyp
*/
package cmd

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
)

var (
	downloadReq = dto.NewDownloadReq()
)

func InitDownloadReq(c *cobra.Command, req *dto.DownloadReq) {

	c.Flags().StringVarP(&req.OutputDir, "output-dir", "d", "~/Downloads", "保存目录。默认为当前目录")
	c.Flags().StringVarP(&req.OutputPath, "output-path", "o", "", "保存地址。覆盖已存在文件，优先级比 --output-dir 高")

	c.Flags().BoolVar(&req.IsSync, "sync", false, "是否同步进行")
	c.Flags().BoolVarP(&req.IsRecursion, "recursion", "r", false, "是否递归下载")

}

func runDownload(cmd *cobra.Command, args []string) error {
	downloadReq.GlobalReq = *GetGlobalReq()
	if len(args) > 0 {
		downloadReq.Path = args[0]
	}
	dir, _ := homedir.Expand(downloadReq.OutputDir)
	downloadReq.OutputDir = dir
	return handler.GetFileHandler().CmdDownload(downloadReq)
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
	InitDownloadReq(downloadCmd, downloadReq)
	rootCmd.AddCommand(downloadCmd)
}
