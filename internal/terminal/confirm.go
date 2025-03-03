package terminal

import (
	"image/color"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/gamut"
	"github.com/wxnacy/bdpan-cli/internal/logger"
)

type FocusMsg tea.Msg
type BlurMsg tea.Msg

func NewConfirm(
	title string,
) *Confirm {
	return &Confirm{
		title: title,
		model: huh.NewConfirm().
			Title(title).
			Affirmative("Yes!").
			Negative("No."),
		keymap: DefaultConfirmKeyMap(),
		// windowView:   windowView,
		// windowWidth:  windowWidth,
		// windowHeight: windowHeight,
	}
}

type Confirm struct {
	title string
	// windowView   string
	// windowWidth  int
	// windowHeight int
	width int

	model  *huh.Confirm
	focus  bool
	value  bool
	keymap ConfirmKeyMap

	task *Task
}

func (m *Confirm) Init() tea.Cmd {
	m.model.Focus()
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
			m.task.IsConfirm = m.value
			m.Blur()
		case key.Matches(msg, m.keymap.Confirm):
			m.value = true
			m.task.IsConfirm = m.value
			m.Blur()
		case key.Matches(msg, m.keymap.Cancel):
			m.value = false
			m.task.IsConfirm = m.value
			m.Blur()
		}
	}

	logger.Infof("Update confirm end =============================")
	return m, tea.Batch(cmds...)
}

func (m *Confirm) view() string {
	view := baseStyle.
		Width(m.width).
		Render(m.model.View())
	logger.Infof("Confirm view size %d %d", m.width, lipgloss.Height(view))
	return view
}

func (m *Confirm) Focused() bool {
	return m.focus
}
func (m *Confirm) Focus() *Confirm {
	m.focus = true
	m.model.Focus()
	return m
}
func (m *Confirm) Blur() *Confirm {
	m.focus = false
	m.model.Blur()
	return m
}

func (m *Confirm) Width(w int) *Confirm {
	m.width = w
	return m
}

func (m *Confirm) GetValue() bool {
	return m.value
}

func (m *Confirm) SetTask(t *Task) *Confirm {
	m.task = t
	return m
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
		Background(lipgloss.Color("#F25D94")).
		// Margin(0, 1).
		Underline(true)

	// dialogBoxStyle := lipgloss.NewStyle().
	// Border(lipgloss.RoundedBorder()).
	// BorderForeground(lipgloss.Color("#874BFD")).
	// Padding(1, 0).
	// BorderTop(true).
	// BorderLeft(true).
	// BorderRight(true).
	// BorderBottom(true)

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
		baseStyle.Render(ui),
		// dialogBoxStyle.Render(ui),
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(subtle),
	)

	return dialog
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
	}
}
