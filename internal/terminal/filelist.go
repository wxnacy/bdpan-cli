package terminal

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
	"github.com/wxnacy/bdpan-cli/pkg/bdtools"
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
			// logger.Infof("NewFileList file %#v", f)
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
		model:   t,
		files:   files,
		KeyMap:  DefaultFileListKeyMap(),
		TaskMap: DefaultFileListTaskMap(),
	}
}

type FileList struct {
	model   table.Model
	files   []*model.File
	KeyMap  FileListKeyMap
	TaskMap FileListTaskMap
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
	Exit        key.Binding
	Enter       key.Binding
	Back        key.Binding
	Refresh     key.Binding
	Space       key.Binding // 空格，选中
	SelectAll   key.Binding // 选中全部
	Delete      key.Binding
	Cut         key.Binding // 剪切
	Paste       key.Binding // 黏贴
	AddQuick    key.Binding // 添加快速访问
	Rename      key.Binding // 重命名
	ShowContent key.Binding // 显示内容
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
		SelectAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "选中全部"),
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
		ShowContent: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "显示内容"),
		),
	}
}

func DefaultFileListTaskMap() FileListTaskMap {
	return FileListTaskMap{
		AddQuick: TaskBinding{
			Title: "Add Quick",
			Type:  "add_quick",
		},
		ShowContent: TaskBinding{
			Title: "Show Content",
			Type:  "show_content",
		},
	}
}

type FileListTaskMap struct {
	AddQuick    TaskBinding
	ShowContent TaskBinding
}

