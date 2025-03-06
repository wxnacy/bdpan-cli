package terminal

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
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
	keymap        QuickKeyMap
	focus         bool
}

func (m Quick) Init() tea.Cmd {
	return nil
}

func (m *Quick) Update(msg tea.Msg) (*Quick, tea.Cmd) {
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
		m.Width(_w).Height(_h)
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
	if m.Focused() {
		view = baseFocusStyle.Render(view)
	} else {
		view = baseStyle.Render(view)
	}
	viewW, viewH = lipgloss.Size(view)
	logger.Infof("ListView Full Size %dx%d", viewW, viewH)
	logger.Infof("ListView time used %v", time.Now().Sub(begin))
	return view
}

func (m *Quick) SetItems(items []list.Item) *Quick {
	m.model.SetItems(items)
	return m
}

func (m *Quick) Select(i int) *Quick {
	m.model.Select(i)
	return m
}

func (m *Quick) GetSelect() *model.Quick {
	return m.model.SelectedItem().(*model.Quick)
}

func (m *Quick) Width(w int) *Quick {
	m.width = w
	m.model.SetWidth(w - 3)
	return m
}

func (m *Quick) Height(h int) *Quick {
	m.height = h
	m.model.SetHeight(h - 3)
	return m
}

func (m *Quick) GetKeyMap() QuickKeyMap {
	return m.keymap
}

func (m *Quick) Focused() bool {
	return m.focus
}

func (m *Quick) Focus() {
	m.focus = true
}
func (m *Quick) Blur() {
	m.focus = false
}

func NewQuick(title string, items []*model.Quick, opts ...interface{}) *Quick {
	m := &Quick{
		model: list.New(
			model.ToList(items), list.NewDefaultDelegate(), 0, 0),
		keymap: DefaultQuickKeyMap(),
	}
	m.model.Title = title
	// 不展示帮助信息
	m.model.SetShowHelp(false)
	for _, v := range opts {
		switch v := v.(type) {
		case lipgloss.Style:
			m.baseStyle = v
		}
	}
	return m
}

type QuickKeyMap struct {
	Enter  key.Binding
	Exit   key.Binding
	Delete key.Binding
}

func DefaultQuickKeyMap() QuickKeyMap {
	return QuickKeyMap{
		Exit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "退出"),
		),
		Enter: key.NewBinding(
			key.WithKeys("o", "enter"),
			key.WithHelp("o/enter", "确认/打开"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "删除"),
		),
	}
}
