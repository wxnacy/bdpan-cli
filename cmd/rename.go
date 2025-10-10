/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
doc: https://pan.baidu.com/union/doc/pkuo3snyp
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/logger"
)

func init() {
	cmd := &cobra.Command{
		Use:                   "rename [path] [newname]",
		Short:                 "重命名",
		Example:               `bdpan rename /apps/test/a.txt b.txt`,
		DisableFlagsInUseLine: true,
		Long:                  ``,
		Args:                  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			name := args[1]
			res, err := GetFileHandler().RenameFile(path, name)
			handleCmdErr(err)

			resInfo := res.Infos[0]
			if resInfo.IsError() {
				handleCmdErr(resInfo)
			}
			newPath := resInfo.ToPath
			logger.Printf("修改成功，新地址: %s", newPath)
			file, err := GetFileHandler().GetFileByPath(newPath)
			handleCmdErr(err)
			handler.PrintFileInfo(file)
		},
	}

	rootCmd.AddCommand(cmd)
}
