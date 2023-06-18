package cli

import (
	"fmt"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

type FileInfo struct {
	*bdpan.FileInfoDto
	MaxTextWidth int
	IsSync       bool
}

func (i FileInfo) String() string {
	text := fmt.Sprintf(" %s %s", i.GetFileTypeIcon(), i.GetFilename())
	if i.IsSync {
		text = terminal.OmitString(text, i.MaxTextWidth-2)
		text = terminal.FillString(text, i.MaxTextWidth-2)
		text += ""
	}
	return text
}

func ConverFilesToSelectItems(s *terminal.Select, files []*bdpan.FileInfoDto) []*terminal.SelectItem {
	var items = make([]*terminal.SelectItem, 0)
	boxWidth := s.Box.Width()
	for _, f := range files {
		item := &terminal.SelectItem{
			Info: &FileInfo{
				FileInfoDto:  f,
				IsSync:       IsSync(f),
				MaxTextWidth: boxWidth,
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
	CacheSelectMap[dir] = s
	CacheFilesMap[dir] = files
	return nil
}

// 填充文件数据到 select 组件
func FillCacheToSelect(s *terminal.Select, dir, selectPath string) error {
	cacheSelect, ok := CacheSelectMap[dir]
	if ok {
		Log.Infof("FillCacheToSelect Dir %s SelectPath %s", dir, selectPath)
		// s.SetItems(cacheSelect.Items)
		s.SetItems(ConverFilesToSelectItems(s, CacheFilesMap[dir]))
		s.SelectIndex = cacheSelect.SelectIndex
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
