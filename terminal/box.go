package terminal

import (
	"github.com/gdamore/tcell/v2"
)

type Box struct {
	S      tcell.Screen
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

// 清除内容
func (b *Box) Clean() *Box {
	for i := 0; i < b.Height(); i++ {
		b.DrawOneLineText(i, StyleDefault, "")
	}
	b.S.Show()
	return b
}

// 宽度
func (b Box) Width() int {
	sx, _, ex, _ := b.DrawRange()
	return ex - sx + 1
}

// 高度
func (b Box) Height() int {
	_, sy, _, ey := b.DrawRange()
	return ey - sy + 1
}

// 一行的数据，超出部分使用 ... 省略
func (b *Box) OmitOneLineText(text string) string {
	return OmitString(text, b.Width())
}

// 一行的数据，不够的部分使用空格填充
func (b *Box) FillOneLineText(text string) string {
	return FillString(text, b.Width())
}

// 绘制多行数据
func (b *Box) DrawMultiLineText(style tcell.Style, text []string) {
	for i, t := range text {
		b.DrawOneLineText(i, style, t)
	}
}

// 绘制当行数据
func (b *Box) DrawOneLineText(StartY int, style tcell.Style, text string) {
	sx, sy, _, _ := b.DrawRange()
	text = b.FillOneLineText(b.OmitOneLineText(text))
	DrawLine(b.S, sx, sy+StartY, style, text)
}

// 绘制多行数据
func (b *Box) DrawText(StartX, StartY int, style tcell.Style, text string) {
	sx, sy, _, _ := b.DrawRange()
	sx += StartX
	sy += StartY
	Log.Infof("Box: %v DrawText StartX: %d StartY: %d Text: %s", b, sx, sy, text)
	b.S.SetCell(sx, sy, style, []rune(text)...)
}
