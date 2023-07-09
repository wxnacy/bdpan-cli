package cli

import (
	"fmt"
	"path/filepath"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
	"github.com/wxnacy/go-tools"
)

func (c *Client) HandleCommonKeymap(k Keymap) error {
	switch k.Command {
	case CommandHelp:
		c.SetHelpMode().DrawCache()
	case CommandQuit:
		c.DrawCacheNormal()
	}
	return nil
}

//---------------------------
// ConfirmMode
//---------------------------
func (c *Client) SetConfirmMode(ensureC Command, msg string) *Client {
	// 设置缓存
	SetCacheSelectIndex(c.normalAction, c.midTerm.SelectIndex)
	m := NewConfirmMode(c.t, ensureC, msg)
	m.SetKeymapFn(c.HandleConfirmKeymap).SetKeymaps(ConfirmKeymaps)
	c.m = m
	// 设置选中条目
	c.SetSelectItems()
	return c.SetMode(ModeConfirm)
}

func (c *Client) HandleConfirmKeymap(k Keymap) error {
	var err error
	ensureFunc := func() error {
		switch c.m.(*ConfirmMode).EnsureCommand {
		case CommandDownloadFile:
			c.DrawMessage("开始下载...")
			cmd := &DownloadCommand{
				IsRecursion: true,
			}
			for _, item := range c.m.GetSelectItems() {
				file := item.Info.(*FileInfo).FileInfoDto
				c.DrawMessage("开始下载 " + file.Path)
				err = cmd.Download(file)
			}
			c.DrawCacheNormal()
			if err != nil {
				c.DrawMessage(fmt.Sprintf("下载失败: %v", err))
			} else {
				c.DrawMessage("下载成功!")
			}
			bdpan.SetOutputFile()
		case CommandDelete:
			c.DrawMessage("开始删除...")
			switch c.normalAction {
			case ActionSync:
				for _, sm := range c.m.GetSelectItems() {
					id := sm.Info.(*SyncInfo).ID
					err = bdpan.DeleteSyncModel(id)
					if err != nil {
						return err
					}
				}
				c.DrawCacheNormal()
				c.DrawMessage("删除成功!")
			default:
				var paths []string
				for _, item := range c.m.GetSelectItems() {
					paths = append(paths, item.Info.(*FileInfo).Path)
				}
				err = bdpan.DeleteFiles(paths)
				if err != nil {
					c.DrawMessage(fmt.Sprintf("删除失败: %v", err))
					c.DrawCacheNormal()
				} else {
					c.DrawNormal()
					c.DrawMessage("删除成功!")
					return c.RefreshUsed()
				}

			}
		case CommandSyncExec:
			c.DrawMessage("开始同步...")
			info := c.midTerm.GetSeleteItem().Info.(*SyncInfo)
			err := info.Exec()
			if err != nil {
				Log.Errorf("SyncModel %s Exec Error: %v", info.ID, err)
				return ErrActionFail
			}
			c.DrawCacheNormal()
			c.DrawMessage(fmt.Sprintf("%s 同步成功", info.Remote))
			return c.RefreshUsed()
		}
		return nil
	}

	term := c.m.(*ConfirmMode).Term
	switch k.Command {
	case CommandCursorMoveLeft:
		term.EnableEnsure().Draw()
	case CommandCursorMoveRight:
		term.EnableCancel().Draw()
	case CommandEnter:
		if term.IsEnsure() {
			err = ensureFunc()
		} else {
			c.DrawCacheNormal()
			c.DrawMessage("操作取消!")
		}
	case CommandEnsure:
		term.EnableEnsure().Draw()
		err = ensureFunc()
	case CommandQuit:
		c.DrawCacheNormal()
	default:
		c.HandleCommonKeymap(k)
	}
	if err != nil {
		return err
	}
	return nil
}

//---------------------------
// KeymapMode
//---------------------------
func (c *Client) SetKeymapMode() *Client {
	m := NewKeymapMode(c.t, c.eventKey)
	m.SetKeymapFn(c.HandleKeymapKeymap).
		SetKeymaps(KeymapKeymaps)
	m.SetSelectItems(c.m.GetSelectItems())
	m.SetPrevCommand(c.m.GetPrevCommand())
	c.m = m
	return c.SetMode(ModeKeymap)
}

