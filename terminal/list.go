package terminal

import "github.com/mattn/go-runewidth"

func NewEmptyList(t *Terminal, StartX, StartY, EndX, EndY int) *List {
	l := &List{
		t: t,
		Box: t.NewBox(
			StartX,
			StartY,
			EndX,
			EndY,
			StyleDefault,
		),
	}
	return l
}

func NewList(t *Terminal, StartX, StartY int, data []string) *List {
	l := &List{
		t: t,
		Box: t.NewBox(
			StartX,
			StartY,
			StartX+1,
			StartY+len(data)+1,
			t.StyleDefault,
		),
		Data: data,
	}
	var maxListLen int
	for _, line := range data {
		lineLen := runewidth.StringWidth(line)
		if lineLen > maxListLen {
			maxListLen = lineLen
		}
	}
	// TODO: 处理中文的宽度，现在先临时做两倍处理
	maxListLen += maxListLen
	l.SetWidth(maxListLen)
	return l
}

type List struct {
	Box  *Box
	Data []string

	t *Terminal
}

func (l *List) SetWidth(w int) *List {
	l.Box.EndX = l.Box.StartX + w - 1
	return l
}

func (l *List) SetMaxWidth() *List {
	w, _ := l.Box.S.Size()
	return l.SetWidth(w)
}

func (l *List) SetData(data []string) *List {
	l.Data = data
	return l
}

func (l *List) Draw() *List {
	l.Box.Draw()
	l.Box.DrawMultiLineText(StyleDefault, l.Data)
	return l
}
