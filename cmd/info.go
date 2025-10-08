/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
doc: https://pan.baidu.com/union/doc/pkuo3snyp
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/pkg/bdtools"
)

func init() {
	cmd := &cobra.Command{
		Use:                   "info",
		Short:                 "展示文件",
		Example:               ``,
		DisableFlagsInUseLine: true,
		Long:                  ``,
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			file, err := bdtools.GetFileByPath(GetFileHandler().GetAccessToken(), path)
			if err == nil {
				err = bdtools.PrintFileInfo(file)
			}
			handleCmdErr(err)
		},
	}

	rootCmd.AddCommand(cmd)
}
