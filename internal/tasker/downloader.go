package tasker

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wxnacy/bdpan-cli/internal/api"
	"github.com/wxnacy/bdpan-cli/internal/common"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/dler/godler"
	"github.com/wxnacy/go-bdpan"
	"github.com/wxnacy/go-tools"
)

func DownloadFile(f *bdpan.FileInfo, path string) {
	begin := time.Now()
	tasker := NewDownloadTasker(f, path)
	tasker.BuildTasks()
	tasker.BeforeRun()
	tasker.Run(tasker.RunTask)
	tasker.AfterRun()
	out := fmt.Sprintf("下载完成，耗时：%v", time.Since(begin))
	fmt.Println(out)
}

type DownloadTaskInfo struct {
	// From string
	FSID uint64
	To   string
}

type DownloadTasker struct {
	*godler.Tasker
	// 迁移的地址
	File  *bdpan.FileInfo
	To    string
	Token string
}

func NewDownloadTasker(f *bdpan.FileInfo, path string) *DownloadTasker {
	t := DownloadTasker{Tasker: godler.NewTasker(godler.NewTaskerConfig())}
	t.File = f
	t.To = path
	t.Token = config.GetAccessToken()
	return &t
}

func (m *DownloadTasker) AfterRun() {
}

func (m *DownloadTasker) BuildTasks() {
	token := config.GetAccessToken()
	if m.File.IsDir() {
		files, err := api.GetAllFileList(token, m.File.Path)
		if err != nil {
			panic(err)
		}
		fmt.Printf("找到文件个数: %d\n", len(files))
		for _, f := range files {
			to := filepath.Join(m.To, f.GetFilename())
			info := DownloadTaskInfo{FSID: f.FSID, To: to}
			m.AddTask(&godler.Task{Info: info})
		}
	}
}

func (m DownloadTasker) RunTask(task *godler.Task) error {
	info := task.Info.(DownloadTaskInfo)
	if tools.FileExists(info.To) {
		return nil
	}
	f, err := api.GetFileInfo(m.Token, info.FSID)
	if err != nil {
		return err
	}
	return common.DownloadFile(f.Dlink, info.To)
}

func (m *DownloadTasker) BeforeRun() {
	if m.File.IsDir() {
		err := os.MkdirAll(m.To, 0o755)
		if err != nil {
			panic(err)
		}
	}
}
