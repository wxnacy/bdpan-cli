package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
)

var backupReq = dto.NewBackupReq()

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "备份本地文件夹到网盘",
	Long: `
备份本地文件夹到网盘
每次执行将在指定 --path 下生成 Backups/2006-01-02-150405 格式目录进行备份

bdpan backup --local 本地文件夹 --path 网盘目录
	`,
	Run: func(cmd *cobra.Command, args []string) {
		backupReq.GlobalReq = *GetGlobalReq()
		handleCmdErr(handler.GetFileHandler().CmdBackup(backupReq))
	},
}

func init() {
	backupCmd.Flags().StringVarP(&backupReq.Local, "local", "l", "", "本地文件夹")
	rootCmd.AddCommand(backupCmd)
}
