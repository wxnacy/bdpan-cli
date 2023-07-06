package cli

import (
	"github.com/gdamore/tcell/v2"
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

type Mode string

const (
	ModeNormal  Mode = "normal"
	ModeConfirm      = "confirm"
	ModeKeymap       = "keymap"
	ModeHelp         = "help"
	ModeSync         = "sync"
	ModeCommand      = "command"
	ModeFilter       = "filter"
)

type ActionFn func(KeymapAction) error
type CommandFn func(Command) error
type KeymapFn func(Keymap) error

type ModeInterface interface {
	GetMode() Mode
	GetActionFn() ActionFn
	GetKeymapFn() KeymapFn
	GetKeymaps() []Keymap
	GetKeymap() Keymap
	GetKeymapActionMap() map[string]KeymapAction
	SetEventKey(ev *tcell.EventKey)
	IsKey(Keymap) bool
	Draw() error
}

type BaseMode struct {
	T               *terminal.Terminal
	Mode            Mode
	ActionFn        ActionFn
	KeymapActionMap map[string]KeymapAction
	KeymapFn        KeymapFn
	Keymaps         []Keymap
	PrevEventKey    *tcell.EventKey
	CurrEventKey    *tcell.EventKey
}

func (b BaseMode) GetMode() Mode {
	return b.Mode
}

func (b *BaseMode) SetEventKey(ev *tcell.EventKey) {
	b.PrevEventKey = b.CurrEventKey
	b.CurrEventKey = ev
}

func (b *BaseMode) IsKey(k Keymap) bool {
	return k.IsKey(b.CurrEventKey, b.PrevEventKey)
}

func (b *BaseMode) SetActionFn(fn ActionFn) *BaseMode {
	b.ActionFn = fn
	return b
}

func (b *BaseMode) GetActionFn() ActionFn {
	return b.ActionFn
}

func (b *BaseMode) SetKeymapFn(fn KeymapFn) *BaseMode {
	b.KeymapFn = fn
	return b
}

func (b *BaseMode) GetKeymapFn() KeymapFn {
	return b.KeymapFn
}

func (b *BaseMode) SetKeymaps(k []Keymap) *BaseMode {
	k = append(k, CommandKeymaps...)
	b.Keymaps = k
	return b
}

func (b *BaseMode) GetKeymaps() []Keymap {
	return b.Keymaps
}

func (b *BaseMode) GetKeymap() Keymap {
	var key Keymap
	for _, k := range b.Keymaps {
		if b.IsKey(k) {
			key = k
		}
	}
	return key
}

func (b *BaseMode) SetKeymapActionMap(m map[string]KeymapAction) *BaseMode {
	b.KeymapActionMap = m
	return b
}

func (b *BaseMode) GetKeymapActionMap() map[string]KeymapAction {
	return b.KeymapActionMap
}

func (b *BaseMode) Draw() error {
	return nil
}

func NewFilterMode(filter string) *FilterMode {
	return &FilterMode{
		BaseMode: &BaseMode{},
		Filter:   filter,
	}
}

type FilterMode struct {
	*BaseMode
	Filter string
}

func (FilterMode) GetMode() Mode {
	return ModeFilter
}

func (f *FilterMode) SetFilter(t string) *FilterMode {
	f.Filter = t
	return f
}

//------------------------------
//  CommandMode
//------------------------------

func NewCommandMode(prefix string) *CommandMode {
	return &CommandMode{
		BaseMode: &BaseMode{
			Mode: ModeCommand,
		},
		Prefix: prefix,
	}
}

type CommandMode struct {
	*BaseMode
	NextMode ModeInterface
	Prefix   string
	Input    string
}

func (c *CommandMode) GetKeymap() Keymap {
	var key Keymap
	for _, k := range c.Keymaps {
		if c.IsKey(k) {
			key = k
		}
	}
	Log.Infof("Command Key %v", key)
	if key.IsNil() {
		key = KeymapInput
	}
	Log.Infof("Command Key %v", key)
	return key
}

func (m *CommandMode) SetInput(t string) *CommandMode {
	m.Input = t
	return m
}

func (m *CommandMode) SetNextMode(mi ModeInterface) *CommandMode {
	m.NextMode = mi
	return m
}

//------------------------------
//  NormalMode
//------------------------------

func NewNormalMode() *NormalMode {
	return &NormalMode{
		BaseMode: &BaseMode{
			Mode: ModeNormal,
		},
	}
}

type NormalMode struct {
	*BaseMode
}

//------------------------------
//  ConfirmMode
//------------------------------

func NewConfirmMode(t *terminal.Terminal, ensureC Command, msg string) *ConfirmMode {
	return &ConfirmMode{
		BaseMode: &BaseMode{
			Mode: ModeConfirm,
		},
		Msg:           msg,
		EnsureCommand: ensureC,
		Term:          terminal.NewConfirm(t, msg),
	}
}

type ConfirmMode struct {
	*BaseMode
	Msg           string
	Term          *terminal.Confirm
	EnsureCommand Command
}

//------------------------------
//  KeymapMode
//------------------------------
func NewKeymapMode(t *terminal.Terminal, firstRune rune) *KeymapMode {
	return &KeymapMode{
		BaseMode:  &BaseMode{T: t, Mode: ModeKeymap},
		FirstRune: firstRune,
	}
}

type KeymapMode struct {
	*BaseMode
	FirstRune  rune
	SecondRune rune
	Term       *terminal.List
}

func (k KeymapMode) GetKeyString() string {
	return string(k.FirstRune) + string(k.SecondRune)
}

func (k *KeymapMode) SetSecondRune(r rune) *KeymapMode {
	k.SecondRune = r
	return k
}

func (k *KeymapMode) Draw() error {
	var data []string
	var keymaps = GetRelKeysByRune(k.FirstRune)
	if keymaps != nil {
		data = GetRelKeysMsgByRune(k.FirstRune)
	}
	_, h := k.T.S.Size()
	startY := h - 3 - len(data)
	k.Term = terminal.NewList(k.T, 0, startY, data).SetMaxWidth().Draw()
	return nil
}

//------------------------------
//  SyncMode
//------------------------------
func NewSyncMode(
	t *terminal.Terminal, item *terminal.SelectItem,
	StartX, StartY, EndX, EndY int,
) *SyncMode {
	return &SyncMode{
		BaseMode:   &BaseMode{T: t, Mode: ModeSync},
		SelectItem: item,
		Term:       terminal.NewEmptySelect(t, StartX, StartY, EndX, EndY),
	}
}

type SyncMode struct {
	*BaseMode
	Term       *terminal.Select
	SelectItem *terminal.SelectItem
}

func (s *SyncMode) Count() int {
	selectFile := s.SelectItem.Info.(*FileInfo).FileInfoDto
	return len(bdpan.GetSyncModelsByRemote(selectFile.Path))
}

func (s *SyncMode) Draw() error {
	selectFile := s.SelectItem.Info.(*FileInfo).FileInfoDto
	syncTermH := len(bdpan.GetSyncModelsByRemote(selectFile.Path))
	if syncTermH == 0 {
		return nil
	}
	s.Term.Box.StartY = s.Term.Box.EndY - syncTermH - 1
	// 填充内容
	FillSyncToSelect(s.Term, selectFile)
	s.Term.Draw()
	return nil
}

//------------------------------
//  HelpMode
//------------------------------
func NewHelpMode(
	t *terminal.Terminal, fromMode Mode,
	StartX, StartY, EndX, EndY int,
) *HelpMode {
	Term := terminal.NewHelp(t, NewHelpItems(fromMode))
	Term.Box.StartX = StartX
	Term.Box.StartY = StartY
	Term.Box.EndX = EndX
	Term.Box.EndY = EndY
	return &HelpMode{
		BaseMode: &BaseMode{T: t, Mode: ModeHelp},
		Term:     Term,
	}
}

type HelpMode struct {
	*BaseMode
	Term *terminal.Help
}

func (h *HelpMode) Draw() error {
	h.Term.Draw()
	return nil
}
