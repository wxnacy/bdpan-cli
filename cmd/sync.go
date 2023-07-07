/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan"
)

var (
	syncCommand = &SyncCommand{}
)

const (
	ActionSync bdpan.SelectAction = iota
)

type SyncCommand struct {
	ID string

	IsCmdDel  bool
	IsCmdList bool
}

func (s SyncCommand) getModelSlice() []*bdpan.SyncModel {
	models := bdpan.GetSyncModels()
	modelSlice := make([]*bdpan.SyncModel, 0)
	for _, f := range models {
		modelSlice = append(modelSlice, f)
	}
	slice := bdpan.SyncModelSlice(modelSlice)
	sort.Sort(slice)
	return slice
}

func (s SyncCommand) getModelItems() []*bdpan.SelectItem {
	models := s.getModelSlice()
	items := make([]*bdpan.SelectItem, 0)
	for _, m := range models {
		item := &bdpan.SelectItem{
			Name:   m.Remote,
			Desc:   m.Desc(),
			Info:   m,
			Action: bdpan.ActionSystem,
		}
		items = append(items, item)
	}
	return items
}

func (s SyncCommand) selectSync() error {
	models := s.getModelItems()
	handle := func(item *bdpan.SelectItem) error {
		return s.handleAction(item)
	}
	return bdpan.PromptSelect("所有同步任务", models, true, 10, func(index int, s string) error {
		item := models[index]
		return handle(item)
	})
}

func (s SyncCommand) handleAction(item *bdpan.SelectItem) error {
	var err error
	switch item.Action {
	case bdpan.ActionSystem:
		return s.selectSystem(item)
	case bdpan.ActionBack:
		return s.selectSync()
	case ActionSync:
		m := item.Info.(*bdpan.SyncModel)
		err = m.Exec()
		if err != nil {
			return err
		}
		return s.selectSync()
	case bdpan.ActionDelete:
		m := item.Info.(*bdpan.SyncModel)
		models := bdpan.GetSyncModels()
		m, flag := models[m.ID]
		if !flag {
			return fmt.Errorf("%s 不存在", m.ID)
		}
		fmt.Println(m.Desc())
		flag = bdpan.PromptConfirm("确定删除")
		if !flag {
			return nil
		}
		err = bdpan.DeleteSyncModel(m.ID)
		if err != nil {
			return err
		}
		bdpan.PrintSyncModelList()

		return s.selectSync()
	}
	return nil
}

func (s SyncCommand) selectSystem(item *bdpan.SelectItem) error {
	systems := []*bdpan.SelectItem{
		&bdpan.SelectItem{
			Name:   "Back",
			Icon:   "",
			Desc:   "返回上一级",
			Info:   item.Info,
			Action: bdpan.ActionBack,
		},
		&bdpan.SelectItem{
			Name:   "Sync",
			Icon:   "",
			Desc:   "进行一次同步操作",
			Info:   item.Info,
			Action: ActionSync,
		},
		&bdpan.SelectItem{
			Name:   "Delete",
			Icon:   "",
			Desc:   "删除操作",
			Info:   item.Info,
			Action: bdpan.ActionDelete,
		},
	}
	handle := func(item *bdpan.SelectItem) error {
		return s.handleAction(item)
	}
	return bdpan.PromptSelect("操作列表", systems, true, 5, func(index int, s string) error {
		item := systems[index]
		return handle(item)
	})
}

func (s SyncCommand) Run() error {
	Log.Debugf("arg: %#v", s)
	if s.IsCmdList {
		bdpan.PrintSyncModelList()
	} else if s.IsCmdDel {
		if s.ID == "" {
			return fmt.Errorf("--delete 缺少参数 --id")
		}
		models := bdpan.GetSyncModels()
		m, flag := models[s.ID]
		if !flag {
			return fmt.Errorf("%s 不存在", s.ID)
		}
		fmt.Println(m.Desc())
		flag = bdpan.PromptConfirm("确定删除")
		if !flag {
			return nil
		}
		err := bdpan.DeleteSyncModel(m.ID)
		if err != nil {
			return err
		}
		bdpan.PrintSyncModelList()
		return nil
	} else {
		return s.selectSync()
	}
	return nil
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "同步文件夹",
	Long:  `可以对本地和远程文件夹做同步和备份两种操作`,
	Run: func(cmd *cobra.Command, args []string) {
		handleCmdErr(syncCommand.Run())
	},
}

func init() {
	syncCmd.Flags().StringVarP(&syncCommand.ID, "id", "", "", "任务 id")
	syncCmd.Flags().BoolVarP(&syncCommand.IsCmdDel, "delete", "", false, "删除同步目录")
	syncCmd.Flags().BoolVarP(&syncCommand.IsCmdList, "list", "", false, "列出同步目录")
	rootCmd.AddCommand(syncCmd)
}
