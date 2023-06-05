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
	box := &terminal.Box{
		StartX: StartX,
		StartY: StartY,
		EndX:   EndX,
		EndY:   EndY,
		Style:  t.StyleDefault,
	}
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
	Box        *terminal.Box
	Select     *terminal.Select
	Dir        string
	SelectPath string // 中间需要选中的地址
	File       *bdpan.FileInfoDto

	t *terminal.Terminal
}

func (b *Box) DrawBox() *Box {
	b.t.DrawBox(*b.Box)
	b.Box.DrawText(b.t.S, b.t.StyleDefault, "load files...")
	b.t.S.Show()
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
	b.Box.DrawText(b.t.S, b.t.StyleDefault, "")
	b.t.S.Show()
	s := b.Select
	if b.Select.Items == nil {
		return nil
	}
	for i, item := range b.Select.GetDrawItems() {
		info := item.Info.(*bdpan.FileInfoDto)
		text := fmt.Sprintf(" %s %s", info.GetFileTypeIcon(), info.GetFilename())
		style := b.t.StyleDefault
		if i == b.Select.SelectIndex {
			style = b.Select.StyleSelect
		}
		b.t.DrawLineText(s.StartX, s.StartY+i, s.MaxWidth, style, text)
	}
	selectItem := s.GetSeleteItem()
	info := selectItem.Info.(*bdpan.FileInfoDto)
	if selectFn != nil {

		selectFn(info)
	}
	return b
}
