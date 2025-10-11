package terminal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	wtea "github.com/wxnacy/bdpan-cli/pkg/whitetea"
)

type Input struct {
	model     textarea.Model
	title     string
	baseStyle lipgloss.Style
	KeyMap    InputKeyMap

	// 从哪个模型跳转过来的，可以退出聚焦回原模型
	fromModel wtea.Model
}

func NewInput(
	title, initText string,
	opts ...any,
) *Input {
	ti := textarea.New()
	ti.InsertString(initText)
	ti.Focus()
	m := Input{
		model:  ti,
		title:  title,
		KeyMap: DefaultInputKeyMap(),
	}
	for _, v := range opts {
		switch v := v.(type) {
		case lipgloss.Style:
			m.baseStyle = v
		}
	}

	return &m
}

func (m Input) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Input) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.model.Focused() {
				cmd = m.model.Focus()
				cmds = append(cmds, cmd)
			}
		}
	}

	m.model, cmd = m.model.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Input) View() string {
	view := fmt.Sprintf(
		"%s\n\n%s",
		m.title,
		m.model.View(),
	)
	return m.baseStyle.Render(view)
}

func (m Input) Value() string {
	v := m.model.Value()
	v = strings.Trim(v, "\n")
	v = strings.TrimSpace(v)
	return v
}

func (m Input) Focus() {
	m.model.Focus()
}

func (m Input) Blur() {
	m.model.Blur()
}

func (m Input) Focused() bool {
	return m.model.Focused()
}

func (m Input) SetWidth(w int) {
	m.model.SetWidth(w)
}

func (m Input) SetHeight(h int) {
	m.model.SetHeight(h)
}

func (m *Input) SetFromModel(model wtea.Model) {
	m.fromModel = model
}

func (m *Input) GetFromModel() wtea.Model {
	return m.fromModel
}

type InputKeyMap struct {
	textarea.KeyMap
	Enter key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km InputKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Enter}
}

// FullHelp implements the KeyMap interface.
func (km InputKeyMap) FullHelp() [][]key.Binding {
	// TODO: 更新帮助文档
	return [][]key.Binding{
		{km.Enter},
	}
}

func DefaultInputKeyMap() InputKeyMap {
	return InputKeyMap{
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "确认"),
		),
	}
}
