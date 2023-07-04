package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
	"github.com/wxnacy/go-tools"
)

var (
	Log = bdpan.Log
)

func NewClient(t *terminal.Terminal) *Client {
	c := &Client{
		t:            t,
		mode:         ModeNormal,
		normalAction: ActionFile,
	}
	return c
}

type Client struct {
	t *terminal.Terminal
	// 模式
	mode Mode
	m    ModeInterface
	// 动作
	normalPrevAction SystemAction
	normalAction     SystemAction
	// 上层文件界面
	leftTerm *terminal.Select
	// 当前文件界面
	midTerm *terminal.Select
	// 当前查看文件
	midFile *bdpan.FileInfoDto
	// 详情界面
	detailTerm *terminal.List
	// 快捷键界面
	keymapTerm *terminal.List
	// 帮助界面
	helpTerm *terminal.Help
	// 确认界面
	// confirmTerm *terminal.Confirm
	// 同步界面
	syncTerm *terminal.Select
	// 上个键位
	prevRune rune
	// 上个动作
	prevAction KeymapAction
	// 当前输入键位
	eventKey *tcell.EventKey
	// 是否需要缓存
	useCache bool
	// 选择文件列表
	selectFiles []*bdpan.FileInfoDto
}

func (c Client) Size() (w, h int) {
	w, h = c.t.S.Size()
	return
}

func (c *Client) EnableCache() *Client {
	c.useCache = true
	return c
}

func (c *Client) DisableCache() *Client {
	c.useCache = false
	return c
}

func (c *Client) SetPrevRune(r rune) *Client {
	c.prevRune = r
	return c
}

func (c *Client) ClearPrevRune() *Client {
	return c.SetPrevRune(0)
}

func (c *Client) SetNormalAction(a SystemAction) *Client {
	c.normalPrevAction = c.normalAction
	c.normalAction = a
	return c
}

func (c *Client) SetPrevAction(a KeymapAction) *Client {
	c.prevAction = a
	return c
}

func (c *Client) ClearPrevAction() *Client {
	return c.SetPrevAction(0)
}

func (c *Client) AppendSelectFile(file *bdpan.FileInfoDto) *Client {
	c.selectFiles = append(c.selectFiles, file)
	return c
}

func (c *Client) SetCurrSelectFiles() *Client {
	c.ClearSelectFiles().selectFiles = append(c.selectFiles, c.GetMidSelectFile())
	return c
}

func (c *Client) ClearSelectFiles() *Client {
	c.selectFiles = make([]*bdpan.FileInfoDto, 0)
	return c
}

func (c *Client) SetMode(m Mode) *Client {
	c.mode = m
	return c
}

func (c *Client) GetMode() Mode {
	if c.m != nil {
		return c.m.GetMode()
	}
	return c.mode
}

func (c *Client) SetMidFile(file *bdpan.FileInfoDto) *Client {
	c.midFile = file
	return c
}

func (c *Client) GetModeDrawRange() (StartX, StartY, EndX, EndY int) {
	w, h := c.Size()
	return 0, 1, w - 1, h - 2
}

// 获取 normal 模式下模型 EndY
func (c *Client) GetModeNormalEndY() int {
	endY := func() int {
		switch c.GetMode() {
		case ModeKeymap:
			return c.keymapTerm.Box.StartY - 1
		case ModeHelp:
			return c.helpTerm.Box.StartY - 1
		case ModeSync:
			return c.syncTerm.Box.StartY - 1
		default:
			_, _, _, ey := c.GetModeDrawRange()
			return ey
		}
	}()
	Log.Debugf("GetModeNormalEndY %d", endY)
	return endY
}

func (c *Client) GetMidDir() string {
	if c.midFile.IsDir() {
		return c.midFile.Path
	}
	return filepath.Dir(c.midFile.Path)
}

func (c *Client) GetLeftDir() string {
	return filepath.Dir(c.GetMidDir())
}

