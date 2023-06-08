package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	Keymaps = []Keymap{
		Keymap{R: 'y', RelKeys: []Keymap{
			Keymap{R: 'd', Msg: "复制所在文件夹名称"},
			Keymap{R: 'n', Msg: "复制文件名称"},
			Keymap{R: 'p', Msg: "复制文件路径"},
			Keymap{R: 'y', Msg: "复制文件"},
		}},
		Keymap{R: 'p', RelKeys: []Keymap{
			Keymap{R: 'p', Msg: "粘贴文件"},
		}},
	}

	KeyActionMap = map[string]KeymapAction{
		"yp": KeymapActionCopyPath,
		"yn": KeymapActionCopyName,
		"yd": KeymapActionCopyDir,
		"yy": KeymapActionCopyFile,

		"pp": KeymapActionPasteFile,
	}
)

type KeymapAction int

const (
	KeymapActionCopyPath KeymapAction = iota + 1
	KeymapActionCopyName
	KeymapActionCopyDir
	KeymapActionCopyFile

	KeymapActionCutFile

	KeymapActionPasteFile
)

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
			rStr := strings.ReplaceAll(strconv.QuoteRune(k.R), "'", "")
			var msg = fmt.Sprintf("  %s\t%s", rStr, k.Msg)
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

type Keymap struct {
	R       rune
	Key     tcell.Key
	Msg     string
	RelKeys []Keymap
}
