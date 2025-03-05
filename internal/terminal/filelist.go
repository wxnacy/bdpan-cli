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
	var cmd tea.Cmd
	// 先做原始修改操作
	m.model, cmd = m.model.Update(msg)
	logger.Infof("光标移动后的对象: %v", m.model.SelectedRow())

	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	// switch msg.String() {
	// case "esc":
	// if m.model.Focused() {
	// m.model.Blur()
	// } else {
	// m.model.Focus()
	// }
	// }
	// }
	return m, cmd
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
			// case key.Matches(msg, m.keymap.Refresh):
			// 刷新列表
			// return true, tea.Quit
			// case m.fileListModel.Focused():
			// // 光标聚焦在文件列表中
			// // 先做原始修改操作
			// if m.FileListModelIsNotNil() {
			// m.fileListModel, cmd = m.fileListModel.Update(msg)
			// }
			// switch {
			// case key.Matches(msg, m.KeyMap.Delete):
			// // 删除
			// m.fileListModel.Blur()
			// if m.FileListModelIsNotNil() {
			// f, err := m.GetSelectFile()
			// if err != nil {
			// return true, tea.Quit
			// }
			// task := m.AddDeleteTask(f)
			// m.confirmModel = NewConfirm("确认删除？").
			// Width(m.GetRightWidth()).
			// SetTask(task).
			// Focus()
			// }

			// case key.Matches(msg, m.KeyMap.Back):
			// // 返回目录
			// m.ChangeDir(filepath.Dir(m.Dir))
			// case key.Matches(msg, m.KeyMap.Enter):
			// selectFile, err := m.fileListModel.GetSelectFile()
			// if err != nil {
			// return true, tea.Quit
			// }
			// if selectFile.IsDir() {
			// m.ChangeDir(selectFile.Path)
			// }
			// }
		}
	default:
		flag = false
	}
	return flag, cmd
}

func (m *FileList) Focus() *FileList {
	m.model.Focus()
	return m
}

func (m *FileList) Blur() *FileList {
	m.model.Blur()
	return m
}

func (m FileList) Focused() bool {
	return m.model.Focused()
}

func (m *FileList) Size(w, h int) *FileList {
	m.model.SetWidth(w)
	m.model.SetHeight(h)
	return m
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
	Enter   key.Binding
	Back    key.Binding
	Delete  key.Binding
	Refresh key.Binding
	Exit    key.Binding
	Right   key.Binding
	Left    key.Binding

	// 复制组合键位
	CopyPath               key.Binding
	CopyDir                key.Binding
	CopyFilename           key.Binding
	CopyFilenameWithoutExt key.Binding
}

func DefaultFileListKeyMap() FileListKeyMap {
	return FileListKeyMap{
		Exit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "退出"),
		),
		Enter: key.NewBinding(
			key.WithKeys("right", "l", "enter"),
			key.WithHelp("right/l/enter", "确认/打开"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "向右"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "向左"),
		),
		Back: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "退回"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "刷新当前目录"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "删除"),
		),
		CopyDir: key.NewBinding(
			key.WithKeys("cd"),
			key.WithHelp("cd", "复制当前目录"),
		),
		CopyPath: key.NewBinding(
			key.WithKeys("cc"),
			key.WithHelp("cc", "复制文件地址"),
		),
		CopyFilename: key.NewBinding(
			key.WithKeys("cf"),
			key.WithHelp("cf", "复制文件名称"),
		),
		CopyFilenameWithoutExt: key.NewBinding(
			key.WithKeys("cn"),
			key.WithHelp("cn", "复制文件名称不含扩展"),
		),
	}
}

func (k FileListKeyMap) GetCopyKeys() []key.Binding {
	return []key.Binding{
		k.CopyDir,
		k.CopyPath,
		k.CopyFilename,
		k.CopyFilenameWithoutExt,
	}
}

func (k FileListKeyMap) GetCombKeys() []key.Binding {
	bindings := make([]key.Binding, 0)
	bindings = append(bindings, k.GetCopyKeys()...)
	return bindings
}