func (c *Client) GetMidSelectSystem() *SystemInfo {
	return c.midTerm.GetSeleteItem().Info.(*SystemInfo)
}

func (c *Client) GetMidSelectFile() *bdpan.FileInfoDto {
	return c.midTerm.GetSeleteItem().Info.(*FileInfo).FileInfoDto
}

func (c *Client) EnableModeNormal() *Client {
	c.m = nil
	return c.ClearPrevRune().ClearPrevAction().SetMode(ModeNormal)
}

func (c *Client) DrawCache() error {
	c.EnableCache()
	defer c.DisableCache()
	return c.Draw()
}

func (c *Client) DrawCacheNormal() error {
	return c.EnableModeNormal().DrawCache()
}

func (c *Client) DrawNormal() error {
	return c.EnableModeNormal().Draw()
}

func (c *Client) Draw() error {
	var err error
	c.t.S.Clear()
	c.t.S.Sync()

	// draw before
	c.DrawTitle(c.GetMidDir())
	c.DrawInputKey()
	switch c.GetMode() {
	case ModeHelp:
		c.DrawHelp()
		return nil
	case ModeKeymap:
		c.DrawKeymap()
	case ModeSync:
		c.DrawSync()
	}
	// draw common
	err = c.DrawLeft()
	if err != nil {
		return err
	}
	err = c.DrawMid()
	if err != nil {
		return err
	}
	// draw after
	switch c.GetMode() {
	case ModeConfirm:
		c.DrawConfirm()
	case ModeFilter:
		c.DrawFilter()
	case ModeCommand:
		c.DrawCommand()
	}
	return nil
}

func (c *Client) DrawLeft() error {
	var err error
	w, _ := c.t.S.Size()
	sx, sy, _, _ := c.GetModeDrawRange()
	c.leftTerm = terminal.
		NewEmptySelect(c.t, sx, sy, int(float64(w)*0.2), c.GetModeNormalEndY()).
		SetLoadingText("Load files...")
	c.leftTerm.DrawLoading()
	switch c.normalAction {
	case ActionFile:
		// 只有非根目录时才会展示左侧目录
		if c.GetMidDir() != "/" {
			if c.useCache {
				err = FillCacheToSelect(c.leftTerm, c.GetLeftDir(), c.GetMidDir())
			} else {
				err = FillFileToSelect(c.leftTerm, c.GetLeftDir(), c.GetMidDir())
			}
			if err != nil {
				return err
			}
		} else {
			FillSystemToSelect(c.leftTerm, c.normalAction)
		}
	case ActionSystem:
	default:
		FillSystemToSelect(c.leftTerm, c.normalAction)
	}
	c.leftTerm.Draw()
	return nil
}

func (c *Client) DrawMid() error {
	var err error
	w, _ := c.t.S.Size()
	_, sy, _, _ := c.GetModeDrawRange()
	startX := c.leftTerm.Box.EndX + 1
	endX := startX + int(float64(w)*0.4)
	c.midTerm = terminal.
		NewEmptySelect(c.t, startX, sy, endX, c.GetModeNormalEndY()).
		SetLoadingText("Load files...")
	c.midTerm.DrawLoading()
	switch c.normalAction {
	case ActionFile:
		c.midTerm.SetEmptyFillText("没有文件")
		if c.useCache {
			err = FillCacheToSelect(c.midTerm, c.GetMidDir(), c.midFile.Path)
		} else {
			err = FillFileToSelect(c.midTerm, c.GetMidDir(), c.midFile.Path)
		}
	case ActionSystem:
		FillSystemToSelect(c.midTerm, ActionFile)
	case ActionBigFile:
		files, err := GetAllLocalFiles()
		if err != nil {
			return err
		}
		files = FilterFileFiles(files)
		sort.Slice(files, func(i, j int) bool {
			return files[i].Size > files[j].Size
		})

		c.midTerm.SetItems(ConverFilesToSelectItems(c.midTerm, files))
	case ActionSync:
		var items = make([]*terminal.SelectItem, 0)
		boxWidth := c.midTerm.Box.Width()
		for _, m := range bdpan.GetSyncModelSlice() {
			item := &terminal.SelectItem{
				Info: &SyncInfo{
					SyncModel:    m,
					MaxTextWidth: boxWidth,
				},
			}
			items = append(items, item)
		}
		c.midTerm.SetItems(items)
	}
	if err != nil {
		return err
	}
	// 设置缓存的索引位置
	cacheindex, ok := GetCacheSelectIndex(c.normalAction)
	if c.useCache && ok {
		c.midTerm.SetSelectIndex(cacheindex)
	}
	// 过滤模式数据过滤
	if c.GetMode() == ModeFilter {
		c.midTerm.Filter(c.m.(*FilterMode).Filter)
	}
	c.DrawMidData()
	return nil
}

