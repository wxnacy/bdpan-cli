package terminal

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))
var baseFocusStyle = baseStyle.
	BorderForeground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"})

func GetBaseStyleWidth() int {
	return 2
}
func GetBaseStyleHeight() int {
	return 2
}

var (
	baseStyleWidth  int
	baseStyleHeight int
)

type FileList struct {
	model  table.Model
	files  []*model.File
	keymap FileListKeyMap
}

func (m *FileList) GetSelectFile() (*model.File, error) {
	id := m.model.SelectedRow()[0]
	fsid, err := strconv.Atoi(id)
	logger.Infof("select file fsid %d", fsid)
	if err != nil {
		return nil, err
	}

	for _, v := range m.files {
		if v.FSID == uint64(fsid) {
			return v, nil
		}
	}
	return nil, fmt.Errorf("%d not found", fsid)
}

func (m *FileList) Init() tea.Cmd { return nil }

func (m *FileList) Update(msg tea.Msg) (*FileList, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	// 先做原始修改操作
	m.model, cmd = m.model.Update(msg)
	cmds = append(cmds, cmd)
	logger.Infof("光标移动后的对象: %v", m.model.SelectedRow())

	_, cmd = m.ListenKeyMsg(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m FileList) View() string {
	var view string
	var viewW, viewH int
	view = m.model.View()
	viewW, viewH = lipgloss.Size(view)
	logger.Infof("FileListView Table Size %dx%d", viewW, viewH)
	if m.Focused() {
		view = baseFocusStyle.Render(view)
	} else {
		view = baseStyle.Render(view)
	}
	viewW, viewH = lipgloss.Size(view)
	logger.Infof("FileListView Full Size %dx%d", viewW, viewH)
	return view
}

func (m *FileList) ListenKeyMsg(msg tea.Msg) (bool, tea.Cmd) {
	flag := true
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Exit):
			// 退出程序
			return true, tea.Quit
		}
	default:
		flag = false
	}
	return flag, cmd
}

func (m *FileList) Focus() {
	m.model.Focus()
}

func (m *FileList) Blur() {
	m.model.Blur()
}

func (m FileList) Focused() bool {
	return m.model.Focused()
}

func (m FileList) GetKeyMap() FileListKeyMap {
	return m.keymap
}

func NewFileList(files []*model.File, width, height int) *FileList {
	var sizeW int = 10
	var typeW int = 10
	var timeW int = 20
	// 8 是为了和传进来的width保持一致设定的数字
	var filenameW int = width - sizeW - typeW - timeW - GetBaseStyleWidth() - 8
	columns := []table.Column{
		{Title: "FSID", Width: 0},
		{Title: "文件名", Width: filenameW},
		{Title: "大小", Width: sizeW},
		{Title: "类型", Width: typeW},
		{Title: "修改时间", Width: timeW},
	}

	rows := make([]table.Row, 0)
	if files == nil {
		rows = append(rows, table.Row{
			"",
			"数据加载中...",
			"",
			"",
			"",
		})
	} else {
		for _, f := range files {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", f.FSID),
				fmt.Sprintf("%s %s", f.GetFileTypeEmoji(), f.GetFilename()),
				f.GetSize(),
				f.GetFileType(),
				f.GetServerMTime(),
			})
		}
		if len(rows) == 0 {
			rows = append(rows, table.Row{
				"",
				"空",
				"",
				"",
				"",
			})
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height-4),
	)
	logger.Infof("Table height %d", t.Height())

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return &FileList{
		model:  t,
		files:  files,
		keymap: DefaultFileListKeyMap(),
	}
}

type FileListKeyMap struct {
	Exit     key.Binding
	AddQuick key.Binding
}

func DefaultFileListKeyMap() FileListKeyMap {
	return FileListKeyMap{
		Exit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "退出"),
		),
		AddQuick: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "添加快速访问"),
		),
	}
}
