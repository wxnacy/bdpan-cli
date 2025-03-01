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

func (i FileInfo) Id() string {
	return strconv.Itoa(int(i.FSID))
}

func (i FileInfo) Name() string {
	return i.GetFilename()
}

func (i FileInfo) String() string {
	text := fmt.Sprintf("%s %s", i.GetFileTypeIcon(), i.GetFilename())
	endSpaceCount := 2
	if i.IsSync {
		endSpaceCount += 2
	}
	if i.Count > 0 {
		endSpaceCount += len(strconv.Itoa(i.Count)) + 1
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
	text = fmt.Sprintf(" %s ", text)
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
	// files, err := handler.GetFileHandler().GetDirAllFiles(dir)
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

func (i SyncInfo) Id() string {
	return i.ID
}

func (i SyncInfo) Name() string {
	return i.ID
}

func (i SyncInfo) String() string {
	return fmt.Sprintf("%s    %s", i.ID, i.Local)
	// return i.Local
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

type SystemAction int

const (
	ActionSystem SystemAction = iota
	ActionFile
	ActionBigFile
	ActionSync
	ActionCollect
)

func NewSystemSelectItem(icon, name string, action SystemAction) *terminal.SelectItem {
	return &terminal.SelectItem{
		Info: &SystemInfo{
			Icon:   icon,
			name:   name,
			Action: action,
		},
	}
}

func FillSystemToSelect(s *terminal.Select, action SystemAction) error {
	s.Items = append(s.Items, NewSystemSelectItem("", "网盘文件", ActionFile))
	s.Items = append(s.Items, NewSystemSelectItem("", "收藏", ActionCollect))
	s.Items = append(s.Items, NewSystemSelectItem("", "同步", ActionSync))
	s.Items = append(s.Items, NewSystemSelectItem("", "查看大文件", ActionBigFile))
	s.SelectIndex = int(action) - 1
	return nil
}

type SystemInfo struct {
	name   string
	Icon   string
	Action SystemAction
}

func (i SystemInfo) Id() string {
	return i.name
}

func (i SystemInfo) Name() string {
	return i.name
}

func (i SystemInfo) String() string {
	return fmt.Sprintf(" %s %s", i.Icon, i.Name())
}
