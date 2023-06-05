package terminal

import "github.com/gdamore/tcell/v2"

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

func (b Box) DrawRange() (StartX, StartY, EndX, EndY int) {
	return b.StartX + 1, b.StartY + 1, b.EndX - 1, b.EndY - 1
}

func (b *Box) DrawText(s tcell.Screen, style tcell.Style, text string) {
	sx, sy, ex, ey := b.DrawRange()
	DrawText(s, sx, sy, ex, ey, style, text)
}