func (c *Client) HandleKeymapKeymap(k Keymap) error {
	var err error
	Log.Infof("HandleKeymap %v", k)
	switch k.Command {
	case CommandCursorMoveHome:
		c.DrawCacheNormal()
		c.MoveUp(c.midTerm.Length())
	case CommandGotoRoot:
		c.SetMidFile(GetRootFile())
		c.DrawCacheNormal()
	case CommandCopyFilepath:
		return c.ActionCopyMsg(c.GetMidSelectFile().Path)
	case CommandCopyFilename:
		return c.ActionCopyMsg(c.GetMidSelectFile().GetFilename())
	case CommandCopyDirpath:
		return c.ActionCopyMsg(filepath.Dir(c.GetMidSelectFile().Path))
	case CommandCopyFile:
		Log.Info(c.m.GetSelectItems())
		c.DrawCacheNormal()
		c.m.SetPrevCommand(k.Command)
		c.m.SetSelectItems([]*terminal.SelectItem{c.midTerm.GetSeleteItem()})
		fromFile := c.GetMidSelectFile()
		c.DrawMessage(fmt.Sprintf("%s 已经复制", fromFile.Path))
	case CommandPasteFile:
		if len(c.m.GetSelectItems()) == 0 {
			return ErrNoFileSelect
		}
		dir := filepath.Dir(c.GetMidSelectFile().Path)
		fromFile := c.GetSelectFile()
		toFile := filepath.Join(dir, fromFile.GetFilename())
		switch c.m.GetPrevCommand() {
		case CommandCutFile:
			err = bdpan.MoveFile(fromFile.Path, toFile)
		case CommandCopyFile:
			err = bdpan.CopyFile(fromFile.Path, toFile)
		default:
			return ErrNoTypeToPaste
		}
		if err != nil {
			return err
		}
		c.m.ClearPrevCommand()
		c.m.ClearSelectItems()
		c.DrawNormal()
		c.DrawMessage(fmt.Sprintf("%s 已经粘贴", toFile))
		return c.RefreshUsed()
	default:
		c.HandleCommonKeymap(k)
	}
	return err
}

func (c *Client) IsKeymapK(k Keymap) bool {
	if len(k.Keys) > 1 {
		return true
	}
	return false
}

//---------------------------
// CommandMode
//---------------------------
func (c *Client) SetCommandMode(nextMode ModeInterface) *Client {
	m := NewCommandMode("/").SetNextMode(nextMode)
	m.SetKeymapFn(c.HandleCommandKeymap).SetKeymaps(CommandKeymaps)
	c.m = m
	c.SetMode(ModeCommand)
	return c
}

func (c *Client) HandleCommandKeymap(k Keymap) error {
	switch k.Command {
	case CommandEnter:
		m := c.m.(*CommandMode)
		switch m.NextMode.(type) {
		// case *FilterMode:
		// nm := m.NextMode.(*FilterMode)
		// nm.SetFilter(m.Input)
		// c.m = nm
		// c.SetMode(ModeFilter)
		// c.DrawCache()
		case *NormalMode:
			nm := m.NextMode.(*NormalMode)
			if c.useFilter {
				c.filterText = m.Input
			}
			c.SetM(nm).DrawCache()
		}
	case CommandQuit:
		c.DrawCacheNormal()
	case CommandInput:
		m := c.m.(*CommandMode)
		m.SetInput(m.Input + string(c.eventKey.Rune()))
		c.DrawCommand()
	case CommandBackspace:
		m := c.m.(*CommandMode)
		if m.Input == "" {
			return nil
		}
		m.SetInput(tools.StringBackspace(m.Input))
		c.DrawCommand()
	}
	return nil
}

//---------------------------
// FilterMode
//---------------------------
// func (c *Client) NewFilterMode() *FilterMode {
// fm := NewFilterMode("")
// // fm.SetActionFn(c.HandleFilterAction).SetKeymapActionMap(ActionFilterMap)
// fm.SetKeymapFn(c.HandleFilterKeymap).SetKeymaps(FilterKeymaps)
// return fm
// }

// func (c *Client) SetFilterMode() *Client {
// fm := NewFilterMode("")
// // fm.SetActionFn(c.HandleFilterAction).SetKeymapActionMap(ActionFilterMap)
// fm.SetKeymapFn(c.HandleFilterKeymap).SetKeymaps(FilterKeymaps)
// m := NewCommandMode("/").SetNextMode(fm)
// m.SetActionFn(c.HandleCommandAction)
// c.m = m
// c.SetMode(ModeCommand)
// return c
// }

// func (c *Client) HandleFilterKeymap(k Keymap) error {
// switch k.Command {
// case CommandQuit:
// c.DrawCacheNormal()
// case CommandCursorMoveUp:
// // c.midTerm.SetAnchorIndex(5)
// c.MoveUp(1)
// case CommandCursorMoveDown:
// // c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() - 5)
// c.MoveDown(1)
// }
// return nil
// }

//---------------------------
// NormalMode
//---------------------------
func (c *Client) SetNormalMode() *Client {
	m := c.NewNormalMode()
	if len(c.m.GetSelectItems()) > 0 {
		m.SetSelectItems(c.m.GetSelectItems())
	}
	if c.m.GetPrevCommand() != "" {
		m.SetPrevCommand(c.m.GetPrevCommand())
	}
	c.m = m
	return c.SetMode(ModeNormal)
}

