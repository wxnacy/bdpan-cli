package terminal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func NewBox(StartX, StartY, EndX, EndY int, Style tcell.Style) *Box {
	return &Box{
		StartX: StartX,
		StartY: StartY,
		EndX:   EndX,
		EndY:   EndY,
		Style:  Style,
	}
}

type Box struct {
	StartX int
	StartY int
	EndX   int
	EndY   int
	Style  tcell.Style
}

// 可以绘制的返回
func (b Box) DrawRange() (StartX, StartY, EndX, EndY int) {
	return b.StartX + 1, b.StartY + 1, b.EndX - 1, b.EndY - 1
}

// 宽度
func (b Box) Width() int {
	sx, _, ex, _ := b.DrawRange()
	return ex - sx
}

// 高度
func (b Box) Height() int {
	_, sy, _, ey := b.DrawRange()
	return ey - sy
}

// 一行的数据，超出部分使用 ... 省略
func (b *Box) OmitOneLineText(text string) string {
	textW := runewidth.StringWidth(text)
	boxW := b.Width()
	if textW > boxW {
		return text[0:boxW-3] + "..."
	}
	return text
}

// 一行的数据，不够的部分使用空格填充
func (b *Box) FillOneLineText(text string) string {
	return runewidth.FillRight(text, b.Width())
}

// 绘制多行数据
func (b *Box) DrawMultiLineText(s tcell.Screen, style tcell.Style, text []string) {
	for i, t := range text {
		b.DrawOneLineText(s, i, style, t)
	}
}

// 绘制当行数据
func (b *Box) DrawOneLineText(s tcell.Screen, StartY int, style tcell.Style, text string) {
	sx, sy, _, _ := b.DrawRange()
	text = b.FillOneLineText(b.OmitOneLineText(text))
	s.SetCell(sx, sy+StartY, style, []rune(text)...)
}