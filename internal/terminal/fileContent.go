package terminal

import (
	"bytes"
	"io/ioutil"
	"log"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/pkg/whitetea"
)

// highlight a string
func highlight(lexerName, code string) (string, error) {
	lexer := lexers.Get(lexerName)
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return "", err
	}
	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func NewFileContent(filePath string, opts ...any) *FileContent {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	lexer := lexers.Match(filePath)
	lexerName := ""
	if lexer != nil {
		lexerName = lexer.Config().Name
	}

	highlightedCode, err := highlight(lexerName, string(content))
	if err != nil {
		log.Fatal(err)
	}

	width := 100
	height := 200
	var style lipgloss.Style

	for _, v := range opts {
		switch v := v.(type) {
		case lipgloss.Style:
			// m.baseStyle = v
			style = v
		case whitetea.Width:
			width = int(v)
		case whitetea.Height:
			height = int(v)
		}
	}
	// 创建一个新的 viewport
	vp := viewport.New(width, height)
	// 设置 viewport 的内容
	vp.SetContent(highlightedCode)
	vp.Style = style

	m := &FileContent{
		Model:  vp,
		KeyMap: DefaultFileContentKeyMap(),
	}
	return m
}

type FileContent struct {
	Model     viewport.Model
	baseStyle lipgloss.Style
	KeyMap    FileContentKeyMap
}

func (m FileContent) Init() tea.Cmd {
	return nil
}

func (m *FileContent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m FileContent) View() string {
	// view := m.Model.View()
	// return m.baseStyle.Render(view)
	return m.Model.View()
}

func (m *BDPan) FileContentFocused() bool {
	return m.fileContentModel != nil
}

func DefaultFileContentKeyMap() FileContentKeyMap {
	return FileContentKeyMap{
		Esc: key.NewBinding(
			key.WithKeys("esc", "i"),
			key.WithHelp("esc/i", "返回"),
		),
	}
}

type FileContentKeyMap struct {
	viewport.KeyMap
	Esc key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km FileContentKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Esc}
}

// FullHelp implements the KeyMap interface.
func (km FileContentKeyMap) FullHelp() [][]key.Binding {
	// TODO: 更新帮助文档
	return [][]key.Binding{
		{km.Esc},
	}
}
