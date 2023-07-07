package cli

import (
	"strings"

	"github.com/wxnacy/bdpan-cli/terminal"
)

// func GetHelpItems() []terminal.HelpItem {
// var items = []terminal.HelpItem{
// terminal.NewHelpItem('?', "帮助"),
// terminal.NewHelpItem('j', "向下移动一行").AddKey(tcell.KeyDown),
// terminal.NewHelpItem('k', "向上移动一行").AddKey(tcell.KeyUp),
// terminal.NewHelpItem('h', "返回上一层目录").AddKey(tcell.KeyLeft),
// terminal.NewHelpItem('l', "进入下一层目录，如果是文件执行下载操作").AddKey(tcell.KeyRight).AddKey(tcell.KeyEnter),
// terminal.NewHelpItem('G', "跳转到最后一个文件"),
// terminal.NewHelpItem(0, "跳转到第一个文件").SetKeyString("gg"),
// terminal.NewHelpItem(0, "向下翻一页").AddKey(tcell.KeyCtrlF),
// terminal.NewHelpItem(0, "向上翻一页").AddKey(tcell.KeyCtrlB),
// terminal.NewHelpItem(0, "向下翻半页").AddKey(tcell.KeyCtrlD),
// terminal.NewHelpItem(0, "向上翻半页").AddKey(tcell.KeyCtrlU),
// terminal.NewHelpItem('x', "剪切文件"),
// terminal.NewHelpItem('d', "下载文件"),
// terminal.NewHelpItem('D', "删除文件"),
// }

// runes := GetKeymapRunes()
// for _, r := range runes {
// relkeys := GetRelKeysByRune(r)
// if relkeys != nil {
// for _, k := range relkeys {
// items = append(items, terminal.NewHelpItem(0, k.Msg).SetKeyString(string(r)+string(k.R)))
// }
// }
// }
// return items
// }

func NewHelpItems(m Mode) []terminal.HelpItem {
	var items = make([]terminal.HelpItem, 0)
	keymaps := ModeKeymapsMap[m]
	for _, k := range keymaps {
		items = append(items, terminal.NewHelpItem(0, k.Desc).
			SetKeyString(strings.Join(k.Keys, "")))
	}
	return items
}
