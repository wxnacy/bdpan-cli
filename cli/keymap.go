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

	KeymapActionDeleteFile

	KeymapActionPasteFile

	KeymapActionDownloadFile
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