func (c *Client) DrawMidData() error {
	c.midTerm.SetSelectFn(func(item *terminal.SelectItem) {
		c.DrawDetail()
		switch c.normalAction {
		case ActionFile, ActionBigFile:
			c.DrawMessage(item.Info.(*FileInfo).Path)
		}
	}).Draw()
	return nil
}

func (c *Client) DrawDetail() {
	_, sy, ex, _ := c.GetModeDrawRange()
	startX := c.midTerm.Box.EndX + 1
	c.detailTerm = terminal.NewEmptyList(c.t, startX, sy, ex, c.GetModeNormalEndY())
	var detail string
	switch c.normalAction {
	case ActionFile, ActionBigFile:
		info := c.GetMidSelectFile()
		detail = info.GetPretty()
	case ActionSync:
		info := c.midTerm.GetSeleteItem().Info.(*SyncInfo)
		detail = info.Desc()
	}
	Log.Debugf("DrawDetail Info %s", detail)
	c.detailTerm.SetData(strings.Split(detail, "\n"))
	c.detailTerm.Draw()

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

func (c *Client) DrawConfirm() {
	c.m.(*ConfirmMode).Term.Draw()
}

// 绘制同步界面
func (c *Client) DrawSync() {
	syncTermH := len(bdpan.GetSyncModelsByRemote(c.GetMidSelectFile().Path))
	if syncTermH == 0 {
		c.EnableModeNormal()
		return
	}
	sx, _, ex, ey := c.GetModeDrawRange()
	startY := ey - syncTermH - 1
	c.syncTerm = terminal.
		NewEmptySelect(c.t, sx, startY, ex, ey)
	// 填充内容
	FillSyncToSelect(c.syncTerm, c.GetMidSelectFile())
	c.syncTerm.Draw()
}

// 绘制命令界面
func (c *Client) DrawCommand() {
	m := c.m.(*CommandMode)
	c.DrawInput(m.Prefix, m.Input)
}

// 绘制过滤界面
func (c *Client) DrawFilter() {
	// m := c.m.(*FilterMode)
	// c.DrawInput("/", m.Input)
}

// 绘制输入界面
func (c *Client) DrawInput(prefix, text string) {
	c.DrawMessage(fmt.Sprintf("%s%s|", prefix, text))
}

// 绘制消息
func (c *Client) DrawMessage(text string) {
	w, h := c.t.S.Size()
	maxLineW := int(float64(w) * 0.9)
	text = fmt.Sprintf("[%s] %s", strings.ToUpper(string(c.GetMode())), text)
	c.t.DrawLineText(0, h-1, maxLineW, terminal.StyleDefault, text)
	c.t.S.Show()
}

// 绘制标题
func (c *Client) DrawTitle(text string) {
	c.t.DrawOneLineText(0, terminal.StyleDefault, text)
}

// 绘制输入键位
func (c *Client) DrawInputKey() {
	w, h := c.t.S.Size()
	maxLineW := int(float64(w) * 0.1)
	var text string
	if c.eventKey == nil {
		return
	}
	keyName, ok := tcell.KeyNames[c.eventKey.Key()]
	if ok {
		text = keyName
	} else {
		inputRune := c.eventKey.Rune()
		text = string(inputRune)
	}
	c.t.DrawLineText(w-maxLineW, h-1, maxLineW, terminal.StyleDefault, text)
	c.t.S.Show()
}

func (c *Client) MoveUp(step int) {
	if c.midTerm.MoveUpSelect(step) {
		GetFileSelectCache(c.GetMidDir()).SelectIndex = c.midTerm.SelectIndex
		c.DrawMidData()
	}
}

func (c *Client) MoveDown(step int) {
	if c.midTerm.MoveDownSelect(step) {
		GetFileSelectCache(c.GetMidDir()).SelectIndex = c.midTerm.SelectIndex
		c.DrawMidData()
	}
}

func (c *Client) MoveLeft() {
	switch c.normalAction {
	case ActionFile:
		if c.GetMidDir() != "/" {
			c.midFile = &bdpan.FileInfoDto{
				Path:     c.GetLeftDir(),
				FileType: 1,
			}
			c.DrawCache()
		} else {
			c.ShowSystem()
		}
	default:
		c.ShowSystem()
	}
}

func (c *Client) Download() error {
	SetCacheSelectIndex(c.normalAction, c.midTerm.SelectIndex)
	var name = c.midTerm.GetSeleteItem().Info.Name()
	msg := fmt.Sprintf("确定下载 %s?", name)
	c.SetConfirmMode(msg).
		SetPrevAction(KeymapActionDownloadFile).
		SetCurrSelectFiles().DrawCache()
	return nil
}

func (c *Client) Enter() error {
	SetCacheSelectIndex(c.normalAction, c.midTerm.SelectIndex)
	switch c.normalAction {
	case ActionFile:
		c.midFile = c.GetMidSelectFile()
		if c.midFile.IsDir() {
			c.DrawCache()
		} else {
			c.Download()
		}
	case ActionBigFile:
		c.Download()
	case ActionSystem:
		systemInfo := c.GetMidSelectSystem()
		switch systemInfo.Action {
		case ActionFile:
			c.midFile = GetRootFile()
		}
		c.SetNormalAction(systemInfo.Action).DrawCacheNormal()
	case ActionSync:
		info := c.midTerm.GetSeleteItem().Info.(*SyncInfo)
		msg := fmt.Sprintf("确定执行 %s?", info.ID)
		c.SetConfirmMode(msg).SetPrevAction(KeymapActionSyncExec).DrawCache()
	}
	return nil
}

func (c *Client) ShowSystem() {
	c.SetNormalAction(ActionSystem).DrawCache()
}

func (c *Client) HandleNormalAction(action KeymapAction) error {
	var err error
	Log.Infof("HandleKeymapAction %v", action)
	switch action {
	case KeymapActionHelp:
		c.SetMode(ModeHelp).DrawCache()
	case KeymapActionFilter:
		fm := NewFilterMode("")
		fm.SetActionFn(c.HandleFilterAction)
		m := NewCommandMode("/").SetNextMode(fm)
		m.SetActionFn(c.HandleCommandAction)
		c.m = m
		c.SetMode(ModeCommand).DrawCache()
	case KeymapActionSync:
		c.SetMode(ModeSync).DrawCache()
	case KeymapActionReload:
		c.DrawNormal()
	// 向下移动
	case KeymapActionMoveDown:
		c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() - 5)
		c.MoveDown(1)
	case KeymapActionMoveDownHalfPage:
		c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() / 2)
		c.MoveDown(c.midTerm.Box.Height() / 2)
	case KeymapActionMoveDownPage:
		c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() - 1)
		c.MoveDown(c.midTerm.Box.Height())
	case KeymapActionMovePageEnd:
		c.MoveDown(c.midTerm.Length())
		// 向上移动
	case KeymapActionMoveUp:
		c.midTerm.SetAnchorIndex(5)
		c.MoveUp(1)
	case KeymapActionMoveUpHalfPage:
		c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() / 2)
		c.MoveUp(c.midTerm.Box.Height() / 2)
	case KeymapActionMoveUpPage:
		c.midTerm.SetAnchorIndex(0)
		c.MoveUp(c.midTerm.Box.Height())
	case KeymapActionMoveLeft:
		c.MoveLeft()
	case KeymapActionMoveLeftHome:
		c.midFile = GetRootFile()
		c.DrawCache()
	case KeymapActionEnter, KeymapActionMoveRight:
		err = c.Enter()
	case KeymapActionCutFile:
		c.SetCurrSelectFiles().SetPrevAction(action).DrawCacheNormal()
		fromFile := c.selectFiles[0]
		c.DrawMessage(fmt.Sprintf("%s 已经剪切", fromFile.Path))
	case KeymapActionDownloadFile:
		c.Download()
	case KeymapActionDeleteFile:
		var msg string
		var name = c.midTerm.GetSeleteItem().Info.Name()
		msg = fmt.Sprintf("确定删除 %s?", name)
		c.SetConfirmMode(msg).SetCurrSelectFiles().SetPrevAction(action).DrawCache()
	case KeymapActionKeymap:
		return c.SetPrevRune(c.eventKey.Rune()).
			SetMode(ModeKeymap).DrawCache()
	case KeymapActionSystem:
		c.ShowSystem()
	case KeymapActionQuit:
		return ErrQuit
	}
	return err
}

