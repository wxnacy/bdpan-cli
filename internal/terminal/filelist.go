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
	"github.com/wxnacy/go-tools"
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

type FileListArgType int

const (
	FLTypeSelect FileListArgType = iota
	FLTypeSelectCut
)

type FileListArg struct {
	Type  FileListArgType
	Files []*model.File
}

func NewFileList(
	files []*model.File, width, height int, selectors []string,
	args ...FileListArg,
) *FileList {
	sizeW := 10
	typeW := 10
	timeW := 20
	// 8 是为了和传进来的width保持一致设定的数字
	filenameW := width - sizeW - typeW - timeW - GetBaseStyleWidth() - 8
	columns := []table.Column{
		{Title: "FSID", Width: 0},
		{Title: "文件名", Width: filenameW},
		{Title: "大小", Width: sizeW},
		{Title: "类型", Width: typeW},
		{Title: "修改时间", Width: timeW},
	}

	// 剪切选中集合
	var cutSelectors []string
	for _, arg := range args {
		switch arg.Type {
		case FLTypeSelectCut:
			for _, f := range arg.Files {
				cutSelectors = append(cutSelectors, f.Path)
			}
		}
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
			selectIcon := ""
			if tools.ArrayContainsString(selectors, f.Path) {
				selectIcon = " "
			}
			if tools.ArrayContainsString(cutSelectors, f.Path) {
				selectIcon = " "
			}
			logger.Infof("NewFileList file %#v", f)
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", f.FSID),
				fmt.Sprintf("%s%s %s", selectIcon, f.GetFileTypeEmoji(), f.GetFilename()),
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
		KeyMap: DefaultFileListKeyMap(),
	}
}

type FileList struct {
	model  table.Model
	files  []*model.File
	KeyMap FileListKeyMap
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

// 设置光标
func (m *FileList) Cursor(i int) *FileList {
	m.model.SetCursor(i)
	return m
}

func (m *FileList) GetCursor() int {
	return m.model.Cursor()
}

func (m *FileList) Init() tea.Cmd { return nil }

func (m *FileList) Update(msg tea.Msg) (*FileList, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	// 先做原始修改操作
	m.model, cmd = m.model.Update(msg)
	cmds = append(cmds, cmd)
	logger.Infof("光标移动后的对象: %v", m.model.SelectedRow())

	// _, cmd = m.ListenKeyMsg(msg)
	// cmds = append(cmds, cmd)

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

// func (m *FileList) ListenKeyMsg(msg tea.Msg, mainM *BDPan) (bool, tea.Cmd) {
// flag := true
// var cmd tea.Cmd
// switch msg := msg.(type) {
// case tea.KeyMsg:
// switch {
// case key.Matches(msg, m.KeyMap.Exit):
// // 退出程序
// return true, tea.Quit
// }
// default:
// flag = false
// }
// return flag, cmd
// }

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
	return m.KeyMap
}

type FileListKeyMap struct {
	table.KeyMap
	Exit     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Refresh  key.Binding
	Space    key.Binding // 空格，选中
	Delete   key.Binding
	Cut      key.Binding // 剪切
	Paste    key.Binding // 黏贴
	AddQuick key.Binding // 添加快速访问
	Rename   key.Binding // 重命名
}

// ShortHelp implements the KeyMap interface.
func (km FileListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.LineUp, km.LineDown, km.Back, km.Enter}
}

// FullHelp implements the KeyMap interface.
func (km FileListKeyMap) FullHelp() [][]key.Binding {
	// TODO: 更新帮助文档
	return km.KeyMap.FullHelp()
}

func DefaultFileListKeyMap() FileListKeyMap {
	return FileListKeyMap{
		KeyMap: table.DefaultKeyMap(),
		Exit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "退出"),
		),
		Enter: KeyEnter,
		Back: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "退回"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "刷新当前目录"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("Space", "选中"),
		),
		Delete: KeyDelete,
		Cut: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "剪切"),
		),
		Paste: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "黏贴"),
		),
		AddQuick: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "添加快速访问"),
		),
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "重命名"),
		),
	}
}
