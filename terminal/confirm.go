package terminal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func NewConfirm(t *Terminal, text string) *Confirm {
	c := &Confirm{
		Text:           text,
		t:              t,
		minPaddingLeft: 2,
		Ensure:         &ConfirmSelect{Text: "确定(y)", StyleSelect: StyleSelect},
		Cancel:         &ConfirmSelect{Text: "取消(N)", StyleSelect: StyleSelect, IsSelect: true},
	}
	return c.init()
}

type Confirm struct {
	Box    *Box
	Text   string
	Ensure *ConfirmSelect
	Cancel *ConfirmSelect

	t              *Terminal
	minPaddingLeft int
	cancelTextLen  int
}

func (c *Confirm) init() *Confirm {
	w, h := c.t.S.Size()
	msgW := runewidth.StringWidth(c.Text)
	var boxW, boxH = msgW + c.minPaddingLeft*2 + 1, 5
	c.cancelTextLen = runewidth.StringWidth(c.Cancel.Text)
	var minBoxW = c.minPaddingLeft*4 + runewidth.StringWidth(c.Ensure.Text) + c.cancelTextLen
	Log.Infof("Confirm: %v init TextBoxW: %d TextBoxH: %d MinBoxW: %d", c, boxW, boxH, minBoxW)
	if boxW < minBoxW {
		boxW = minBoxW
	}
	startX := w/2 - boxW
	startY := h/2 - boxH
	box := c.t.NewBox(
		startX,
		startY,
		startX+boxW,
		startY+boxH,
		c.t.StyleDefault,
	)
	c.Box = box

	return c
}

func (c Confirm) IsEnsure() bool {
	if c.Ensure.IsSelect {
		return true
	}
	return false
}

func (c *Confirm) EnableEnsure() *Confirm {
	c.Ensure.IsSelect = true
	c.Cancel.IsSelect = false
	return c
}

func (c *Confirm) EnableCancel() *Confirm {
	c.Ensure.IsSelect = false
	c.Cancel.IsSelect = true
	return c
}

func (c *Confirm) Draw() *Confirm {
	c.t.DrawBox(*c.Box)
	c.Box.Clean()
	// text
	c.Box.DrawText(c.minPaddingLeft, 1, c.t.StyleDefault, c.Text)
	// ensure text
	c.Box.DrawText(c.minPaddingLeft, 3, c.Ensure.Style(), c.Ensure.Text)
	// cancel text
	boxStartX, _, boxEndX, _ := c.Box.GetDrawRange()
	cancelStartX := boxEndX - c.cancelTextLen - c.minPaddingLeft - boxStartX
	c.Box.DrawText(cancelStartX, 3, c.Cancel.Style(), c.Cancel.Text)
	return c
}

type ConfirmSelect struct {
	Text        string
	IsSelect    bool
	StyleSelect tcell.Style
}

func (s ConfirmSelect) Style() tcell.Style {
	if s.IsSelect {
		return s.StyleSelect
	}
	return StyleDefault
}
