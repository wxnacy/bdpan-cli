package cli

import (
	"fmt"
	"strconv"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

type FileInfo struct {
	*bdpan.FileInfoDto
	MaxTextWidth int
	IsSync       bool
	Count        int
}

func (i FileInfo) String() string {
	text := fmt.Sprintf(" %s %s", i.GetFileTypeIcon(), i.GetFilename())
	endSpaceCount := 0
	if i.IsSync {
		endSpaceCount += 2
	}
	if i.Count > 0 {
		endSpaceCount += len(strconv.Itoa(i.Count)) + 1
	}
	// 为最后空格留的位置
	if endSpaceCount > 0 {
		endSpaceCount += 1
	}
	wholeWidth := i.MaxTextWidth - endSpaceCount
	text = terminal.OmitString(text, wholeWidth)
	text = terminal.FillString(text, wholeWidth)
	if i.Count > 0 {
		text = fmt.Sprintf("%s %d", text, i.Count)
	}
	if i.IsSync {
		text += " "
	}
	if endSpaceCount > 0 {
		text += " "
	}
	return text
}

func ConverFilesToSelectItems(s *terminal.Select, files []*bdpan.FileInfoDto) []*terminal.SelectItem {
	var items = make([]*terminal.SelectItem, 0)
	boxWidth := s.Box.Width()
	for _, f := range files {
		count := 0
		if f.IsDir() {
			cache := GetFileSelectCache(f.Path)
			if cache != nil {
				count = len(cache.Files)
			}
		}
		item := &terminal.SelectItem{
			Info: &FileInfo{
				FileInfoDto:  f,
				IsSync:       IsSync(f),
				MaxTextWidth: boxWidth,
				Count:        count,
			},
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
	s.SetItems(ConverFilesToSelectItems(s, files))
	if selectPath != "" {
		for i, f := range files {
			if f.Path == selectPath {
				s.SetSelectIndex(i)
			}
		}
	}
	NewFileSelectCache(dir, s, files).Save()
	return nil
}

// 填充文件数据到 select 组件
func FillCacheToSelect(s *terminal.Select, dir, selectPath string) error {
	cache := GetFileSelectCache(dir)
	if cache != nil {
		Log.Infof("FillCacheToSelect Dir %s SelectPath %s", dir, selectPath)
		s.SetItems(ConverFilesToSelectItems(s, cache.Files))
		s.SelectIndex = cache.SelectIndex
		return nil
	}
	return FillFileToSelect(s, dir, selectPath)
}

type SyncInfo struct {
	*bdpan.SyncModel
	MaxTextWidth int
}

func (i SyncInfo) String() string {
	return fmt.Sprintf("%s    %s", i.ID, i.Local)
}

func FillSyncToSelect(s *terminal.Select, file *bdpan.FileInfoDto) error {
	syncModels := bdpan.GetSyncModelsByRemote(file.Path)
	var items = make([]*terminal.SelectItem, 0)
	boxWidth := s.Box.Width()
	for _, m := range syncModels {
		item := &terminal.SelectItem{
			Info: &SyncInfo{
				SyncModel:    m,
				MaxTextWidth: boxWidth,
			},
		}
		items = append(items, item)
	}
	s.SetItems(items)
	return nil
}
