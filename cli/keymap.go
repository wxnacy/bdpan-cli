package cli

import (
	"fmt"
	"strings"

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
		tcell.KeyNames[tcell.KeyEnter]: KeymapActionEnter,
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

	KeymapCursorMoveDown1 = NewKeymapN([]string{"j"}, CommandCursorMoveDown).
				SetDesc("光标向下移动一行")
	KeymapCursorMoveDown2 = NewKeymapN([]string{tcell.KeyNames[tcell.KeyDown]}, CommandCursorMoveDown).
				SetDesc("光标向下移动一行")
	KeymapCursorMoveUp1 = NewKeymapN([]string{"k"}, CommandCursorMoveUp).
				SetDesc("光标向上移动一行")
	KeymapCursorMoveUp2 = NewKeymapN([]string{tcell.KeyNames[tcell.KeyUp]}, CommandCursorMoveUp).
				SetDesc("光标向上移动一行")

	KeymapEnter = NewKeymapN([]string{tcell.KeyNames[tcell.KeyEnter]}, CommandEnter).
			SetDesc("确定")

	KeymapInput = NewKeymapN([]string{""}, CommandInput).
			SetDesc("输入")

	KeymapBackspace1 = NewKeymapN([]string{tcell.KeyNames[tcell.KeyBackspace]}, CommandBackspace).
				SetDesc("退格")
	KeymapBackspace2 = NewKeymapN([]string{tcell.KeyNames[tcell.KeyBackspace2]}, CommandBackspace).
				SetDesc("退格")

	KeymapQuit1 = NewKeymapN([]string{tcell.KeyNames[tcell.KeyEscape]}, CommandQuit).
			SetDesc("退出")
	KeymapQuit2 = NewKeymapN([]string{tcell.KeyNames[tcell.KeyCtrlC]}, CommandQuit).
			SetDesc("退出")
	KeymapQuit3 = NewKeymapN([]string{"q"}, CommandQuit).
			SetDesc("退出")
	KeymapHelp = NewKeymapN([]string{"?"}, CommandHelp).
			SetDesc("帮助")

	NormalKeymaps = []Keymap{
		// 模式
		NewKeymapN([]string{"?"}, CommandHelp).SetDesc("帮助"),
		NewKeymapN([]string{"/"}, CommandFilter).SetDesc("过滤"),
		NewKeymapN([]string{"s"}, CommandSync).SetDesc("同步"),

		NewKeymapN([]string{","}, CommandSystem).SetDesc("文件源"),
		// 光标移动
		KeymapCursorMoveDown1,
		KeymapCursorMoveDown2,
		KeymapCursorMoveUp1,
		KeymapCursorMoveUp2,
		NewKeymapN([]string{"h"}, CommandGotoPrev).
			SetDesc("返回上一层目录"),
		NewKeymapN([]string{"G"}, CommandCursorMoveEnd).
			SetDesc("光标跳转到最后一行"),

		NewKeymapN(
			[]string{tcell.KeyNames[tcell.KeyCtrlD]},
			CommandCursorMoveHalfPageDown,
		).SetDesc("光标向下移动半页"),
		NewKeymapN(
			[]string{tcell.KeyNames[tcell.KeyCtrlU]},
			CommandCursorMoveHalfPageUp,
		).SetDesc("光标向上移动半页"),
		NewKeymapN(
			[]string{tcell.KeyNames[tcell.KeyCtrlF]},
			CommandCursorMovePageDown,
		).SetDesc("光标向下移动一页"),
		NewKeymapN(
			[]string{tcell.KeyNames[tcell.KeyCtrlB]},
			CommandCursorMovePageUp,
		).SetDesc("光标向上移动一页"),

		// 操作文件
		NewKeymapN([]string{"l"}, CommandEnter).
			SetDesc("选中文件，进入下一层目录或下载文件"),
		NewKeymapN([]string{tcell.KeyNames[tcell.KeyRight]}, CommandEnter).
			SetDesc("选中文件，进入下一层目录或下载文件"),
		NewKeymapN([]string{tcell.KeyNames[tcell.KeyEnter]}, CommandEnter).
			SetDesc("选中文件，进入下一层目录或下载文件"),
		NewKeymapN([]string{"x"}, CommandCutFile).
			SetDesc("剪切文件"),
		NewKeymapN([]string{"D"}, CommandDeleteFile).
			SetDesc("删除文件"),
		NewKeymapN([]string{"d"}, CommandDownloadFile).
			SetDesc("下载文件"),
		NewKeymapN([]string{"R"}, CommandReload).
			SetDesc("重新加载目录"),
	}

	FilterKeymaps = []Keymap{
		KeymapCursorMoveDown1,
		KeymapCursorMoveDown2,
		KeymapCursorMoveUp1,
		KeymapCursorMoveUp2,
		KeymapQuit1,
		KeymapQuit2,
		KeymapQuit3,
	}

	CommandKeymaps = []Keymap{
		KeymapEnter,
		KeymapBackspace1,
		KeymapBackspace2,
		KeymapQuit1,
		KeymapQuit2,
		KeymapQuit3,
	}

	SyncKeymaps = []Keymap{
		NewKeymapN([]string{"e"}, CommandSyncExec).
			SetDesc("执行同步"),
		KeymapEnter,
		KeymapCursorMoveDown1,
		KeymapCursorMoveDown2,
		KeymapCursorMoveUp1,
		KeymapCursorMoveUp2,
	}

	ConfirmKeymaps = []Keymap{
		NewKeymapN([]string{"y"}, CommandEnsure).
			SetDesc("确认"),
		NewKeymapN([]string{"h"}, CommandCursorMoveLeft).
			SetDesc("光标向左移动一个"),
		NewKeymapN([]string{"l"}, CommandCursorMoveRight).
			SetDesc("光标向右移动一个"),
	}

	HelpKeymaps = []Keymap{}

	CommonKeymaps = []Keymap{
		KeymapQuit1,
		KeymapQuit2,
		KeymapQuit3,
		KeymapHelp,
	}

	KeymapKeymaps = []Keymap{
		KeymapHelp,
		NewKeymapN([]string{"y", "d"}, CommandCopyDirpath).
			SetDesc("复制文件夹路径"),
		NewKeymapN([]string{"y", "n"}, CommandCopyFilename).
			SetDesc("复制文件名"),
		NewKeymapN([]string{"y", "p"}, CommandCopyFilepath).
			SetDesc("复制文件路径"),
		NewKeymapN([]string{"y", "y"}, CommandCopyFile).
			SetDesc("复制文件"),

		NewKeymapN([]string{"g", "g"}, CommandCursorMoveHome).
			SetDesc("移动光标到最上层"),
		NewKeymapN([]string{"g", "r"}, CommandGotoRoot).
			SetDesc("进入根目录"),

		NewKeymapN([]string{"p", "p"}, CommandPasteFile).SetDesc("粘贴文件"),
	}

	ModeKeymapsMap = map[Mode][]Keymap{
		ModeConfirm: ConfirmKeymaps,
		ModeNormal:  NormalKeymaps,
		ModeSync:    SyncKeymaps,
		ModeKeymap:  KeymapKeymaps,
		ModeFilter:  FilterKeymaps,
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

func NewKeymapN(keys []string, c Command) Keymap {
	return Keymap{Keys: keys, Command: c}
}

type Keymap struct {
	R             rune
	Key           tcell.Key
	Keys          []string
	Command       Command
	CommandString string
	Msg           string
	Desc          string
	RelKeys       []Keymap
}

func (k Keymap) IsNil() bool {
	if len(k.Keys) == 0 {
		return true
	}
	return false
}

func (k Keymap) GetKey() string {
	return strings.Join(k.Keys, "")
}

func (k Keymap) IsKey(currEK *tcell.EventKey, prevEK *tcell.EventKey) bool {
	var key string
	if len(k.Keys) == 1 {
		key = k.Keys[0]
		if key == string(currEK.Rune()) {
			return true
		}
		if key == tcell.KeyNames[currEK.Key()] {
			return true
		}
	} else if len(k.Keys) > 1 {
		key = strings.Join(k.Keys, "")
		if key == string(prevEK.Rune())+string(currEK.Rune()) {
			return true
		}
	}
	return false
}

func (k Keymap) SetCommandString(c string) Keymap {
	k.CommandString = c
	return k
}

func (k Keymap) SetDesc(d string) Keymap {
	k.Desc = d
	return k
}

func (k Keymap) AddRelKey(km Keymap) Keymap {
	k.RelKeys = append(k.RelKeys, km)
	return k
}

type Command string

const (
	CommandGoto     Command = "goto"
	CommandGotoPrev         = "goto prev"
	CommandGotoNext         = "goto next"
	CommandGotoRoot         = "goto /"

	CommandCursorMoveHome         = "cursor_move_home"
	CommandCursorMoveEnd          = "cursor_move_end"
	CommandCursorMoveUp           = "cursor_move_up"
	CommandCursorMovePageUp       = "cursor_move_page_up"
	CommandCursorMoveHalfPageUp   = "cursor_move_half_page_up"
	CommandCursorMoveDown         = "cursor_move_down"
	CommandCursorMovePageDown     = "cursor_move_page_down"
	CommandCursorMoveHalfPageDown = "cursor_move_half_page_down"
	CommandCursorMoveLeft         = "cursor_move_left"
	CommandCursorMoveRight        = "cursor_move_right"

	CommandCopyDirpath  = "copy_dirpath"
	CommandCopyFilename = "copy_filename"
	CommandCopyFilepath = "copy_filepath"
	CommandCopyFile     = "copy_file"

	CommandCutFile = "cut_file"

	CommandPasteFile = "paste_file"

	CommandDownloadFile = "download_file"

	CommandDeleteFile = "delete_file"

	CommandSyncExec = "sync_exec"

	// 模式切换
	CommandKeymap = "keymap"
	CommandHelp   = "help"
	CommandFilter = "filter"
	CommandSync   = "sync"

	CommandSystem    = "system"
	CommandInput     = "input"
	CommandBackspace = "backspace"
	CommandEnter     = "enter"
	CommandEnsure    = "ensure"
	CommandReload    = "reload"

	CommandQuit = "quit"
)

func IsKeymapK(k Keymap, key *tcell.EventKey) bool {
	if k.Keys[0] == string(key.Rune()) {
		return true
	}
	return false
}
