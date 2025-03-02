package terminal

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
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

type ChangeMessageMsg struct {
	Message string
}

func NewBDPan(dir string) (*BDPan, error) {
	files, err := handler.GetFileHandler().GetFiles(dir, 1)
	if err != nil {
		return nil, err
	}

	return &BDPan{
		Dir:      dir,
		files:    files,
		filesMap: make(map[string][]*model.File, 0),
		tasks:    make([]*Task, 0),
		// message:     "初始化",
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
	message  string

	KeyMap  KeyMap
	lastKey *tea.KeyMsg

	// state
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
	begin := time.Now()
	logger.Infof("BDPan Update begin ===========================")
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
	case ChangeMessageMsg:
		// 接收信息
		m.SetMessage(msg.Message)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Delete):
			logger.Infof("do delete")
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

		// 监听两个键位的组合键位
		switch {
		case m.MatcheKeys(msg, m.KeyMap.CopyPath):
			logger.Infof("复制文件地址")
			selectFile, err := m.GetSelectFile()
			if !m.IsLoadingFileList() && err == nil {
				clipboard.WriteAll(selectFile.Path)
				m.SetMessage(fmt.Sprintf("地址 '%s' 复制到剪切板中", selectFile.Path))
			} else {
				m.SetMessage("数据加载中，稍后再试...")
			}
			m.ClearLastKey()
		case m.MatcheKeys(msg, m.KeyMap.CopyFilename):
			logger.Infof("复制文件名称")
			selectFile, err := m.GetSelectFile()
			if !m.IsLoadingFileList() && err == nil {
				filename := selectFile.GetFilename()
				clipboard.WriteAll(filename)
				m.SetMessage(fmt.Sprintf("文件名 '%s' 复制到剪切板中", filename))
			} else {
				m.SetMessage("数据加载中，稍后再试...")
			}
			m.ClearLastKey()
		case m.MatcheKeys(msg, m.KeyMap.CopyFilenameWithoutExt):
			logger.Infof("复制文件名称不含扩展")
			selectFile, err := m.GetSelectFile()
			if !m.IsLoadingFileList() && err == nil {
				filename := selectFile.GetFilename()
				// 获取文件名（包含扩展名）
				baseName := filepath.Base(filename)
				// 获取扩展名
				ext := filepath.Ext(filename)
				// 获取文件名（不包含扩展名）
				fileNameWithoutExt := baseName[:len(baseName)-len(ext)]
				clipboard.WriteAll(fileNameWithoutExt)
				m.SetMessage(fmt.Sprintf("文件名 '%s' 复制到剪切板中", fileNameWithoutExt))
			} else {
				m.SetMessage("数据加载中，稍后再试...")
			}
			m.ClearLastKey()
		default:
			// 监听不到组合键位才设置最后一个键位
			m.SetLastKey(msg)
		}
	}
	logger.Infof("记录最后的键位是 %v", m.lastKey)
	logger.Infof("BDPan Update time used %v ==================", time.Now().Sub(begin))
	return m, cmd
}

func (m *BDPan) View() string {
	begin := time.Now()
	logger.Infof("BDPan View begin ===========================")
	logger.Infof("Window size: %dx%d", m.width, m.height)

	midView := m.GetMidView()

	statusView := m.GetStatusView()

	messageView := m.GetMessageView()

	logger.Infof("BDPan View time used %v ====================", time.Now().Sub(begin))
	return lipgloss.JoinVertical(
		lipgloss.Top,
		midView,
		statusView,
		messageView,
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

func (m *BDPan) GetMessageView() string {
	commonStyle := lipgloss.
		NewStyle().
		Width(m.GetWidth() / 2)
	confirmView := commonStyle.
		Align(lipgloss.Left).
		Render("确认")
	messageView := commonStyle.
		Align(lipgloss.Right).
		Render(m.message)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		confirmView,
		messageView,
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
	height := m.GetHeight() - 1 - lipgloss.Height(m.GetMessageView())
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

// 设置消息
func (m *BDPan) SetMessage(msg string) {
	m.message = msg
}

func (m *BDPan) MessageIsNotNil() bool {
	return m.message != ""
}

func (m *BDPan) GetSelectFile() (*model.File, error) {
	if m.FileListModelIsNotNil() {
		return m.fileListModel.GetSelectFile()
	}
	return nil, fmt.Errorf("not found select file")
}

func (m *BDPan) SetLastKey(msg tea.KeyMsg) *BDPan {
	m.lastKey = &msg
	return m
}

func (m *BDPan) ClearLastKey() *BDPan {
	m.lastKey = nil
	return m
}

func (m *BDPan) MatcheKeys(msg tea.KeyMsg, b ...key.Binding) bool {
	curKey := msg.String()
	lastKey := ""
	if m.lastKey != nil {
		lastKey = m.lastKey.String()
	}
	combKey := lastKey + curKey
	for _, binding := range b {
		for _, k := range binding.Keys() {
			if binding.Enabled() {
				if combKey == k {
					return true
				}
				if curKey == k {
					return true
				}
			}
		}
	}
	return false
}

type KeyMap struct {
	Enter  key.Binding
	Back   key.Binding
	Delete key.Binding

	// 复制组合键位
	CopyPath               key.Binding
	CopyFilename           key.Binding
	CopyFilenameWithoutExt key.Binding
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
		// Delete: key.NewBinding(
		// key.WithKeys("d", "d"),
		// key.WithHelp("dd", "删除"),
		// ),
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
