package terminal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/wxnacy/bdpan"
)

var (
	Log          = bdpan.GetLogger()
	StyleDefault = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
)

func DrawText(s tcell.Screen, StartX, StartY, EndX, EndY int, style tcell.Style, text string) error {
	x1, y1, x2, y2 := StartX, StartY, EndX, EndY
	row := y1
	col := x1
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
	return nil
}

func NewTerminal() (*Terminal, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	StyleDefault := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	s.SetStyle(StyleDefault)
	return &Terminal{
		S:            s,
		StyleDefault: StyleDefault,
	}, nil
}

type Terminal struct {
	S            tcell.Screen
	StyleDefault tcell.Style
}

func (t *Terminal) Exec() error {
	return nil
}

func (t *Terminal) DrawLineText(StartX, StartY, MaxLineW int, style tcell.Style, text string) error {
	text = OmitString(text, MaxLineW)
	text = FillString(text, MaxLineW)
	t.S.SetCell(StartX, StartY, style, []rune(text)...)
	return nil
}

func (t *Terminal) DrawOneLineText(StartY int, style tcell.Style, text string) error {
	w, _ := t.S.Size()
	t.DrawLineText(0, StartY, w, style, text)
	return nil
}

func (t *Terminal) DrawText(StartX, StartY, EndX, EndY int, style tcell.Style, text string) error {
	x1, y1, x2, y2 := StartX, StartY, EndX, EndY
	row := y1
	col := x1
	for _, r := range []rune(text) {
		t.S.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
	return nil
}

func (t *Terminal) DrawBox(b Box) error {
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
			t.S.SetContent(col, row, ' ', nil, style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		t.S.SetContent(col, y1, tcell.RuneHLine, nil, style)
		t.S.SetContent(col, y2, tcell.RuneHLine, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		t.S.SetContent(x1, row, tcell.RuneVLine, nil, style)
		t.S.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		t.S.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		t.S.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		t.S.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
		t.S.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
	}
	// return t.DrawText(x1+1, y1+1, x2-1, y2-1, style, text)
	return nil
}

func (t *Terminal) NewBox(StartX, StartY, EndX, EndY int, Style tcell.Style) *Box {
	return &Box{
		S:      t.S,
		StartX: StartX,
		StartY: StartY,
		EndX:   EndX,
		EndY:   EndY,
		Style:  Style,
	}
}

func (t *Terminal) Quit() {
	maybePanic := recover()
	t.S.Fini()
	if maybePanic != nil {
		// fmt.Println(maybePanic)
		panic(maybePanic)
	}
}