func (c *Client) NewNormalMode() *NormalMode {
	m := NewNormalMode()
	m.SetKeymapFn(c.HandleNormalKeymap).
		SetKeymaps(NormalKeymaps)
	return m
}

func (c *Client) HandleNormalKeymap(k Keymap) error {
	var err error
	switch k.Command {
	case CommandHelp:
		c.SetHelpMode().DrawCache()
	case CommandFilter:
		c.useFilter = true
		c.SetCommandMode(c.m).DrawCache()
	case CommandSync:
		c.SetSyncMode().DrawCache()
	case CommandReload:
		c.DrawNormal()
	// 向下移动
	case CommandCursorMoveDown:
		c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() - 5)
		c.MoveDown(1)
	case CommandCursorMoveHalfPageDown:
		c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() / 2)
		c.MoveDown(c.midTerm.Box.Height() / 2)
	case CommandCursorMovePageDown:
		c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() - 1)
		c.MoveDown(c.midTerm.Box.Height())
	case CommandCursorMoveEnd:
		c.MoveDown(c.midTerm.Length())
		// 向上移动
	case CommandCursorMoveUp:
		c.midTerm.SetAnchorIndex(5)
		c.MoveUp(1)
	case CommandCursorMoveHalfPageUp:
		c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() / 2)
		c.MoveUp(c.midTerm.Box.Height() / 2)
	case CommandCursorMovePageUp:
		c.midTerm.SetAnchorIndex(0)
		c.MoveUp(c.midTerm.Box.Height())
	case CommandGotoPrev:
		c.MoveLeft()
	case CommandEnter, CommandGotoNext:
		err = c.Enter()
	case CommandCutFile:
		c.m.SetPrevCommand(k.Command)
		c.m.SetSelectItems([]*terminal.SelectItem{c.midTerm.GetSeleteItem()})
		fromFile := c.GetMidSelectFile()
		c.DrawMessage(fmt.Sprintf("%s 已经剪切", fromFile.Path))
	case CommandDownloadFile:
		c.Download()
	case CommandDelete:
		var msg string
		var name = c.midTerm.GetSeleteItem().Info.Name()
		msg = fmt.Sprintf("确定删除 %s?", name)
		c.SetConfirmMode(k.Command, msg).DrawConfirm()
	case CommandKeymap:
		return c.SetKeymapMode().DrawCache()
	case CommandSystem:
		c.ShowSystem()
	case CommandSelect:
		item := c.midTerm.GetSeleteItem()
		item.IsSelect = !item.IsSelect
		c.midTerm.Draw()
		c.MoveDown(1)
	case CommandQuit:
		if c.useFilter {
			c.DisableFilter().DrawCache()
		} else {
			return ErrQuit
		}
	default:
		if c.IsKeymapK(k) {
			c.SetKeymapMode().DrawCache()
		}
	}
	return err
}

//---------------------------
// SyncMode
//---------------------------
func (c *Client) SetSyncMode() *Client {
	sx, sy, ex, ey := c.GetModeDrawRange()
	m := NewSyncMode(c.t, c.midTerm.GetSeleteItem(), sx, sy, ex, ey)
	m.SetKeymapFn(c.HandleSyncKeymap).SetKeymaps(SyncKeymaps)
	if m.Count() == 0 {
		return c
	}
	c.m = m
	return c.SetMode(ModeSync)
}

func (c *Client) HandleSyncKeymap(k Keymap) error {
	Term := c.m.(*SyncMode).Term
	switch k.Command {
	case CommandSyncExec:
		info := Term.GetSeleteItem().Info.(*SyncInfo)
		err := info.Exec()
		if err != nil {
			Log.Errorf("SyncModel %s Exec Error: %v", info.ID, err)
			return ErrActionFail
		}
		c.DrawCacheNormal()
		c.DrawMessage(fmt.Sprintf("%s 同步成功", c.GetMidSelectFile().Path))
	case CommandCursorMoveDown:
		if Term.MoveDownSelect(1) {
			Term.Draw()
		}
	case CommandCursorMoveUp:
		if Term.MoveUpSelect(1) {
			Term.Draw()
		}
	case CommandEnter:
		c.SetNormalMode().SetNormalAction(ActionSync).DrawCache()
	default:
		return c.HandleCommonKeymap(k)
	}
	return nil
}

//---------------------------
// HelpMode
//---------------------------
func (c *Client) SetHelpMode() *Client {
	sx, sy, ex, ey := c.GetModeDrawRange()
	m := NewHelpMode(c.t, c.GetMode(), sx, sy, ex, ey)
	m.SetKeymapFn(c.HandleHelpKeymap).SetKeymaps(HelpKeymaps)
	c.m = m
	return c.SetMode(ModeHelp)
}

func (c *Client) HandleHelpKeymap(k Keymap) error {
	switch k.Command {
	case CommandQuit:
		c.DrawCacheNormal()
	}
	return nil
}