func (m *BDPan) ListenFileListKeyMsg(msg tea.Msg) (bool, tea.Cmd) {
	flag := true
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 光标聚焦在文件列表中
		if m.FileListModelIsNotNil() {
			// 先做原始修改操作
			// 空格操作不使用原始操作
			if !key.Matches(msg, m.fileListModel.KeyMap.Space) {
				m.fileListModel, cmd = m.fileListModel.Update(msg)
				cmds = append(cmds, cmd)
			}

			// 记录光标位置
			m.fileCursorMap[m.Dir] = m.fileListModel.GetCursor()
		}
		switch {
		case key.Matches(msg, m.KeyMap.MovePaneLeft):
			// 向左移动面板
			m.quickModel.Focus()
			m.fileListModel.Blur()
		case key.Matches(msg, m.fileListModel.KeyMap.Delete):
			// 删除
			if m.CanSelectFile() {
				f, err := m.GetSelectFile()
				if err != nil {
					return true, tea.Quit
				}
				task := m.AddFileTask(f, TypeDelete)
				cmd = m.SendShowConfirm(
					fmt.Sprintf("确认删除 %s?", f.GetFilename()),
					task,
					m.fileListModel,
				)
				cmds = append(cmds, cmd)
			}
		case key.Matches(msg, m.fileListModel.KeyMap.Space):
			// 选中
			if m.CanSelectFile() {
				f, err := m.GetSelectFile()
				if err != nil {
					return true, tea.Quit
				}
				path := f.Path
				_, exist := m.selectFileMap[path]
				if exist {
					delete(m.selectFileMap, path)
				} else {
					m.selectFileMap[path] = f
				}
				cmds = append(cmds, m.SendMessage("选中文件: %s", f.Path))
				// 选中后向下移动一行
				m.fileListModel.model.MoveDown(1)
				// 记录光标位置
				m.fileCursorMap[m.Dir] = m.fileListModel.GetCursor()
				// 重新设置文件列表，带有选中效果
				m.fileListModel = m.NewFileList(m.files)
			}
		case key.Matches(msg, m.fileListModel.KeyMap.SelectAll):
			// 选中全部
			if m.CanSelectFile() {
				for _, file := range m.GetFiles(m.Dir) {

					path := file.Path
					_, exist := m.selectFileMap[path]
					if exist {
						delete(m.selectFileMap, path)
					} else {
						m.selectFileMap[path] = file
					}
				}
				// 重新设置文件列表，带有选中效果
				m.fileListModel = m.NewFileList(m.files)
			}
		case key.Matches(msg, m.fileListModel.KeyMap.Cut):
			// 剪切
			if m.CanSelectFile() {
				// 确认剪切文件
				if m.HasSelectFile() {
					m.ClearCutSelectFileMap()
					m.cutSelectFileMap = m.selectFileMap
					m.ClearSelectFileMap()
				} else {
					f, err := m.GetSelectFile()
					if err != nil {
						return true, tea.Quit
					}
					m.cutSelectFileMap[f.Path] = f
				}
				// 重新设置文件列表，带有选中效果
				m.fileListModel = m.NewFileList(m.files)
				cmds = append(cmds, m.SendMessage(
					"剪切文件: %s",
					strings.Join(m.GetCutSelectFilePaths(), " "),
				))
			}
		case key.Matches(msg, m.fileListModel.KeyMap.Paste):
			// 黏贴
			task := m.AddFileTask(nil, TypePaste)
			task.Dir = m.Dir
			cmds = append(cmds, m.SendRunTask(task))
		case key.Matches(msg, m.fileListModel.KeyMap.Back):
			// 返回目录
			cmds = append(cmds, m.Goto(filepath.Dir(m.Dir)))
		case key.Matches(msg, m.fileListModel.KeyMap.Enter):
			if m.CanSelectFile() {
				f, err := m.fileListModel.GetSelectFile()
				if err != nil {
					return true, tea.Quit
				}
				if f.IsDir() {
					cmds = append(cmds, m.Goto(f.Path))
				} else {
					task := m.AddFileTask(f, TypeDownload)
					cmd = m.SendShowConfirm(
						fmt.Sprintf("确认下载 %s?", f.GetFilename()),
						task,
						m.fileListModel,
					)
					cmds = append(cmds, cmd)
				}
			}
		case key.Matches(msg, m.fileListModel.KeyMap.AddQuick):
			// 添加快速访问
			if m.CanSelectFile() {
				f, err := m.GetSelectFile()
				if err != nil {
					return true, tea.Quit
				}
				if f.IsDir() {
					q := m.GetQuickByPath(f.Path)
					if q != nil {
						cmds = append(cmds, m.SendMessage("该目录已存在快速访问"))
					} else {
						// 添加快速访问请求
						// task := m.AddFileTask(f, TaskAddQuick)
						task := m.AddTask(m.fileListModel.TaskMap.AddQuick)
						task.Data = f
						cmds = append(cmds, m.SendShowInput(
							"请输入快速访问 Key", "",
							task,
							m.fileListModel,
						))
					}

				} else {
					cmds = append(cmds, m.SendMessage("文件不支持添加快速访问"))
				}
			}
		case key.Matches(msg, m.fileListModel.KeyMap.Rename):
			// 重命名
			if m.CanSelectFile() {
				if m.HasSelectFile() {
					// 批量重命名: 设置状态并退出TUI以运行编辑器
					m.NextAction = "batch-rename"
					m.ActionPayload = m.GetSelectFiles()
					return true, tea.Quit
				} else {
					// 单个
					f, err := m.GetSelectFile()
					if err != nil {
						return true, tea.Quit
					}
					filename := f.GetFilename()
					task := m.AddFileTask(f, TypeRename)
					cmds = append(cmds, m.SendShowInput(
						"请输入新名称", filename,
						task,
						m.fileListModel,
					))
				}
			}
		case key.Matches(msg, m.fileListModel.KeyMap.ShowContent):
			// 显示内容
			if !m.CanSelectFile() {
				return true, m.SendMessage("没有文件展示")
			}
			f, err := m.GetSelectFile()
			if err != nil {
				return true, m.SendMessage("选择文件失败 %s", err.Error())
			}
			flag, cmd := m.CanPreviewFile(f.FileInfo)
			if !flag {
				return true, cmd
			}
			if bdtools.HasLocalFile(f.FileInfo) {
				return m.PreviewFile(f.FileInfo)
			} else {
				task := m.AddTask(m.fileListModel.TaskMap.ShowContent)
				task.Data = f.FileInfo
				cmds = append(
					cmds,
					// m.SendMessage("开始请求文件"),
					m.SendRunTask(task),
				)
			}
		}
	}
	return flag, tea.Batch(cmds...)
}
