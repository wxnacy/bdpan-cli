/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
doc: https://pan.baidu.com/union/doc/pkuo3snyp
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
)

var (
	listReq = dto.NewListReq()
)

func init() {
	var cmd = &cobra.Command{
		Use:                   "list",
		Short:                 "展示文件",
		Example:               ``,
		DisableFlagsInUseLine: true,
		Long:                  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			listReq.GlobalReq = *GetGlobalReq()
			if len(args) > 0 {
				listReq.Path = args[0]
			}
			return handler.GetFileHandler().CmdList(listReq)
		},
	}

	cmd.Flags().IntVarP(&listReq.Page, "page", "P", 1, "页码")
	cmd.Flags().Int32VarP(&listReq.Limit, "limit", "l", 10, "每页条数")
	cmd.Flags().BoolVarP(&listReq.WithoutTui, "without-tui", "T", false, "是否用不用 TUI 展示")
	rootCmd.AddCommand(cmd)
}
