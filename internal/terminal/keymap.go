package terminal

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Enter   key.Binding
	Back    key.Binding
	Delete  key.Binding
	Refresh key.Binding
	Exit    key.Binding
	Right   key.Binding
	Left    key.Binding

	// Pane
	MovePaneLeft  key.Binding
	MovePaneRight key.Binding

	// 复制组合键位
	CopyPath               key.Binding
	CopyDir                key.Binding
	CopyFilename           key.Binding
	CopyFilenameWithoutExt key.Binding

	// Goto 组合键位
	GotoRoot key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Exit,
		k.MovePaneLeft,
		k.MovePaneRight,
		k.Enter,
		k.Back,
	}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Back, k.Delete, k.Refresh}, // first column
		{k.Exit, k.Right, k.Left},              // second column
	}
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Exit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "退出"),
		),
		Enter: key.NewBinding(
			key.WithKeys("right", "l", "enter"),
			key.WithHelp("→/l/enter", "确认/打开"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "向右"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "向左"),
		),
		Back: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "退回"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "刷新当前目录"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "删除"),
		),

		MovePaneLeft: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "跳转左侧面板"),
		),
		MovePaneRight: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "跳转右侧面板"),
		),

		CopyDir: key.NewBinding(
			key.WithKeys("cd"),
			key.WithHelp("cd", "复制当前目录"),
		),
		CopyPath: key.NewBinding(
			key.WithKeys("cc"),
			key.WithHelp("cc", "复制文件地址"),
		),
		CopyFilename: key.NewBinding(
			key.WithKeys("cf"),
			key.WithHelp("cf", "复制文件名称"),
		),
		CopyFilenameWithoutExt: key.NewBinding(
			key.WithKeys("cn"),
			key.WithHelp("cn", "复制文件名称不含扩展"),
		),

		GotoRoot: key.NewBinding(
			key.WithKeys("g/"),
			key.WithHelp("g/", "Go to the Root dir"),
		),
	}
}

func (k KeyMap) GetCopyKeys() []key.Binding {
	return []key.Binding{
		k.CopyDir,
		k.CopyPath,
		k.CopyFilename,
		k.CopyFilenameWithoutExt,
	}
}

func (k KeyMap) GetCombKeys(keys []key.Binding) []key.Binding {
	bindings := make([]key.Binding, 0)
	bindings = append(bindings, k.GetCopyKeys()...)
	bindings = append(bindings, k.GetGotoKeys(nil)...)
	bindings = append(bindings, keys...)
	return bindings
}

func (k KeyMap) GetGotoKeys(keys []key.Binding) []key.Binding {
	bindings := []key.Binding{
		k.GotoRoot,
	}
	if keys != nil {
		bindings = append(bindings, keys...)
	}
	return bindings
}