func (c *Client) HandleKeymapAction(action KeymapAction) error {
	var err error
	Log.Infof("HandleKeymapAction %v", action)
	switch action {
	case KeymapActionMovePageHome:
		c.DrawCacheNormal()
		c.MoveUp(c.midTerm.Length())
	case KeymapActionCopyPath:
		return c.ActionCopyMsg(c.GetMidSelectFile().Path)
	case KeymapActionCopyName:
		return c.ActionCopyMsg(c.GetMidSelectFile().GetFilename())
	case KeymapActionCopyDir:
		return c.ActionCopyMsg(filepath.Dir(c.GetMidSelectFile().Path))
	case KeymapActionCopyFile:
		c.SetCurrSelectFiles().SetPrevAction(action).DrawCacheNormal()
		fromFile := c.selectFiles[0]
		c.DrawMessage(fmt.Sprintf("%s 已经复制", fromFile.Path))
	case KeymapActionPasteFile:
		if len(c.selectFiles) == 0 {
			return ErrNotCopyFile
		}
		dir := filepath.Dir(c.GetMidSelectFile().Path)
		fromFile := c.selectFiles[0]
		toFile := filepath.Join(dir, fromFile.GetFilename())
		if c.prevAction == KeymapActionCutFile {
			err = bdpan.MoveFile(fromFile.Path, toFile)
		} else {
			err = bdpan.CopyFile(fromFile.Path, toFile)
		}
		if err != nil {
			return err
		}
		c.ClearSelectFiles().DrawNormal()
		c.DrawMessage(fmt.Sprintf("%s 已经粘贴", toFile))
	}
	return nil
}

