package cli

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

var (
	Keymaps = []Keymap{
		NewKeymap('y', "").
			AddRelKey(NewKeymap('d', "复制所在文件夹名称")).
			AddRelKey(NewKeymap('n', "复制文件名称")).
			AddRelKey(NewKeymap('p', "复制文件路径")).
			AddRelKey(NewKeymap('y', "复制文件")),
		NewKeymap('p', "").
			AddRelKey(NewKeymap('p', "粘贴文件")),
		// NewKeymap('s', "").
		// AddRelKey(NewKeymap('e', "执行同步")),
		NewKeymap('g', "").
			AddRelKey(NewKeymap('g', "跳转页面首行")),
	}

	ActionNormalMap = map[string]KeymapAction{
		// 帮助
		"?": KeymapActionHelp,
		// 搜索
		"/": KeymapActionFilter,
		// 同步页面
		"s": KeymapActionSync,
		// 光标操作
		"j": KeymapActionMoveDown,
		"k": KeymapActionMoveUp,
		"h": KeymapActionMoveLeft,
		"H": KeymapActionMoveLeftHome,
		"l": KeymapActionEnter,
		"G": KeymapActionMovePageEnd,

		tcell.KeyNames[tcell.KeyCtrlD]: KeymapActionMoveDownHalfPage,
		tcell.KeyNames[tcell.KeyCtrlU]: KeymapActionMoveUpHalfPage,
		tcell.KeyNames[tcell.KeyCtrlF]: KeymapActionMoveDownPage,
		tcell.KeyNames[tcell.KeyCtrlB]: KeymapActionMoveUpPage,
		tcell.KeyNames[tcell.KeyUp]:    KeymapActionMoveUp,
		tcell.KeyNames[tcell.KeyDown]:  KeymapActionMoveDown,
		tcell.KeyNames[tcell.KeyLeft]:  KeymapActionMoveLeft,
		tcell.KeyNames[tcell.KeyRight]: KeymapActionEnter,
		// 文件操作
		"x": KeymapActionCutFile,

		"D": KeymapActionDeleteFile,

		"d": KeymapActionDownloadFile,

		"R": KeymapActionReload,

		// 设置
		",": KeymapActionSystem,
		// 退出
		"q":                             KeymapActionQuit,
		tcell.KeyNames[tcell.KeyEscape]: KeymapActionQuit,
		tcell.KeyNames[tcell.KeyCtrlC]:  KeymapActionQuit,
	}

	ActionConfirmMap = map[string]KeymapAction{
		// 光标操作
		"h":                            KeymapActionMoveLeft,
		"l":                            KeymapActionMoveRight,
		"y":                            KeymapActionEnsure,
		tcell.KeyNames[tcell.KeyLeft]:  KeymapActionMoveLeft,
		tcell.KeyNames[tcell.KeyRight]: KeymapActionMoveRight,
		tcell.KeyNames[tcell.KeyEnter]: KeymapActionEnter,
	}

	ActionKeymapMap = map[string]KeymapAction{
		"gg": KeymapActionMovePageHome,

		"yp": KeymapActionCopyPath,
		"yn": KeymapActionCopyName,
		"yd": KeymapActionCopyDir,
		"yy": KeymapActionCopyFile,

		"pp": KeymapActionPasteFile,

		// 同步操作
		// "se": KeymapActionSyncExec,
	}

	ActionSyncMap = map[string]KeymapAction{
		"e": KeymapActionSyncExec,
		// 光标操作
		"j": KeymapActionMoveDown,
		"k": KeymapActionMoveUp,
	}

	ActionFilterMap = map[string]KeymapAction{
		"j": KeymapActionMoveDown,
		"k": KeymapActionMoveUp,
		// 退出
		"q":                             KeymapActionQuit,
		tcell.KeyNames[tcell.KeyEscape]: KeymapActionQuit,
		tcell.KeyNames[tcell.KeyCtrlC]:  KeymapActionQuit,
	}

	KeyActionMap = map[string]KeymapAction{
		// 帮助
		// "?": KeymapActionHelp,
		// // 光标操作
		// "j":  KeymapActionMoveDown,
		// "k":  KeymapActionMoveUp,
		// "h":  KeymapActionMoveLeft,
		// "l":  KeymapActionMoveRight,
		// "gg": KeymapActionMovePageHome,
		// "G":  KeymapActionMovePageEnd,

		// tcell.KeyNames[tcell.KeyCtrlD]: KeymapActionMoveDownHalfPage,
		// tcell.KeyNames[tcell.KeyCtrlU]: KeymapActionMoveUpHalfPage,
		// tcell.KeyNames[tcell.KeyCtrlF]: KeymapActionMoveDownPage,
		// tcell.KeyNames[tcell.KeyCtrlB]: KeymapActionMoveUpPage,
		// tcell.KeyNames[tcell.KeyUp]:    KeymapActionMoveUp,
		// tcell.KeyNames[tcell.KeyDown]:  KeymapActionMoveDown,
		// tcell.KeyNames[tcell.KeyLeft]:  KeymapActionMoveLeft,
		// tcell.KeyNames[tcell.KeyRight]: KeymapActionMoveRight,
		// tcell.KeyNames[tcell.KeyEnter]: KeymapActionEnter,
		// // 文件操作
		// "x": KeymapActionCutFile,

		// "D": KeymapActionDeleteFile,

		// "d": KeymapActionDownloadFile,

		// "yp": KeymapActionCopyPath,
		// "yn": KeymapActionCopyName,
		// "yd": KeymapActionCopyDir,
		// "yy": KeymapActionCopyFile,

		// "pp": KeymapActionPasteFile,

		// // 同步操作
		// "se": KeymapActionSyncExec,
	}
)

