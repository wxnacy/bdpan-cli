package whitetea

import (
	"image/color"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/gamut"
	"github.com/wxnacy/bdpan-cli/internal/logger"
)

func NewConfirm(
	title string,
	opts ...interface{},
) *Confirm {
	c := &Confirm{
		title:  title,
		keymap: DefaultConfirmKeyMap(),
	}
	for _, v := range opts {
		switch v := v.(type) {
		case lipgloss.Style:
			c.baseStyle = v
		}
	}
	return c
}

type Confirm struct {
	title     string
	width     int
	baseStyle lipgloss.Style

	focus  bool
	value  bool
	keymap ConfirmKeyMap

	// 从哪个模型跳转过来的，可以退出聚焦回原模型
	fromModel Model
	data      ExtData
}

type ExtData any

func (m *Confirm) Init() tea.Cmd {
	return nil
}

func (m *Confirm) Update(msg tea.Msg) (*Confirm, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	logger.Infof("Update confirm begin ===========================")
	logger.Infof("Update confirm %v", msg)

	cmds = append(cmds, cmd)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Right):
			m.value = false
		case key.Matches(msg, m.keymap.Left):
			m.value = true
		case key.Matches(msg, m.keymap.Enter):
			m.Blur()
		case key.Matches(msg, m.keymap.Confirm):
			m.value = true
			m.Blur()
		case key.Matches(msg, m.keymap.Cancel, m.keymap.Exit):
			m.value = false
			m.Blur()
		}
	}
	logger.Infof("Update confirm end =============================")
	return m, tea.Batch(cmds...)
}

func (m *Confirm) View() string {
	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF7DB")).
		Background(lipgloss.Color("#888B7E")).
		Padding(0, 3).
		Margin(0, 1).
		MarginTop(1)

	activeButtonStyle := buttonStyle.
		Foreground(lipgloss.Color("#FFF7DB")).
		Background(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		// Background(lipgloss.Color("#F25D94")).
		// Margin(0, 1).
		Underline(true)

	okButton := buttonStyle.Render("[Y]es")
	cancelButton := activeButtonStyle.Render("[N]o")
	if m.value {
		okButton = activeButtonStyle.Render("[Y]es")
		cancelButton = buttonStyle.Render("[N]o")
	}

	blends := gamut.Blends(lipgloss.Color("#F25D94"), lipgloss.Color("#EDFF82"), 50)
	question := lipgloss.
		NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(rainbow(lipgloss.NewStyle(), m.title, blends))
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, okButton, cancelButton)
	ui := lipgloss.JoinVertical(lipgloss.Center, question, buttons)

	subtle := lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	dialog := lipgloss.Place(m.width, 5,
		lipgloss.Center, lipgloss.Center,
		m.baseStyle.Render(ui),
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(subtle),
	)

	return dialog
}

func (m *Confirm) Focused() bool {
	return m.focus
}

func (m *Confirm) Focus() {
	m.focus = true
}

func (m *Confirm) Blur() {
	m.focus = false
}

func (m *Confirm) Width(w int) *Confirm {
	m.width = w
	return m
}

func (m *Confirm) GetValue() bool {
	return m.value
}

func (m *Confirm) Value(v bool) *Confirm {
	m.value = v
	return m
}

func (m *Confirm) Data(d ExtData) *Confirm {
	m.data = d
	return m
}

func (m *Confirm) GetData() ExtData {
	return m.data
}

func (m *Confirm) Title(t string) *Confirm {
	m.title = t
	return m
}

func (m *Confirm) FromModel(model Model) *Confirm {
	m.fromModel = model
	return m
}

func (m *Confirm) GetFromModel() Model {
	return m.fromModel
}

func rainbow(base lipgloss.Style, s string, colors []color.Color) string {
	var str string
	for i, ss := range s {
		color, _ := colorful.MakeColor(colors[i%len(colors)])
		str = str + base.Foreground(lipgloss.Color(color.Hex())).Render(string(ss))
	}
	return str
}

type ConfirmKeyMap struct {
	Enter   key.Binding
	Right   key.Binding
	Left    key.Binding
	Confirm key.Binding
	Cancel  key.Binding
	Exit    key.Binding
}

func DefaultConfirmKeyMap() ConfirmKeyMap {
	return ConfirmKeyMap{
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "向右"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "向左"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter", "o"),
			key.WithHelp("enter/o", "确定"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y", "Y"),
			key.WithHelp("y/Y", "确认"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "N"),
			key.WithHelp("n/N", "取消"),
		),
		Exit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "退出"),
		),
	}
}
