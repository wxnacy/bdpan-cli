package cli

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
