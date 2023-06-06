package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

func NewBox(t *terminal.Terminal, StartX, StartY, EndX, EndY int) *Box {
	boxWidth := EndX - StartX
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
		MaxWidth:  boxWidth - 2,
		MaxHeight: EndY - 2,
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

	t *terminal.Terminal
}

func (b *Box) DrawBox() *Box {
	b.t.DrawBox(*b.Box)
	// b.Box.DrawOneLineText(0, b.t.StyleDefault, "load files...")
	// b.t.S.Show()
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
	}
	return nil
}

func (b *Box) DrawSelect(selectFn func(*bdpan.FileInfoDto)) *Box {
	// 没有数据时直接返回
	s := b.Select
	drawItems := b.Select.GetDrawItems()
	if drawItems == nil || len(drawItems) == 0 {
		// 没有内容时进行填充
		b.Box.DrawOneLineText(0, b.t.StyleDefault, b.EmptySelectFillText)
		return b
	}
	// 填充 select 信息
	for i, item := range b.Select.GetDrawItems() {
		info := item.Info.(*bdpan.FileInfoDto)
		text := fmt.Sprintf(" %s %s", info.GetFileTypeIcon(), info.GetFilename())
		style := b.t.StyleDefault
		if i == b.Select.SelectIndex {
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
