/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
doc:
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
)

var (
	refreshReq = dto.NewRefreshReq()
)

func init() {
	var cmd = &cobra.Command{
		Use:                   "refresh",
		Short:                 "刷新数据",
		Example:               ``,
		DisableFlagsInUseLine: true,
		Long:                  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			refreshReq.GlobalReq = *GetGlobalReq()
			if len(args) > 0 {
				refreshReq.Path = args[0]
			}
			return handler.GetFileHandler().CmdRefresh(refreshReq)
		},
	}

	// cmd.Flags().IntVarP(&listReq.Page, "page", "P", 1, "页码")
	// cmd.Flags().Int32VarP(&listReq.Limit, "limit", "l", 10, "每页条数")
	cmd.Flags().BoolVarP(&refreshReq.IsSync, "sync", "s", false, "是否同步网盘")
	rootCmd.AddCommand(cmd)
}
