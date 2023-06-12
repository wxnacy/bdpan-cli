/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan"
)

var syncAddCommand = &SyncAddCommand{}

type SyncAddCommand struct {
	Remote   string
	Local    string
	IsBackup bool // 是否为备份
	HasHide  bool //是否备份隐藏文件
}

func (s SyncAddCommand) Run() error {
	mode := bdpan.ModeSync
	if s.IsBackup {
		mode = bdpan.ModeBackup
	}

	model := bdpan.NewSyncModel(s.Local, s.Remote, mode)
	model.HasHide = s.HasHide
	Log.Debugf("add model: %#v", model)

	key := model.ID
	models := bdpan.GetSyncModels()
	_, exits := models[key]
	if exits {
		return fmt.Errorf("已存在该记录")
	}
	models[key] = model
	err := bdpan.SaveModels(models)
	if err != nil {
		return err
	}
	bdpan.PrintSyncModelList()
	return nil
}

// syncAddCmd represents the syncAdd command
var syncAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加同步任务",
	Run: func(cmd *cobra.Command, args []string) {
		handleCmdErr(syncAddCommand.Run())
	},
}

func init() {
	syncAddCmd.Flags().StringVarP(&syncAddCommand.Remote, "remote", "r", "", "远程文件夹")
	syncAddCmd.Flags().StringVarP(&syncAddCommand.Local, "local", "L", "", "本地文件夹")
	syncAddCmd.Flags().BoolVarP(&syncAddCommand.HasHide, "hide", "H", false, "是否包含隐藏文件")
	syncAddCmd.Flags().BoolVarP(&syncAddCommand.IsBackup, "backup", "", false, "是否为备份目录")
	syncAddCmd.MarkFlagsRequiredTogether("remote", "local")
	syncCmd.AddCommand(syncAddCmd)
}
