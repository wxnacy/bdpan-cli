package terminal

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/pkg/bdtools"
	"github.com/wxnacy/bdpan-cli/pkg/whitetea"
	wtea "github.com/wxnacy/bdpan-cli/pkg/whitetea"
	"github.com/wxnacy/go-bdpan"
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

func NewFileContent(filePath string, opts ...any) (*FileContent, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lexer := lexers.Match(filePath)
	lexerName := ""
	if lexer != nil {
		lexerName = lexer.Config().Name
	} else {
		contentType := http.DetectContentType(content)
		if strings.HasPrefix(contentType, "text/") {
			lexerName = "bash"
		}
	}
	if lexerName == "" {
		return nil, fmt.Errorf("该文件不支持预览")
	}

	var highlightedCode string
	if lexerName == "markdown" {
		highlightedCode, err = glamour.Render(string(content), "dark")
	} else {
		highlightedCode, err = highlight(lexerName, string(content))
	}
	if err != nil {
		return nil, err
	}

	width := 100
	height := 200
	var style lipgloss.Style

	for _, v := range opts {
		switch v := v.(type) {
		case lipgloss.Style:
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
	return m, nil
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

func (m *BDPan) CanPreviewFile(f *bdpan.FileInfo) (bool, tea.Cmd) {
	if f.IsDir() {
		return false, m.SendMessage("文件夹不支持预览")
	}
	if f.Category != 4 && f.Category != 6 {
		return false, m.SendMessage("该文件不支持预览")
	}
	filename := f.GetFilename()
	unsupportedSuffixes := []string{
		".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".dmg", ".iso", ".exe", ".jar", ".bin", ".img", ".dat",
	}
	for _, suffix := range unsupportedSuffixes {
		if strings.HasSuffix(filename, suffix) {
			return false, m.SendMessage("该文件不支持预览")
		}
	}

	if f.Size > 1024*1024 {
		return false, m.SendMessage("文件过大，请下载后再查看")
	}
	var cmd tea.Cmd
	return true, cmd
}

func (m *BDPan) PreviewFile(f *bdpan.FileInfo) (bool, tea.Cmd) {
	p, err := bdtools.DownloadFileToLocal(m.fileHandler.GetAccessToken(), f)
	if err != nil {
		return true, m.SendMessage("预览文件失败: %s", err.Error())
	}

	model, err := NewFileContent(p,
		baseFocusStyle,
		wtea.Width(m.GetRightWidth()),
		wtea.Height(m.GetMidHeight()-3),
	)
	if err != nil {
		return true, m.SendMessage("预览文件失败: %s", err.Error())
	}
	m.fileContentModel = model
	m.fileListModel.Blur()
	var cmd tea.Cmd
	return true, cmd
}