type KeymapAction int

const (
	KeymapActionCopyPath KeymapAction = iota + 1
	KeymapActionCopyName
	KeymapActionCopyDir
	KeymapActionCopyFile

	KeymapActionCutFile

	KeymapActionDeleteFile

	KeymapActionPasteFile

	KeymapActionDownloadFile

	KeymapActionSyncExec // 执行同步

	KeymapActionMoveDown
	KeymapActionMoveUp
	KeymapActionMoveLeft
	KeymapActionMoveLeftHome
	KeymapActionMoveRight
	KeymapActionMovePageHome
	KeymapActionMovePageEnd
	KeymapActionMoveDownHalfPage
	KeymapActionMoveUpHalfPage
	KeymapActionMoveDownPage
	KeymapActionMoveUpPage

	KeymapActionEnter  //回车
	KeymapActionEnsure //确认

	// 进入模式
	KeymapActionHelp
	KeymapActionKeymap
	KeymapActionNormal
	KeymapActionSync
	KeymapActionFilter

	KeymapActionReload

	KeymapActionInput
	KeymapActionBackspace

	KeymapActionSystem
	KeymapActionQuit
)

func GetKeymapActionByEventKey(ev *tcell.EventKey, actionMap map[string]KeymapAction) (a KeymapAction, ok bool) {
	key := string(ev.Rune())
	a, ok = actionMap[key]
	if ok {
		Log.Infof("GetKeymapActionByKey EventKey %v Key %s Action %v", ev, key, a)
		return
	}
	key = tcell.KeyNames[ev.Key()]
	a, ok = actionMap[key]
	if ok {
		Log.Infof("GetKeymapActionByKey EventKey %v Key %s Action %v", ev, key, a)
		return
	}
	return
}

func IsKeymap(r rune) bool {
	for _, k := range Keymaps {
		if k.R == r {
			return true
		}
	}
	return false
}

func GetKeymapRunes() []rune {
	var runes = make([]rune, 0)
	for _, k := range Keymaps {
		if k.R != 0 {
			runes = append(runes, k.R)
		}
	}
	return runes
}

func GetRelKeysByRune(r rune) []Keymap {
	for _, k := range Keymaps {
		if k.R == r {
			return k.RelKeys
		}
	}
	return nil
}

func GetRelKeysMsgByRune(r rune) []string {
	var msgs = make([]string, 0)
	relkeys := GetRelKeysByRune(r)
	if relkeys != nil {
		for _, k := range relkeys {
			var msg = fmt.Sprintf("  %s\t%s", string(k.R), k.Msg)
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

func NewKeymap(r rune, doc string) Keymap {
	return Keymap{R: r, Msg: doc}
}

type Keymap struct {
	R       rune
	Key     tcell.Key
	Msg     string
	RelKeys []Keymap
}

func (k Keymap) AddRelKey(km Keymap) Keymap {
	k.RelKeys = append(k.RelKeys, km)
	return k
}
