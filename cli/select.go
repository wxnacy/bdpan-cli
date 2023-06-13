package cli

import (
	"fmt"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

type FileInfo struct {
	*bdpan.FileInfoDto
}

func (i FileInfo) String() string {
	return fmt.Sprintf(" %s %s", i.GetFileTypeIcon(), i.GetFilename())
}

func ConverFilesToSelectItems(files []*bdpan.FileInfoDto) []*terminal.SelectItem {
	var items = make([]*terminal.SelectItem, 0)
	for _, f := range files {
		item := &terminal.SelectItem{
			Info: &FileInfo{FileInfoDto: f},
		}
		items = append(items, item)
	}
	return items
}
