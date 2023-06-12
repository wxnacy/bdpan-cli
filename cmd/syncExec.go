/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan"
)

var syncExecCommand = &SyncExecCommand{SyncCommand: syncCommand}

type SyncExecCommand struct {
	*SyncCommand
	isOnce bool
	id     string
}

func (s SyncExecCommand) Run() error {
	fmt.Println("开始进行同步操作")
	for {
		for _, m := range bdpan.GetSyncModels() {
			if s.id != "" && s.id != m.ID {
				continue
			}
			err := m.Exec()
			if err != nil {
				return err
			}
		}
		if s.isOnce {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
}

// syncExecCmd represents the syncExec command
var syncExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "执行同步操作",
	Run: func(cmd *cobra.Command, args []string) {
		_, ok := bdpan.GetSyncModels()[syncExecCommand.id]
		if !ok {
			handleCmdErr(fmt.Errorf("ID: %s 不存在", syncExecCommand.id))
			return
		}
		handleCmdErr(syncExecCommand.Run())
	},
}

func init() {
	syncExecCmd.Flags().BoolVarP(&syncExecCommand.isOnce, "once", "o", false, "是否执行单次")
	syncExecCmd.Flags().StringVarP(&syncExecCommand.id, "id", "", "", "执行 id")
	syncCmd.AddCommand(syncExecCmd)
}
