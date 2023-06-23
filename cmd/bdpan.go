package cmd

import (
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/cli"
	"github.com/wxnacy/bdpan-cli/terminal"
)

var (
	bdpanCommand = &BdpanCommand{}
)

type BdpanCommand struct {
	// 参数
	Path string

	client *cli.Client
}

func (r *BdpanCommand) Exec(args []string) error {
	var err error
	var file *bdpan.FileInfoDto
	file = &bdpan.FileInfoDto{
		Path:     r.Path,
		FileType: 1,
	}
	if r.Path != "/" {
		file, err = bdpan.GetFileByPath(r.Path)
		if err != nil {
			return err
		}
	}
	bdpan.SetOutputFile()
	// bdpan.SetLogLevel(logrus.DebugLevel)
	t, err := terminal.NewTerminal()
	if err != nil {
		return err
	}
	defer t.Quit()
	r.client = cli.NewClient(t).SetMidFile(file)
	return r.client.Exec()
}
