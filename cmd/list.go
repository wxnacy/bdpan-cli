/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
doc: https://pan.baidu.com/union/doc/pkuo3snyp
*/
package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/terminal"
)

func init() {
	var req = dto.NewListReq()
	run := func(req *dto.ListReq) error {
		var width, height, maxFilenameW int
		limit := req.Limit
		files, err := handler.
			GetFileHandler().
			Limit(limit).
			GetFilesAndSave(req.Path, req.Page)
		if err != nil {
			return err
		}

		// 获取文件名的最大长度
		for _, v := range files {
			w := lipgloss.Width(v.GetFilename())
			if w > maxFilenameW {
				maxFilenameW = w
			}
		}

		// limit + 固定富裕长度
		height = int(limit) + 5
		// 最大文件名长度+其他列宽度+固定富裕长度
		width = maxFilenameW + 40 + 15
		view := terminal.
			NewFileList(files, width, height).
			View()
		fmt.Println(view)
		return nil
	}
	var cmd = &cobra.Command{
		Use:                   "list",
		Short:                 "展示文件",
		Example:               ``,
		DisableFlagsInUseLine: true,
		Long:                  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			req.GlobalReq = *GetGlobalReq()
			if len(args) > 0 {
				req.Path = args[0]
			}
			return run(req)
		},
	}

	cmd.Flags().IntVarP(&req.Page, "page", "P", 1, "页码")
	cmd.Flags().Int32VarP(&req.Limit, "limit", "l", 10, "每页条数")
	rootCmd.AddCommand(cmd)
}
