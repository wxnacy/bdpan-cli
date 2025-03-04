package whitetea

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/logger"
)

// var docStyle = lipgloss.NewStyle().Margin(1, 2)

// type ListItem struct {
// title, desc string
// }

// func (i ListItem) Title() string       { return i.title }
// func (i ListItem) Description() string { return i.desc }
// func (i ListItem) FilterValue() string { return i.title }

type List struct {
	model         list.Model
	width, height int
	baseStyle     lipgloss.Style
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
		// h, v := m.baseStyle.GetFrameSize()
		_w := msg.Width
		if m.width > 0 {
			_w = m.width
		}
		_h := msg.Height
		if m.height > 0 {
			_h = m.height
		}
		logger.Infof("View List WindowSize %dx%d", _w, _h)
		m.model.SetSize(_w-3, _h-3)
	}

	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

func (m List) View() string {
	begin := time.Now()
	var viewW, viewH int
	var view = m.model.View()
	viewW, viewH = lipgloss.Size(view)
	logger.Infof("ListView Size %dx%d", viewW, viewH)
	view = m.baseStyle.Render(view)
	viewW, viewH = lipgloss.Size(view)
	logger.Infof("ListView Full Size %dx%d", viewW, viewH)
	logger.Infof("ListView time used %v", time.Now().Sub(begin))
	return view
}

func (m *List) SetItems(items []list.Item) *List {
	m.model.SetItems(items)
	return m
}

func (m *List) SetSelect(i int) *List {
	m.model.Select(i)
	return m
}

func (m *List) Width(w int) *List {
	m.width = w
	return m
}

func (m *List) GetWidth(w int) *List {
	m.width = w
	return m
}

func (m *List) Height(h int) *List {
	m.height = h
	return m
}

// func (m *List) SetItems(items []ListItem) *List {
// m.model.SetItems(toList(items))
// return m
// }

func NewList(title string, items []list.Item, opts ...interface{}) *List {
	m := &List{model: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.model.Title = title
	for _, v := range opts {
		switch v := v.(type) {
		case lipgloss.Style:
			m.baseStyle = v
		}
	}
	return m
}

// func toList(items []ListItem) []list.Item {
// _items := make([]list.Item, 0)
// for _, v := range items {
// _items = append(_items, v)
// }
// return _items
// }
