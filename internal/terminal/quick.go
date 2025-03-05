package terminal

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
)

type Quick struct {
	model         list.Model
	width, height int
	baseStyle     lipgloss.Style
}

func (m Quick) Init() tea.Cmd {
	return nil
}

func (m *Quick) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m Quick) View() string {
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

func (m *Quick) SetItems(items []list.Item) *Quick {
	m.model.SetItems(items)
	return m
}

func (m *Quick) SetSelect(i int) *Quick {
	m.model.Select(i)
	return m
}

func (m *Quick) Width(w int) *Quick {
	m.width = w
	return m
}

func (m *Quick) GetWidth(w int) *Quick {
	m.width = w
	return m
}

func (m *Quick) Height(h int) *Quick {
	m.height = h
	return m
}

func NewQuick(title string, items []*model.Quick, opts ...interface{}) *Quick {
	m := &Quick{model: list.New(
		model.ToList(items), list.NewDefaultDelegate(), 0, 0)}
	m.model.Title = title
	for _, v := range opts {
		switch v := v.(type) {
		case lipgloss.Style:
			m.baseStyle = v
		}
	}
	return m
}