func (c *Client) HandleSyncAction(action KeymapAction) error {
	switch action {
	case KeymapActionSyncExec:
		info := c.syncTerm.GetSeleteItem().Info.(*SyncInfo)
		err := info.Exec()
		if err != nil {
			Log.Errorf("SyncModel %s Exec Error: %v", info.ID, err)
			return ErrActionFail
		}
		c.DrawCacheNormal()
		c.DrawMessage(fmt.Sprintf("%s 同步成功", c.GetMidSelectFile().Path))
	case KeymapActionMoveDown:
		if c.syncTerm.MoveDownSelect(1) {
			c.syncTerm.Draw()
		}
	case KeymapActionMoveUp:
		if c.syncTerm.MoveUpSelect(1) {
			c.syncTerm.Draw()
		}
	}
	return nil
}

func (c *Client) HandleCommandAction(action KeymapAction) error {
	switch action {
	case KeymapActionEnter:
		m := c.m.(*CommandMode)
		switch m.NextMode.(type) {
		case *FilterMode:
			nm := m.NextMode.(*FilterMode)
			nm.SetFilter(m.Input)
			c.m = nm
			c.SetMode(ModeFilter)
			c.DrawCache()
		}
	case KeymapActionQuit:
		c.DrawCacheNormal()
	case KeymapActionInput:
		m := c.m.(*CommandMode)
		m.SetInput(m.Input + string(c.eventKey.Rune()))
		c.DrawCommand()
	case KeymapActionBackspace:
		m := c.m.(*CommandMode)
		if m.Input == "" {
			return nil
		}
		m.SetInput(tools.StringBackspace(m.Input))
		c.DrawCommand()
	}
	return nil
}

