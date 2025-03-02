package terminal

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Enter   key.Binding
	Back    key.Binding
	Delete  key.Binding
	Refresh key.Binding
	Exit    key.Binding

	// 复制组合键位
	CopyPath               key.Binding
	CopyDir                key.Binding
	CopyFilename           key.Binding
	CopyFilenameWithoutExt key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Exit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "退出"),
		),
		Enter: key.NewBinding(
			key.WithKeys("right", "l", "enter"),
			key.WithHelp("right/l/enter", "确认/打开"),
		),
		Back: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "退回"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "刷新当前目录"),
		),
		// Delete: key.NewBinding(
		// key.WithKeys("d", "d"),
		// key.WithHelp("dd", "删除"),
		// ),
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

func (k KeyMap) GetCombKeys() []key.Binding {
	bindings := make([]key.Binding, 0)
	bindings = append(bindings, k.GetCopyKeys()...)
	return bindings
}
