package terminal

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
)

type TaskType int

const (
	TaskDelete TaskType = iota
	TaskDownload
)

type Task struct {
	FSID uint64
	Path string
	Type TaskType
}

type ChangeFilesMsg struct {
	Files []*model.File
}

type ChangePanMsg struct {
	Pan *model.Pan
}

type ChangeUserMsg struct {
	User *model.User
}

func NewBDPan(dir string) (*BDPan, error) {
	files, err := handler.GetFileHandler().GetFiles(dir, 1)
	if err != nil {
		return nil, err
	}

	return &BDPan{
		Dir: dir,
		// useCache:    true,
		files:       files,
		filesMap:    make(map[string][]*model.File, 0),
		tasks:       make([]*Task, 0),
		fileHandler: handler.GetFileHandler(),
		authHandler: handler.GetAuthHandler(),
		KeyMap:      DefaultKeyMap(),
	}, nil
}

type BDPan struct {
	Dir string

	// Data
	// useCache bool
	files    []*model.File
	filesMap map[string][]*model.File
	tasks    []*Task
	pan      *model.Pan
	user     *model.User
	KeyMap   KeyMap

	// state
	// viewState         bool
	fileListViewState bool

	fileHandler *handler.FileHandler
	authHandler *handler.AuthHandler

	// Terminal
	width  int
	height int

	// model
	fileListModel *FileList
}

func (m *BDPan) GetWidth() int {
	if m.width > 0 {
		return m.width
	}
	return 100
}

func (m *BDPan) GetHeight() int {
	if m.height > 0 {
		return m.height
	}
	return 20
}

func (m *BDPan) getFiles(dir string) []*model.File {
	files, _ := m.filesMap[dir]
	return files
}

func (m *BDPan) SetFiles(files []*model.File) {
	m.files = files
	m.filesMap[m.Dir] = files
	m.fileListModel = m.NewFileList(m.files)
	m.DisableLoadingFileList()
}

func (m *BDPan) Init() tea.Cmd {
	begin := time.Now()
	logger.Infof("BDPan Init begin ============================")
	logger.Infof("Window size: %dx%d", m.width, m.height)

	// m.viewState = true
	logger.Infof("BDPan Init time used %v ====================", time.Now().Sub(begin))
	return tea.SetWindowTitle("bdpan")
}

func (m *BDPan) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger.Infof("BDPan Update =================================")
	var cmd tea.Cmd
	// var err error
	logger.Infof("Update by msg: %v", msg)

	// 先做原始修改操作
	if m.FileListModelIsNotNil() {
		m.fileListModel, cmd = m.fileListModel.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// 更改尺寸后,重新获取模型
		m.ChangeDir(m.Dir)
	case ChangeFilesMsg:
		// 异步加载文件列表
		m.SetFiles(msg.Files)
	case ChangePanMsg:
		// 异步加载 pan 信息
		m.pan = msg.Pan
	case ChangeUserMsg:
		// 异步加载 user 信息
		m.user = msg.User
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Back):
			m.ChangeDir(filepath.Dir(m.Dir))
		case key.Matches(msg, m.KeyMap.Enter):
			selectFile, err := m.fileListModel.GetSelectFile()
			if err != nil {
				tea.Quit()
				return m, tea.Quit
			}
			if selectFile.IsDir() {
				m.ChangeDir(selectFile.Path)
			}
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, cmd
}

func (m *BDPan) View() string {
	logger.Infof("BDPan View =================================")
	logger.Infof("Window size: %dx%d", m.width, m.height)

	midView := m.GetMidView()

	statusView := m.GetStatusView()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		midView,
		statusView,
	)
}

func (m *BDPan) GetFileListView() string {
	if !m.IsLoadingFileList() && m.FileListModelIsNotNil() {
		fileListView := m.fileListModel.View()
		logger.Infof("FileListView height %d", lipgloss.Height(fileListView))
		return fileListView
	} else {
		return m.NewFileList(nil).View()
	}
}

func (m *BDPan) GetMidView() string {

	// filelist
	centerView := m.GetFileListView()

	// fileinfo
	rightView := m.GetFileInfoView(nil)
	if m.FileListModelIsNotNil() {
		f, err := m.fileListModel.GetSelectFile()
		if err != nil {
			tea.Quit()
			return ""
		}
		rightView = m.GetFileInfoView(f)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		centerView,
		rightView,
	)
}

func (m *BDPan) GetStatusView() string {
	w := lipgloss.Width
	statusNugget := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 1)

	statusBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
		Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle := lipgloss.NewStyle().
		Inherit(statusBarStyle).
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#FF5F87")).
		Padding(0, 1).
		MarginRight(1)

	encodingStyle := statusNugget.
		Background(lipgloss.Color("#A550DF")).
		Align(lipgloss.Right)

	statusText := lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle := statusNugget.Background(lipgloss.Color("#6124DF"))

	// left
	used := "-/-"
	if m.pan != nil {
		used = fmt.Sprintf("%s/%s", m.pan.GetUsedStr(), m.pan.GetTotalStr())
	}
	statusKey := statusStyle.Render(fmt.Sprintf(
		"容量 %s",
		used,
	))

	// right
	encoding := ""
	fishCake := ""
	if !m.UserIsNil() {
		encoding = encodingStyle.Render(m.user.GetNetdiskName())
		fishCake = fishCakeStyle.Render(m.user.GetVipName())
	}

	// mid
	fileLineText := ""
	if m.FileListModelIsNotNil() {
		f, err := m.fileListModel.GetSelectFile()
		if err != nil {
			tea.Quit()
			return ""
		}
		fileLineText = fmt.Sprintf("%s", f.Path)
	}
	statusVal := statusText.
		Width(m.GetWidth() - w(statusKey) - w(encoding) - w(fishCake)).
		Render(fileLineText)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		statusKey,
		statusVal,
		encoding,
		fishCake,
	)
}

