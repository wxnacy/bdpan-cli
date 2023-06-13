package cli

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

var (
	Log = bdpan.Log
)

func NewClient(t *terminal.Terminal) *Client {
	return &Client{
		t:    t,
		mode: ModeNormal,
	}
}

type Client struct {
	t *terminal.Terminal
	// 模式
	mode Mode
	// 上层文件界面
	leftTerm *terminal.Select
	// 快捷键界面
	keymapTerm *terminal.List
	// 帮助界面
	helpTerm *terminal.Help
	// 上个键位
	prevRune rune
}

func (c Client) Size() (w, h int) {
	w, h = c.t.S.Size()
	return
}

func (c *Client) SetPrevRune(r rune) *Client {
	c.prevRune = r
	return c
}

func (c *Client) ClearPrevRune() *Client {
	return c.SetPrevRune(0)
}

func (c *Client) SetMode(m Mode) *Client {
	c.mode = m
	return c
}

func (c *Client) GetModeDrawRange() (StartX, StartY, EndX, EndY int) {
	w, h := c.Size()
	return 0, 1, w - 1, h - 2
}

func (c *Client) Draw() {
	// draw before
	switch c.mode {
	case ModeHelp:
		c.DrawHelp()
	}
	// draw common
	// draw after
	switch c.mode {
	case ModeKeymap:
		c.DrawKeymap()
	}
}

func (c *Client) DrawLeft() error {
	w, h := c.t.S.Size()
	c.leftTerm = terminal.NewEmptySelect(c.t, 0, 1, int(float64(w)*0.2), h-2).
		SetLoadingText("Load files...")
	c.leftTerm.DrawLoading()
	files, err := bdpan.GetDirAllFiles("/")
	if err != nil {
		return err
	}
	c.leftTerm.SetItems(ConverFilesToSelectItems(files))
	c.leftTerm.Draw()
	return nil
}

func (c *Client) DrawHelp() {
	c.helpTerm = terminal.NewHelp(c.t, GetHelpItems())
	sx, sy, ex, ey := c.GetModeDrawRange()
	c.helpTerm.Box.StartX = sx
	c.helpTerm.Box.StartY = sy
	c.helpTerm.Box.EndX = ex
	c.helpTerm.Box.EndY = ey
	c.helpTerm.Draw()
}

func (c *Client) DrawKeymap() {
	var data []string
	var keymaps = GetRelKeysByRune(c.prevRune)
	if keymaps != nil {
		data = GetRelKeysMsgByRune(c.prevRune)
	}
	_, h := c.t.S.Size()
	startY := h - 3 - len(data)
	c.keymapTerm = terminal.NewList(c.t, 0, startY, data).SetMaxWidth().Draw()
}

// 绘制消息
func (c *Client) DrawMessage(text string) {
	w, h := c.t.S.Size()
	maxLineW := int(float64(w) * 0.9)
	text = fmt.Sprintf("[%s] %s", strings.ToUpper(string(c.mode)), text)
	c.t.DrawLineText(0, h-1, maxLineW, terminal.StyleDefault, text)
	c.t.S.Show()
}

// 绘制标题
func (c *Client) DrawTitle(text string) {
	c.t.DrawOneLineText(0, terminal.StyleDefault, text)
}

func (c *Client) Exec() error {
	defer c.t.Quit()
	for {
		sx, sy, ex, ey := c.GetModeDrawRange()
		Log.Infof("Range %d %d %d %d", sx, sy, ex, ey)
		c.t.S.SetContent(sx, sy, '1', nil, terminal.StyleDefault)
		c.t.S.SetContent(ex, ey, '1', nil, terminal.StyleDefault)
		c.t.S.Show()
		ev := c.t.S.PollEvent()
		// Process event
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Rune() {
			case 'q':
				return nil
			}
		}
	}
}
