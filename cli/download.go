package cli

import (
	"github.com/wxnacy/bdpan"
)

type DownloadCommand struct {
	From        string
	OutputDir   string
	OutputPath  string
	IsSync      bool
	IsRecursion bool
	IsVerbose   bool
}

func (d DownloadCommand) Download(file *bdpan.FileInfoDto) error {
	var err error
	Log.Debugf("是否同步: %v", d.IsSync)
	Log.Info("开始下载")
	if file.IsDir() {
		dlTasker := bdpan.NewDownloadTasker(file)
		dlTasker.Path = d.OutputPath
		if d.OutputDir != "" {
			dlTasker.Dir = d.OutputDir
		}
		dlTasker.IsRecursion = d.IsRecursion
		err = dlTasker.Exec()
		if err == nil {
			total := len(dlTasker.GetTasks())
			succ := total - len(dlTasker.GetErrorTasks())
			Log.Infof("下载完成: %d/%d", succ, total)
		}
	} else {
		dler := bdpan.NewDownloader()
		dler.UseProgressBar = true
		dler.Path = d.OutputPath
		if d.OutputDir != "" {
			dler.Dir = d.OutputDir
		}
		if d.IsVerbose {
			dler.EnableVerbose()
		}
		err = dler.DownloadFile(file)
	}
	if err != nil {
		Log.Error(err)
		return nil
	}
	return err
}

func (d DownloadCommand) Run() error {
	from := d.From
	file, err := bdpan.GetFileByPath(from)
	if err != nil {
		return err
	}
	return d.Download(file)
}