func (m *BDPan) GetFileInfoView(f *model.File) string {

	leftStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderLeft(true).
		Align(lipgloss.Left).
		// Foreground(lipgloss.Color("#FAFAFA")).
		// Background(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}).
		// Margin(1, 3, 0, 0).
		Padding(0, 1).
		Height(1).
		Width(10)
	rightStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderRight(true).
		Align(lipgloss.Left).
		// Foreground(lipgloss.Color("#FAFAFA")).
		// Background(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}).
		// Margin(1, 3, 0, 0).
		Padding(0, 1).
		Height(1).
		Width(50)

	lines := make([]string, 0)
	lines = append(lines, lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.
			// Border(fieldBorder).
			BorderTop(true).
			Render("字段"),
		rightStyle.
			BorderTop(true).
			// Border(contentBorder).
			Render("详情"),
	))

	if f != nil {

		lines = append(lines, lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.
				BorderTop(true).
				Render("FSID"),
			rightStyle.
				BorderTop(true).
				Render(fmt.Sprintf("%d", f.FSID)),
		))

		filename := fmt.Sprintf("%s %s", f.GetFileTypeEmoji(), f.GetFilename()) + "\n"
		nameStr := rightStyle.Render(filename)
		nameH := lipgloss.Height(nameStr)
		lines = append(lines, lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Height(nameH).Render("文件名"),
			nameStr,
		))

		lines = append(lines, lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Render("大小"),
			rightStyle.Render(f.GetSize()),
		))

		lines = append(lines, lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Render("类型"),
			rightStyle.Render(f.GetFileType()),
		))

		pathStr := rightStyle.Render(f.Path)
		pathH := lipgloss.Height(pathStr)
		lines = append(lines, lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Height(pathH).Render("地址"),
			pathStr,
		))

		if !f.IsDir() {
			lines = append(lines, lipgloss.JoinHorizontal(
				lipgloss.Top,
				leftStyle.Render("MD5"),
				rightStyle.Render(f.MD5),
			))
		}

		lines = append(lines, lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Render("创建时间"),
			rightStyle.Render(f.GetServerCTime()),
		))

		lines = append(lines, lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Render("修改时间"),
			rightStyle.Render(f.GetServerMTime()),
		))
	} else {
		lines = append(lines, lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Render(""),
			rightStyle.Render("数据加载中..."),
		))
	}

	lastBeforeH := lipgloss.Height(strings.Join(lines, "\n"))
	logger.Infof("lastBeforeH %d", lastBeforeH)
	lastH := m.GetMidHeight() - lastBeforeH - 2

	lines = append(lines, lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.
			Height(lastH).
			BorderBottom(true).
			Render(""),
		rightStyle.
			Height(lastH).
			BorderBottom(true).
			Render(""),
	))

	res := lipgloss.JoinVertical(
		lipgloss.Top,
		lines...,
	)
	return res

}

func (m *BDPan) GetMidHeight() int {
	height := m.GetHeight() - 2
	logger.Infof("GetMidHeight %d", height)
	return height
}

func (m *BDPan) NewFileList(files []*model.File) *FileList {
	return NewFileList(files, m.GetWidth()/2, m.GetMidHeight())
}

func (m *BDPan) EnableLoadingFileList() *BDPan {
	m.fileListViewState = false
	return m
}

func (m *BDPan) DisableLoadingFileList() *BDPan {
	m.fileListViewState = true
	return m
}

func (m *BDPan) IsLoadingFileList() bool {
	return !m.fileListViewState
}

func (m *BDPan) PanIsNil() bool {
	return m.pan == nil
}

func (m *BDPan) UserIsNil() bool {
	return m.user == nil
}

func (m *BDPan) FileListModelIsNotNil() bool {
	return m.fileListModel != nil
}

func (m *BDPan) FileListModelIsNil() bool {
	return m.fileListModel == nil
}

// 改变显示的目录
func (m *BDPan) ChangeDir(dir string) {
	m.Dir = dir
	files := m.getFiles(m.Dir)
	if files == nil {
		// 没有缓存时打开 Loading
		m.EnableLoadingFileList()
	} else {
		m.files = files
		m.fileListModel = m.NewFileList(m.files)
	}
}

type KeyMap struct {
	Enter  key.Binding
	Back   key.Binding
	Delete key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Enter: key.NewBinding(
			key.WithKeys("right", "l", "enter"),
			key.WithHelp("right/l/enter", "确认/打开"),
		),
		Back: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "退回"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "删除"),
		),
	}
}
