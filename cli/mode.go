package cli

import "github.com/wxnacy/bdpan-cli/terminal"

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

type ModeInterface interface {
	GetMode() Mode
	GetActionFn() ActionFn
}

type BaseMode struct {
	ActionFn ActionFn
}

func (b *BaseMode) SetActionFn(fn ActionFn) *BaseMode {
	b.ActionFn = fn
	return b
}

func (b *BaseMode) GetActionFn() ActionFn {
	return b.ActionFn
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
		BaseMode: &BaseMode{},
		Prefix:   prefix,
	}
}

type CommandMode struct {
	*BaseMode
	NextMode ModeInterface
	Prefix   string
	Input    string
}

func (CommandMode) GetMode() Mode {
	return ModeCommand
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

func NewNormalMode(prefix string) *NormalMode {
	return &NormalMode{
		BaseMode: &BaseMode{},
	}
}

type NormalMode struct {
	*BaseMode
}

func (NormalMode) GetMode() Mode {
	return ModeNormal
}

//------------------------------
//  ConfirmMode
//------------------------------

func NewConfirmMode(t *terminal.Terminal, msg string) *ConfirmMode {
	return &ConfirmMode{
		BaseMode: &BaseMode{},
		Msg:      msg,
		Term:     terminal.NewConfirm(t, msg),
	}
}

type ConfirmMode struct {
	*BaseMode
	Msg  string
	Term *terminal.Confirm
}

func (ConfirmMode) GetMode() Mode {
	return ModeConfirm
}
