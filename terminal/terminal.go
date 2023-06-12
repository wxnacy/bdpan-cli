package terminal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/wxnacy/bdpan"
)

var (
	Log          = bdpan.GetLogger()
	StyleDefault = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	StyleSelect  = tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite)
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
	Log.Infof("Terminal DrawLineText StartX: %d StartY: %d MaxLineW: %d Text: %s", StartX, StartY, MaxLineW, text)
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

// 该方法会删除
// 建议使用 Box.Draw()
func (t *Terminal) DrawBox(b Box) error {
	b.Draw()
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
