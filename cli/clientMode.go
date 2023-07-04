package cli

import (
	"fmt"

	"github.com/wxnacy/bdpan"
)

//---------------------------
// ConfirmMode
//---------------------------

func (c *Client) SetConfirmMode(msg string) *Client {
	m := NewConfirmMode(c.t, msg)
	m.SetActionFn(c.HandleConfirmAction)
	c.m = m
	return c.SetMode(ModeConfirm)
}

func (c *Client) HandleConfirmAction(action KeymapAction) error {
	var err error
	Log.Infof("HandleKeymapAction %v", action)
	ensureFunc := func() error {
		var action = c.prevAction
		switch action {
		case KeymapActionDeleteFile:
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
		case KeymapActionDownloadFile:
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
		case KeymapActionSyncExec:
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
	switch action {
	case KeymapActionMoveLeft:
		term.EnableEnsure().Draw()
	case KeymapActionMoveRight:
		term.EnableEnsure().Draw()
	case KeymapActionEnter:
		if term.IsEnsure() {
			err = ensureFunc()
		} else {
			c.DrawCacheNormal()
			c.DrawMessage("操作取消!")
		}
	case KeymapActionEnsure:
		term.EnableEnsure().Draw()
		err = ensureFunc()
	}
	if err != nil {
		return err
	}
	return nil
}
