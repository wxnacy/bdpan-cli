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

// 填充文件数据到 select 组件
func FillFileToSelect(s *terminal.Select, dir, selectPath string) error {
	Log.Infof("FillFileToSelect Dir %s SelectPath %s", dir, selectPath)
	files, err := bdpan.GetDirAllFiles(dir)
	if err != nil {
		return err
	}
	s.SetItems(ConverFilesToSelectItems(files))
	if selectPath != "" {
		for i, f := range files {
			if f.Path == selectPath {
				s.SetSelectIndex(i)
			}
		}
	}
	CacheSelectMap[dir] = s
	return nil
}

// 填充文件数据到 select 组件
func FillCacheToSelect(s *terminal.Select, dir, selectPath string) error {
	cacheSelect, ok := CacheSelectMap[dir]
	if ok {
		Log.Infof("FillCacheToSelect Dir %s SelectPath %s", dir, selectPath)
		s.SetItems(cacheSelect.Items)
		s.SelectIndex = cacheSelect.SelectIndex
		return nil
	}
	return FillFileToSelect(s, dir, selectPath)
}
