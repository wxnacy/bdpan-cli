/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
)

// uploadCommand *UploadCommand
var uploadReq = dto.NewUploadReq()

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "上传文件",
	Long: `
上传文件
bdpan upload --local 本地文件夹 --path 网盘目录
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// err := uploadCommand.Run()
		uploadReq.GlobalReq = *GetGlobalReq()
		handleCmdErr(handler.GetFileHandler().CmdUpload(uploadReq))
	},
}

func init() {
	uploadCmd.Flags().StringVarP(&uploadReq.Local, "local", "l", "", "本地文件")
	uploadCmd.Flags().BoolVarP(&uploadReq.IsRewrite, "rewrite", "r", false, "是否覆盖同名文件，上传文件夹时不可指定，一直是 true")
	rootCmd.AddCommand(uploadCmd)
}
