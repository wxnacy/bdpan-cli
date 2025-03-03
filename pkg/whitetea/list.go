package whitetea

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type ListItem struct {
	title, desc string
}

func (i ListItem) Title() string       { return i.title }
func (i ListItem) Description() string { return i.desc }
func (i ListItem) FilterValue() string { return i.title }

type List struct {
	model         list.Model
	width, height int
}

func (m List) Init() tea.Cmd {
	return nil
}

func (m *List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.model.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

func (m List) View() string {
	return docStyle.Render(m.model.View())
}

func (m *List) SetItems(items []ListItem) *List {
	m.model.SetItems(toList(items))
	return m
}

// func (m *List) SetItems(items []ListItem) *List {
// m.model.SetItems(toList(items))
// return m
// }

func NewList(title string, items []ListItem) *List {
	m := &List{model: list.New(toList(items), list.NewDefaultDelegate(), 0, 0)}
	m.model.Title = title
	return m
}

func toList(items []ListItem) []list.Item {
	_items := make([]list.Item, 0)
	for _, v := range items {
		_items = append(_items, v)
	}
	return _items
}
