package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

var (
	SelectCache = make(map[string]*terminal.Select, 0)
)

func NewBox(t *terminal.Terminal, StartX, StartY, EndX, EndY int) *Box {
	box := t.NewBox(
		StartX,
		StartY,
		EndX,
		EndY,
		t.StyleDefault,
	)
	Select := &terminal.Select{
		StartX:    StartX + 1,
		StartY:    StartY + 1,
		MaxWidth:  box.Width(),
		MaxHeight: box.Height(),
		StyleSelect: tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorDarkCyan),
	}

	return &Box{
		t:      t,
		Box:    box,
		Select: Select,
	}
}

type Box struct {
	Box                 *terminal.Box
	Select              *terminal.Select
	EmptySelectFillText string // select 为空时需要填充的内容
	Dir                 string
	SelectPath          string // 中间需要选中的地址
	File                *bdpan.FileInfoDto
	UseCache            bool

	t *terminal.Terminal
}

func (b *Box) DrawBox() *Box {
	b.t.DrawBox(*b.Box)
	b.Box.Clean()
	return b
}

func (b *Box) SaveCache() {
	SelectCache[b.Dir] = b.Select
}

func (b *Box) EnableUseCache() *Box {
	b.UseCache = true
	return b
}

func (b *Box) SetFile(file *bdpan.FileInfoDto) *Box {
	b.File = file
	if file.IsDir() {
		b.Dir = file.Path
	} else {
		b.Dir = filepath.Dir(file.Path)
		b.SelectPath = file.Path
	}
	return b
}

func (b *Box) SetDir(dir string) *Box {
	b.Dir = dir
	return b
}

func (b *Box) SetSelectPath(path string) *Box {
	b.SelectPath = path
	return b
}

func (b *Box) SetEmptySelectFillText(text string) *Box {
	b.EmptySelectFillText = text
	return b
}

// 填充 select 数据
func (b *Box) FillSelect() error {
	s := b.Select
	if b.Dir == "" {
		return nil
	}
	if len(s.Items) == 0 {
		if b.UseCache {
			selCache, ok := SelectCache[b.Dir]
			if ok {
				s.Items = selCache.Items
				s.SelectIndex = selCache.SelectIndex
				return nil
			}
		}
		files, err := bdpan.GetDirAllFiles(b.Dir)
		if err != nil {
			return err
		}
		var items = make([]*terminal.SelectItem, 0)
		for i, f := range files {
			item := &terminal.SelectItem{
				Info: f,
			}
			items = append(items, item)
			if f.Path == b.SelectPath {
				s.SelectIndex = i
			}
		}
		s.Items = items
		b.SaveCache()
	}
	return nil
}

func (b *Box) DrawSelect(anchorIndex int, selectFn func(*bdpan.FileInfoDto)) *Box {
	// 没有数据时直接返回
	b.Box.Clean()
	s := b.Select
	// 计算获取需要绘制的列表开始索引
	// TODO: 需要重构
	var offset int = 0
	selectItemLen := len(b.Select.Items)
	selectMaxH := b.Select.MaxHeight
	if selectItemLen > selectMaxH {
		selectI := b.Select.SelectIndex
		selectOffset := selectMaxH - anchorIndex
		if selectI+selectOffset > selectMaxH {
			offset = selectI + selectOffset - selectMaxH
			if offset > selectItemLen-selectMaxH {
				offset = selectItemLen - selectMaxH
			}
		}
	}
	drawItems := b.Select.GetDrawItems(offset)
	if drawItems == nil || len(drawItems) == 0 {
		// 没有内容时进行填充
		b.Box.DrawOneLineText(0, b.t.StyleDefault, b.EmptySelectFillText)
		return b
	}
	// 填充 select 信息
	for i, item := range drawItems {
		info := item.Info.(*bdpan.FileInfoDto)
		text := fmt.Sprintf(" %s %s", info.GetFileTypeIcon(), info.GetFilename())
		style := b.t.StyleDefault
		if i+offset == b.Select.SelectIndex {
			style = b.Select.StyleSelect
		}
		b.Box.DrawOneLineText(i, style, text)
	}
	// 绘制选中状态的额外操作
	selectItem := s.GetSeleteItem()
	if selectItem != nil {
		info := selectItem.Info.(*bdpan.FileInfoDto)
		if selectFn != nil {
			selectFn(info)
		}
	}
	return b
}
