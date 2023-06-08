package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/sirupsen/logrus"
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeConfirm
	ModeKeymap
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

	leftBox   *Box
	midBox    *Box
	rightBox  *Box
	bottomBox *Box

	// 按键
	prevRune   rune
	prevAction KeymapAction
	useCache   bool

	fromFile *bdpan.FileInfoDto
}

func (r *BdpanCommand) initViewDir(file *bdpan.FileInfoDto) error {
	path := file.Path
	r.midBox.SetFile(file)
	if path != "/" {
		r.leftBox.SetDir(filepath.Dir(r.midBox.Dir)).SetSelectPath(r.midBox.Dir)
	}
	return nil
}

func (r *BdpanCommand) RefreshScreen() error {
	defer func() {
		r.useCache = false
	}()
	r.useCache = true
	return r.InitScreen(r.midBox.File)
}

func (r *BdpanCommand) ReloadScreen() error {
	return r.InitScreen(r.GetSelectInfo())
}

func (r *BdpanCommand) InitScreen(file *bdpan.FileInfoDto) error {
	Log.Infof("InitScreen UseCache: %v", r.useCache)
	r.T.S.Clear()
	r.T.S.Sync()
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
	// draw bottom
	switch r.mode {
	case ModeConfirm:
		switch r.prevAction {
		case KeymapActionDeleteFile:
			r.DrawBottomLeft(fmt.Sprintf("确定删除 %s (y/N)", r.GetSelectInfo().GetFilename()))
		}
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
	var bottomBoxH = 1
	var keymaps = GetRelKeysByRune(r.prevRune)
	if keymaps != nil {
		bottomBoxH = len(keymaps) + 1
	}
	var bottomBoxEndY = endY
	var bottomBoxStartY = endY - bottomBoxH
	if r.mode == ModeKeymap {
		endY = endY - bottomBoxH - 1
	}
	r.leftBox = NewBox(r.T, startX, startY, endX, endY).SetUseCache(r.useCache).DrawBox()
	// mid box
	startX = endX
	boxWidth = int(float64(w) * 0.4)
	endX = startX + boxWidth
	r.midBox = NewBox(r.T, startX, startY, endX, endY).
		SetUseCache(r.useCache).SetEmptySelectFillText("没有内容").DrawBox()
	// right box
	startX = endX
	endX = startX + int(float64(w)*0.4)
	r.rightBox = NewBox(r.T, startX, startY, endX, endY).DrawBox()
	if r.mode == ModeKeymap {
		r.bottomBox = NewBox(r.T, 0, bottomBoxStartY, w-1, bottomBoxEndY).DrawBox()
	}
	return nil
}

func (r *BdpanCommand) DrawSelect() error {
	// 绘制等待信息
	r.midBox.Box.DrawOneLineText(0, terminal.StyleDefault, "load files...")
	r.leftBox.Box.DrawOneLineText(0, terminal.StyleDefault, "load files...")
	r.T.S.Show()
	// 绘制 mid box
	err := r.midBox.FillSelect()
	if err != nil {
		return err
	}
	r.DrawMidSelect(5)
	// 绘制 left box
	err = r.leftBox.FillSelect()
	if err != nil {
		return err
	}
	r.leftBox.DrawSelect(5, nil)

	if r.mode == ModeKeymap {
		var keymaps = GetRelKeysByRune(r.prevRune)
		if keymaps != nil {
			r.bottomBox.Box.DrawMultiLineText(terminal.StyleDefault, GetRelKeysMsgByRune(r.prevRune))
		}
	}
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
func (r *BdpanCommand) DrawMidSelect(aIndex int) {
	r.midBox.DrawSelect(aIndex, func(info *bdpan.FileInfoDto) {
		r.rightBox.Box.DrawMultiLineText(
			r.T.StyleDefault, strings.Split(info.GetPretty(), "\n"))
	})
	r.DrawBottomLeft(r.GetSelectInfo().Path)
}

// 左上角输入内容
func (r *BdpanCommand) DrawTopLeft(text string) error {
	return r.T.DrawOneLineText(0, r.T.StyleDefault, text)
}

// 左下角输入内容
func (r *BdpanCommand) DrawBottomLeft(text string) error {
	w, h := r.T.S.Size()
	return r.T.DrawLineText(0, h-1, w/2, r.T.StyleDefault, text)
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
	if item == nil {
		return nil
	}
	info := item.Info.(*bdpan.FileInfoDto)
	Log.Infof("GetSelectInfo %s", info.Path)
	return info
}

func (r *BdpanCommand) MoveUp(step int) {
	if r.midBox.Select.MoveUpSelect(step) {
		r.DrawMidSelect(5)
	}
	r.midBox.SaveCache()
}

func (r *BdpanCommand) MoveDown(step int) {
	if r.midBox.Select.MoveDownSelect(step) {
		r.DrawMidSelect(r.midBox.Box.Height() - 5)
	}
	r.midBox.SaveCache()
}

func (r *BdpanCommand) MovePageEnd() {
	r.MoveDown(len(r.midBox.Select.Items))
}

func (r *BdpanCommand) CopyInModeNormal(msg string) error {
	err := clipboard.WriteAll(msg)
	if err != nil {
		return err
	}
	msg = fmt.Sprintf("%s 已经复制到剪切板", msg)
	r.mode = ModeNormal
	r.RefreshScreen()
	r.DrawBottomLeft(msg)
	return nil
}

func (r *BdpanCommand) ListenEventKeyInModeKeymap(ev *tcell.EventKey) error {
	var err error
	// 处理退出的快捷键
	if ev.Rune() == 'q' || ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		r.mode = ModeNormal
		return ErrQuit
	}
	keyString := fmt.Sprintf("%s%s", string(r.prevRune), string(ev.Rune()))
	r.DrawBottomRight(keyString)
	keyAction, ok := KeyActionMap[keyString]
	if ok {
		switch keyAction {
		case KeymapActionCopyPath:
			path := r.GetSelectInfo().Path
			r.CopyInModeNormal(path)
		case KeymapActionCopyName:
			name := r.GetSelectInfo().GetFilename()
			r.CopyInModeNormal(name)
		case KeymapActionCopyDir:
			path := r.GetSelectInfo().Path
			r.CopyInModeNormal(filepath.Dir(path))
		case KeymapActionCopyFile:
			r.fromFile = r.GetSelectInfo()
			msg := fmt.Sprintf("%s 已经复制", r.fromFile.Path)
			r.mode = ModeNormal
			r.RefreshScreen()
			r.DrawBottomLeft(msg)
		case KeymapActionPasteFile:
			if r.fromFile == nil {
				return ErrNotCopyFile
			}
			dir := filepath.Dir(r.GetSelectInfo().Path)
			toFile := filepath.Join(dir, r.fromFile.GetFilename())
			if r.prevAction == KeymapActionCutFile {
				err = bdpan.MoveFile(r.fromFile.Path, toFile)
			} else {
				err = bdpan.CopyFile(r.fromFile.Path, toFile)
			}
			if err != nil {
				return err
			}
			msg := fmt.Sprintf("%s 已经粘贴", r.fromFile.Path)
			r.mode = ModeNormal
			// TODO: 粘贴后需要将光标停留在指定文件
			r.ReloadScreen()
			r.DrawBottomLeft(msg)
			r.fromFile = nil
			r.prevAction = 0
		}
	} else {
		r.mode = ModeNormal
		r.RefreshScreen()
	}
	r.prevRune = 0
	return nil
}

func (r *BdpanCommand) ListenEventKeyInModeConfirm(ev *tcell.EventKey) error {
	var err error
	if ev.Rune() != 'y' {
		r.mode = ModeNormal
		r.RefreshScreen()
		r.DrawBottomLeft("操作取消!")
		return nil
	}
	var action = r.prevAction
	r.prevAction = 0
	switch action {
	case KeymapActionDeleteFile:
		err = bdpan.DeleteFile(r.GetSelectInfo().Path)
		r.mode = ModeNormal
		if err != nil {
			r.DrawBottomLeft(fmt.Sprintf("删除失败: %v", err))
		} else {
			r.ReloadScreen()
			r.DrawBottomLeft("删除成功!")
		}
	}
	return nil
}

func (r *BdpanCommand) ListenEventKeyInModeNormal(ev *tcell.EventKey) error {
	// 处理退出的快捷键
	if ev.Rune() == 'q' || ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		return ErrQuit
	}
	switch ev.Rune() {
	case 'j':
		r.MoveDown(1)
	case 'k':
		r.MoveUp(1)
	case 'l':
		r.InitScreen(r.GetSelectInfo())
	case 'h':
		leftSelectFile := r.getSelectInfo(r.leftBox.Select)
		file := &bdpan.FileInfoDto{
			Path:     filepath.Dir(leftSelectFile.Path),
			FileType: 1,
		}
		r.InitScreen(file)
	case 'G':
		r.MovePageEnd()
	case 'g':
		switch r.prevRune {
		case 0:
			r.prevRune = 'g'
		case 'g':
			r.MoveUp(len(r.midBox.Select.Items))
			r.prevRune = 0
		}
	case 'x':
		r.fromFile = r.GetSelectInfo()
		r.DrawBottomLeft(fmt.Sprintf("%s 已经剪切", r.fromFile.Path))
		r.prevAction = KeymapActionCutFile
	case 'd':
		r.fromFile = r.GetSelectInfo()
		r.prevAction = KeymapActionDeleteFile
		r.mode = ModeConfirm
		r.RefreshScreen()
	default:
		if IsKeymap(ev.Rune()) {
			r.prevRune = ev.Rune()
			r.mode = ModeKeymap
			r.RefreshScreen()
		} else {
			switch ev.Key() {
			case tcell.KeyCtrlL:
				r.RefreshScreen()
			case tcell.KeyCtrlD:
				_, h := r.T.S.Size()
				r.MoveDown(h / 2)
			case tcell.KeyCtrlF:
				_, h := r.T.S.Size()
				r.MoveDown(h)
			case tcell.KeyCtrlU:
				_, h := r.T.S.Size()
				r.MoveUp(h / 2)
			case tcell.KeyCtrlB:
				_, h := r.T.S.Size()
				r.MoveUp(h)
			case tcell.KeyEnter:
				r.DrawBottomLeft("确定要下载?(y/N)")
				r.mode = ModeConfirm
			}
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
			r.RefreshScreen()
		case *tcell.EventKey:
			Log.Infof("ListenEventKey Mode: %v Rune: %v(%s) prevRune: %v(%s) Key: %v",
				r.mode, ev.Rune(), strconv.QuoteRune(ev.Rune()),
				r.prevRune, strconv.QuoteRune(r.prevRune), ev.Key())
			err := r.DrawEventKey(ev)
			if err != nil {
				return err
			}
			switch r.mode {
			case ModeNormal:
				err = r.ListenEventKeyInModeNormal(ev)
			case ModeConfirm:
				err = r.ListenEventKeyInModeConfirm(ev)
			case ModeKeymap:
				err = r.ListenEventKeyInModeKeymap(ev)
			}
			if err != nil {
				if IsInErrors(err, BottomErrs) {
					r.DrawBottomLeft(err.Error())
				} else {
					return err
				}
			}
		}
	}
}
