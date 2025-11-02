/*
Copyright 2023 NAME HERE EMAIL ADDRESS
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/config"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "清理缓存数据",
	// 功能需求:
	// - 通过 config.GetCacheDir() 获取 cache 目录
	// - 删除 cache 内所有内容，但是保留 cache 目录
	Run: func(cmd *cobra.Command, args []string) {
		dir := config.GetCacheDir()
		entries, err := os.ReadDir(dir)
		if err != nil {
			handleCmdErr(err)
			return
		}
		for _, e := range entries {
			p := filepath.Join(dir, e.Name())
			if err := os.RemoveAll(p); err != nil {
				handleCmdErr(err)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
