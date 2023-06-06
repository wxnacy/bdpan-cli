package cmd

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/sirupsen/logrus"
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeConfirm
)

var (
	bdpanCommand = &BdpanCommand{
		mode: ModeNormal,
	}
)

type BdpanCommand struct {
	// 参数
	Path string

	T    *terminal.Terminal
	mode Mode

	leftBox  *Box
	midBox   *Box
	rightBox *Box

	// 按键
	prevRune rune
}

func (r *BdpanCommand) initViewDir(file *bdpan.FileInfoDto) error {
	path := file.Path
	r.midBox.SetFile(file)
	if path != "/" {
		r.leftBox.SetDir(filepath.Dir(r.midBox.Dir)).SetSelectPath(r.midBox.Dir)
	}
	return nil
}

func (r *BdpanCommand) InitScreen(file *bdpan.FileInfoDto) error {
	var err error
	if err = r.DrawTopLeft(file.Path); err != nil {
		return err
	}
	if err = r.DrawLayout(); err != nil {
		return err
	}
	err = r.initViewDir(file)
	if err != nil {
		return err
	}
	if err = r.DrawSelect(); err != nil {
		return err
	}
	return nil
}

// 画布局
func (r *BdpanCommand) DrawLayout() error {
	t := r.T
	w, h := t.S.Size()
	Log.Debugf("window size (%d, %d)", w, h)
	// left box
	var boxWidth = int(float64(w) * 0.2)
	var startX = 0
	var startY = 1
	var endX = startX + boxWidth
	var endY = h - 2
	r.leftBox = NewBox(r.T, startX, startY, endX, endY).DrawBox()
	// mid box
	startX = endX
	boxWidth = int(float64(w) * 0.4)
	endX = startX + boxWidth
	r.midBox = NewBox(r.T, startX, startY, endX, endY).DrawBox()
	// right box
	startX = endX
	endX = startX + int(float64(w)*0.4)
	r.rightBox = NewBox(r.T, startX, startY, endX, endY).DrawBox()
	return nil
}

func (r *BdpanCommand) DrawSelect() error {
	err := r.midBox.FillSelect()
	if err != nil {
		return err
	}
	r.DrawMidSelect()

	err = r.leftBox.FillSelect()
	if err != nil {
		return err
	}
	r.leftBox.DrawSelect(nil)
	return nil
}

func (r *BdpanCommand) initSelect(s *terminal.Select, dir string) error {
	if len(s.Items) == 0 {
		files, err := bdpan.GetDirAllFiles(dir)
		if err != nil {
			return err
		}
		var items = make([]*terminal.SelectItem, 0)
		for _, f := range files {
			item := &terminal.SelectItem{
				Info: f,
			}
			items = append(items, item)
		}
		// items[0].IsSelect = true
		s.SelectIndex = 0
		s.Items = items
	}
	return nil
}

func (r *BdpanCommand) DrawEventKey(ev *tcell.EventKey) error {
	// 写入 rune
	runeStr := strings.ReplaceAll(strconv.QuoteRune(ev.Rune()), "'", "")
	if runeStr == " " {
		runeStr = "space"
	}
	err := r.DrawBottomRight(runeStr)
	if err != nil {
		return err
	}
	// 写入 key
	keyStr, ok := tcell.KeyNames[ev.Key()]
	if ok {
		err = r.DrawBottomRight(keyStr)
		if err != nil {
			return err
		}
	}
	return nil
}

// 绘制中间的 select
func (r *BdpanCommand) DrawMidSelect() {
	r.midBox.DrawSelect(func(info *bdpan.FileInfoDto) {
		r.rightBox.Box.DrawText(r.T.S, r.T.StyleDefault, info.GetPretty())

	})
}

// 左上角输入内容
func (r *BdpanCommand) DrawTopLeft(text string) error {
	w, _ := r.T.S.Size()
	return r.T.DrawText(0, 0, w-1, 0, r.T.StyleDefault, text)
}

// 左下角输入内容
func (r *BdpanCommand) DrawBottomLeft(text string) error {
	w, h := r.T.S.Size()
	return r.T.DrawText(0, h-1, w-10, h-1, r.T.StyleDefault, text)
}

// 右下角输入内容
func (r *BdpanCommand) DrawBottomRight(text string) error {
	w, h := r.T.S.Size()
	drawW := 10
	return r.T.DrawLineText(w-drawW-1, h-1, drawW, r.T.StyleDefault, text)
}

// 获取被选中的文件对象
func (r *BdpanCommand) GetSelectInfo() *bdpan.FileInfoDto {
	return r.getSelectInfo(r.midBox.Select)
}

func (r *BdpanCommand) getSelectInfo(s *terminal.Select) *bdpan.FileInfoDto {
	item := s.GetSeleteItem()
	info := item.Info.(*bdpan.FileInfoDto)
	Log.Infof("GetSelectInfo %s", info.Path)
	return info
}

func (r *BdpanCommand) ListenEventKeyInModeConfirm(ev *tcell.EventKey) error {
	// 处理退出的快捷键
	if ev.Rune() == 'q' || ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		r.mode = ModeNormal
		return ErrQuit
	}
	return nil
}
func (r *BdpanCommand) ListenEventKeyInModeNormal(ev *tcell.EventKey) error {
	// 处理退出的快捷键
	if ev.Rune() == 'q' || ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		return ErrQuit
	}
	err := r.DrawEventKey(ev)
	if err != nil {
		return err
	}
	switch ev.Rune() {
	case 'j':
		if r.midBox.Select.MoveDownSelect(1) {
			r.DrawMidSelect()
		}
	case 'k':
		if r.midBox.Select.MoveUpSelect(1) {
			r.DrawMidSelect()
		}
	case 'l':
		r.InitScreen(r.GetSelectInfo())
	case 'y':
		switch r.prevRune {
		case 0:
			r.prevRune = 'y'
		}
	case 'h':
		leftSelectFile := r.getSelectInfo(r.leftBox.Select)
		file := &bdpan.FileInfoDto{
			Path:     filepath.Dir(leftSelectFile.Path),
			FileType: 1,
		}
		r.InitScreen(file)
	default:
		switch ev.Key() {
		case tcell.KeyCtrlL:
			r.T.S.Sync()
		case tcell.KeyEnter:
			r.DrawBottomLeft("确定要下载?(y/N)")
			r.mode = ModeConfirm
		}
	}
	return nil
}

func (r *BdpanCommand) Exec(args []string) error {
	var err error
	var file *bdpan.FileInfoDto
	file = &bdpan.FileInfoDto{
		Path:     r.Path,
		FileType: 1,
	}
	if r.Path != "/" {
		file, err = bdpan.GetFileByPath(r.Path)
		if err != nil {
			return err
		}
	}
	bdpan.SetOutputFile()
	bdpan.SetLogLevel(logrus.DebugLevel)
	t, err := terminal.NewTerminal()
	if err != nil {
		return err
	}
	defer t.Quit()
	r.T = t
	r.InitScreen(file)
	for {
		// Update screen
		t.S.Show()
		// Poll event
		ev := t.S.PollEvent()
		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			t.S.Clear()
			t.S.Sync()
			r.InitScreen(r.midBox.File)
		case *tcell.EventKey:
			switch r.mode {
			case ModeNormal:
				err = r.ListenEventKeyInModeNormal(ev)
			case ModeConfirm:
				err = r.ListenEventKeyInModeConfirm(ev)
			}
			if err != nil {
				return err
			}
		}
	}
}
