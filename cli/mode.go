package cli

type Mode string

const (
	ModeNormal  Mode = "normal"
	ModeConfirm      = "confirm"
	ModeKeymap       = "keymap"
	ModeHelp         = "help"
)
