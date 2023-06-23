package terminal

import (
	"github.com/gdamore/tcell/v2"
)

func NewBox(s tcell.Screen, StartX, StartY, EndX, EndY int) *Box {
	return &Box{
		S:      s,
		StartX: StartX,
		StartY: StartY,
		EndX:   EndX,
		EndY:   EndY,
		Style:  StyleDefault,
	}
}

type Box struct {
	S            tcell.Screen
	StartX       int
	StartY       int
	EndX         int
	EndY         int
	Style        tcell.Style
	PaddingLeft  int
	PaddingRight int
}

func (b *Box) SetPaddingLeft(p int) *Box {
	b.PaddingLeft = p
	return b
}

// 可以绘制的返回
func (b Box) GetDrawRange() (StartX, StartY, EndX, EndY int) {
	return b.StartX + 1 + b.PaddingLeft, b.StartY + 1, b.EndX - 1, b.EndY - 1
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
	sx, _, ex, _ := b.GetDrawRange()
	return ex - sx + 1
}

// 高度
func (b Box) Height() int {
	_, sy, _, ey := b.GetDrawRange()
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
	b.DrawLineText(StartY, style, text)
}

// 绘制当行数据
func (b *Box) DrawLineText(StartY int, style tcell.Style, text string) {
	sx, sy, _, _ := b.GetDrawRange()
	text = b.FillOneLineText(b.OmitOneLineText(text))
	DrawLine(b.S, sx, sy+StartY, style, text)
}

// 绘制多行数据
func (b *Box) DrawText(StartX, StartY int, style tcell.Style, text string) {
	sx, sy, _, _ := b.GetDrawRange()
	sx += StartX
	sy += StartY
	Log.Infof("Box: %v DrawText StartX: %d StartY: %d Text: %s", b, sx, sy, text)
	b.S.SetCell(sx, sy, style, []rune(text)...)
}

func (b *Box) Draw() {
	x1, y1, x2, y2, style := b.StartX, b.StartY, b.EndX, b.EndY, b.Style
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			b.S.SetContent(col, row, ' ', nil, style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		b.S.SetContent(col, y1, tcell.RuneHLine, nil, style)
		b.S.SetContent(col, y2, tcell.RuneHLine, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		b.S.SetContent(x1, row, tcell.RuneVLine, nil, style)
		b.S.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		b.S.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		b.S.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		b.S.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
		b.S.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
	}
	b.S.Show()
}
