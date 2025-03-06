package whitetea

type Model interface {
	// Init is the first function that will be called. It returns an optional
	// initial command. To not perform an initial command return nil.
	// Init() tea.Cmd

	// Update is called when a message is received. Use it to inspect messages
	// and, in response, update the model and/or send a command.
	// Update(tea.Msg) (Model, tea.Cmd)

	// View renders the program's UI, which is just a string. The view is
	// rendered after every Update.
	// View() string

	// 是否聚焦
	Focused() bool

	// 聚焦
	Focus()

	// 取消聚焦
	Blur()
}