func (c *Client) HandleFilterAction(action KeymapAction) error {
	switch action {
	case KeymapActionQuit:
		c.DrawCacheNormal()
	case KeymapActionMoveUp:
		c.midTerm.SetAnchorIndex(5)
		c.MoveUp(1)
	case KeymapActionMoveDown:
		c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() - 5)
		c.MoveDown(1)
	}
	return nil
}

// 操作复制信息
func (c *Client) ActionCopyMsg(msg string) error {
	err := clipboard.WriteAll(msg)
	if err != nil {
		return err
	}
	msg = fmt.Sprintf("%s 已经复制到剪切板", msg)
	c.DrawCacheNormal()
	c.DrawMessage(msg)
	return nil
}

func (c *Client) GetAction() (KeymapAction, bool) {
	var actionMap map[string]KeymapAction
	switch c.GetMode() {
	case ModeNormal:
		if IsKeymap(c.eventKey.Rune()) {
			return KeymapActionKeymap, true
		} else {
			actionMap = ActionNormalMap
		}
	case ModeConfirm:
		actionMap = ActionConfirmMap
	case ModeSync:
		actionMap = ActionSyncMap
	case ModeFilter:
		actionMap = ActionFilterMap
	case ModeKeymap:
		key := string(c.prevRune) + string(c.eventKey.Rune())
		a, ok := ActionKeymapMap[key]
		return a, ok
	case ModeCommand:
		switch c.eventKey.Key() {
		case tcell.KeyEnter:
			return KeymapActionEnter, true
		case tcell.KeyEsc, tcell.KeyCtrlC:
			return KeymapActionQuit, true
		case tcell.KeyBackspace2, tcell.KeyBackspace:
			return KeymapActionBackspace, true
		}
		return KeymapActionInput, true
	default:
		return 0, false
	}
	return GetKeymapActionByEventKey(c.eventKey, actionMap)
}

func (c *Client) Exec() error {
	var err error
	defer c.t.Quit()
	c.Draw()
	for {
		c.t.S.Show()
		ev := c.t.S.PollEvent()
		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			Log.Infof("PollEvent EventResize %v", ev)
			c.DrawCache()
		case *tcell.EventKey:
			c.eventKey = ev
			Log.Infof("PollEvent Mode %v EventKey %v", c.GetMode(), ev)
			action, ok := c.GetAction()
			var actionFunc func(KeymapAction) error
			switch c.GetMode() {
			case ModeNormal:
				actionFunc = c.HandleNormalAction
			case ModeKeymap:
				actionFunc = c.HandleKeymapAction
			case ModeSync:
				actionFunc = c.HandleSyncAction
			default:
				if c.m != nil {
					actionFunc = c.m.GetActionFn()
				}
			}
			if actionFunc != nil && ok {
				err = actionFunc(action)
			} else {
				c.DrawCacheNormal()
			}
		}
		if err != nil {
			if CanCacheError(err) {
				c.DrawCacheNormal()
				c.DrawMessage(err.Error())
			} else {
				return err
			}
		}
	}
}
