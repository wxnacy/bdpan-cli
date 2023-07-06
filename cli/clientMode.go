package cli

import (
	"fmt"
	"path/filepath"

	"github.com/wxnacy/bdpan"
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
	m := NewConfirmMode(c.t, ensureC, msg)
	m.SetKeymapFn(c.HandleConfirmKeymap).SetKeymaps(ConfirmKeymaps)
	c.m = m
	return c.SetMode(ModeConfirm)
}

func (c *Client) HandleConfirmKeymap(k Keymap) error {
	var err error
	ensureFunc := func() error {
		switch c.m.(*ConfirmMode).EnsureCommand {
		case CommandDeleteFile:
			c.DrawMessage("开始删除...")
			err = bdpan.DeleteFile(c.GetMidSelectFile().Path)
			c.ClearSelectFiles()
			if err != nil {
				c.DrawMessage(fmt.Sprintf("删除失败: %v", err))
				c.DrawCacheNormal()
			} else {
				c.DrawNormal()
				c.DrawMessage("删除成功!")
			}
		case CommandDownloadFile:
			c.DrawMessage("开始下载...")
			cmd := &DownloadCommand{
				IsRecursion: true,
			}
			err = cmd.Download(c.selectFiles[0])
			if err != nil {
				c.DrawMessage(fmt.Sprintf("下载失败: %v", err))
			} else {
				c.DrawCacheNormal()
				c.DrawMessage("下载成功!")
			}
			bdpan.SetOutputFile()
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
		}
		return nil
	}

	term := c.m.(*ConfirmMode).Term
	switch k.Command {
	case CommandCursorMoveLeft:
		term.EnableEnsure().Draw()
	case CommandCursorMoveRight:
		term.EnableEnsure().Draw()
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
	m := NewKeymapMode(c.t, c.eventKey.Rune())
	// m.SetActionFn(c.HandleKeymapAction).SetKeymapActionMap(ActionKeymapMap)
	m.SetKeymapFn(c.HandleKeymapKeymap).
		SetKeymaps(KeymapKeymaps).
		SetEventKey(c.eventKey)
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
	case CommandCopyFilepath:
		return c.ActionCopyMsg(c.GetMidSelectFile().Path)
	case CommandCopyFilename:
		return c.ActionCopyMsg(c.GetMidSelectFile().GetFilename())
	case CommandCopyDirpath:
		return c.ActionCopyMsg(filepath.Dir(c.GetMidSelectFile().Path))
	// case CommandCopyFile:
	// c.SetCurrSelectFiles().SetPrevAction(action).DrawCacheNormal()
	// fromFile := c.selectFiles[0]
	// c.DrawMessage(fmt.Sprintf("%s 已经复制", fromFile.Path))
	case CommandPasteFile:
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
	default:
		c.HandleCommonKeymap(k)
	}
	return err
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
		case *FilterMode:
			nm := m.NextMode.(*FilterMode)
			nm.SetFilter(m.Input)
			c.m = nm
			c.SetMode(ModeFilter)
			c.DrawCache()
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
func (c *Client) NewFilterMode() *FilterMode {
	fm := NewFilterMode("")
	// fm.SetActionFn(c.HandleFilterAction).SetKeymapActionMap(ActionFilterMap)
	fm.SetKeymapFn(c.HandleFilterKeymap).SetKeymaps(FilterKeymaps)
	return fm
}

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

func (c *Client) HandleFilterKeymap(k Keymap) error {
	switch k.Command {
	case CommandQuit:
		c.DrawCacheNormal()
	case CommandCursorMoveUp:
		// c.midTerm.SetAnchorIndex(5)
		c.MoveUp(1)
	case CommandCursorMoveDown:
		// c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() - 5)
		c.MoveDown(1)
	}
	return nil
}

//---------------------------
// NormalMode
//---------------------------
func (c *Client) SetNormalMode() *Client {
	m := NewNormalMode()
	m.SetKeymapFn(c.HandleNormalKeymap).
		SetKeymaps(NormalKeymaps)
		// SetEventKey(c.eventKey)
	c.m = m
	// c.normalAction = a
	return c.SetMode(ModeNormal)
}

func (c *Client) HandleNormalKeymap(k Keymap) error {
	var err error
	switch k.Command {
	// case KeymapActionHelp:
	// c.SetMode(ModeHelp).DrawCache()
	// case KeymapActionFilter:
	// c.SetFilterMode().DrawCache()
	// case KeymapActionSync:
	// c.SetMode(ModeSync).DrawCache()
	// case KeymapActionReload:
	// c.DrawNormal()
	// // 向下移动
	// case KeymapActionMoveDown:
	// c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() - 5)
	// c.MoveDown(1)
	// case KeymapActionMoveDownHalfPage:
	// c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() / 2)
	// c.MoveDown(c.midTerm.Box.Height() / 2)
	// case KeymapActionMoveDownPage:
	// c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() - 1)
	// c.MoveDown(c.midTerm.Box.Height())
	// case KeymapActionMovePageEnd:
	// c.MoveDown(c.midTerm.Length())
	// // 向上移动
	// case KeymapActionMoveUp:
	// c.midTerm.SetAnchorIndex(5)
	// c.MoveUp(1)
	// case KeymapActionMoveUpHalfPage:
	// c.midTerm.SetAnchorIndex(c.midTerm.Box.Height() / 2)
	// c.MoveUp(c.midTerm.Box.Height() / 2)
	// case KeymapActionMoveUpPage:
	// c.midTerm.SetAnchorIndex(0)
	// c.MoveUp(c.midTerm.Box.Height())
	// case KeymapActionMoveLeft:
	// c.MoveLeft()
	// case KeymapActionMoveLeftHome:
	// c.midFile = GetRootFile()
	// c.DrawCache()
	// case KeymapActionEnter, KeymapActionMoveRight:
	// err = c.Enter()
	// case KeymapActionCutFile:
	// c.SetCurrSelectFiles().SetPrevAction(action).DrawCacheNormal()
	// fromFile := c.selectFiles[0]
	// c.DrawMessage(fmt.Sprintf("%s 已经剪切", fromFile.Path))
	// case KeymapActionDownloadFile:
	// c.Download()
	// case KeymapActionDeleteFile:
	// var msg string
	// var name = c.midTerm.GetSeleteItem().Info.Name()
	// msg = fmt.Sprintf("确定删除 %s?", name)
	// c.SetConfirmMode(msg).SetCurrSelectFiles().SetPrevAction(action).DrawCache()
	case CommandKeymap:
		return c.SetKeymapMode().DrawCache()
	// case KeymapActionSystem:
	// c.ShowSystem()
	case CommandQuit:
		return ErrQuit
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
	case CommandQuit:
		c.DrawCacheNormal()
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
