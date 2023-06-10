package terminal

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func NewHelpItem(r rune, doc string) HelpItem {
	return HelpItem{
		Runes: []rune{r},
		Doc:   doc,
	}
}

type HelpItem struct {
	Runes     []rune
	Keys      []tcell.Key
	KeyString string
	Doc       string
}

func (h HelpItem) AddRune(r rune) HelpItem {
	h.Runes = append(h.Runes, r)
	return h
}

func (h HelpItem) AddKey(key tcell.Key) HelpItem {
	h.Keys = append(h.Keys, key)
	return h
}

func (h HelpItem) SetKeyString(s string) HelpItem {
	h.KeyString = s
	return h
}

func (h HelpItem) String() string {
	var keyName string
	if h.KeyString != "" {
		keyName = h.KeyString
	}
	for _, r := range h.Runes {
		if r > 0 {
			keyName += string(r) + "/"
		}
	}
	for _, k := range h.Keys {
		name, ok := tcell.KeyNames[k]
		if ok {
			keyName += name + "/"
		}
	}
	keyName = strings.TrimRight(keyName, "/")
	return fmt.Sprintf("%s\t\t%s", keyName, h.Doc)
}

func NewHelp(t *Terminal, items []HelpItem) *Help {
	w, h := t.S.Size()
	return &Help{
		Box:   t.NewBox(0, 0, w-1, h-1, StyleDefault),
		Items: items,
		t:     t,
	}
}

type Help struct {
	Box   *Box
	Items []HelpItem

	t *Terminal
}

func (h *Help) Draw() *Help {
	h.t.DrawBox(*h.Box)
	h.Box.Clean()
	for i, item := range h.Items {
		h.Box.DrawOneLineText(i, StyleDefault, "  "+item.String())
	}
	return h
}
