/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
doc: https://pan.baidu.com/union/doc/pkuo3snyp
*/
package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
)

var (
	deleteReq = dto.NewDeleteReq()
)

func init() {
	var cmd = &cobra.Command{
		Use:                   "delete",
		Short:                 "删除文件",
		Example:               ``,
		DisableFlagsInUseLine: true,
		Long:                  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			deleteReq.GlobalReq = *GetGlobalReq()
			if len(args) > 0 {
				fsid, err := strconv.Atoi(args[0])
				if err == nil {
					deleteReq.FSID = uint64(fsid)
				} else {
					deleteReq.Path = args[0]
				}
			}
			return handler.GetFileHandler().CmdDelete(deleteReq)
		},
	}

	// cmd.Flags().IntVarP(&listReq.Page, "page", "P", 1, "页码")
	// cmd.Flags().Int32VarP(&listReq.Limit, "limit", "l", 10, "每页条数")
	cmd.Flags().BoolVarP(&deleteReq.Yes, "yes", "y", false, "是否回答yes")
	rootCmd.AddCommand(cmd)
}
