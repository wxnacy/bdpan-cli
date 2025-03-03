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
	wtea "github.com/wxnacy/bdpan-cli/pkg/whitetea"
)

type RunTaskMsg struct {
	Task *Task
}

type GotoMsg struct {
	Dir string
}

type ChangeFilesMsg struct {
	Files []*model.File
}

type RefreshPanMsg struct{}
type RefreshUserMsg struct{}

type ChangePanMsg struct {
	Pan *model.Pan
}

type ChangeUserMsg struct {
	User *model.User
}

type MessageTimeoutMsg struct{}

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
		taskMap:  make(map[int]*Task, 0),
		// message:     "初始化",
		messageLifetime: time.Second,
		fileHandler:     handler.GetFileHandler(),
		authHandler:     handler.GetAuthHandler(),
		KeyMap:          DefaultKeyMap(),
	}, nil
}

type BDPan struct {
	Dir string

	// Data
	// useCache bool
	files    []*model.File
	filesMap map[string][]*model.File
	taskMap  map[int]*Task
	pan      *model.Pan
	user     *model.User

	// message
	message         string
	messageTimer    *time.Timer
	messageLifetime time.Duration

	// state
	// 改变窗口尺寸
	changeWindowSizeState bool

	KeyMap  KeyMap
	lastKey *tea.KeyMsg

	fileHandler *handler.FileHandler
	authHandler *handler.AuthHandler

	// Terminal
	width  int
	height int

	// filelist
	fileListModel     *FileList
	fileListViewState bool

	// quick
	quicks     []*model.Quick
	quickModel *wtea.List

	// confirm
	confirmModel *wtea.Confirm
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

	// m.quicks = []*model.Quick{
	// &model.Quick{

	// },
	// }

	logger.Infof("BDPan Init time used %v ====================", time.Now().Sub(begin))
	return tea.Batch(
		m.SendGoto(m.Dir),
		m.SendRefreshPan(),
		m.SendRefreshUser(),
	)
}

func (m *BDPan) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	begin := time.Now()
	logger.Infof("BDPan Update begin ===========================")
	var cmds []tea.Cmd
	var cmd tea.Cmd
	// var err error
	logger.Infof("Update by msg: %v", msg)

	flag, cmd := m.ListenCombKeyMsg(msg)
	cmds = append(cmds, cmd)
	if flag {
		// 监听到组合键位后的操作
		// 清理上一个key
		m.ClearLastKey()
	} else {
		// m.SetLastKey(msg)
		var isKeyMsg bool
		isKeyMsg, cmd = m.ListenKeyMsg(msg)
		if isKeyMsg {
			// 监听不到组合键位才设置最后一个键位
			m.SetLastKey(msg.(tea.KeyMsg))
		} else {
			_, cmd = m.ListenOtherMsg(msg)
		}
	}

	logger.Infof("记录最后的键位是 %v", m.lastKey)
	logger.Infof("BDPan Update time used %v ==================", time.Now().Sub(begin))
	return m, cmd
}

func (m *BDPan) ListenOtherMsg(msg tea.Msg) (bool, tea.Cmd) {
	var flag bool = true
	var cmds []tea.Cmd
	// var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// 更改尺寸后,重新获取模型
		// m.ChangeDir(m.Dir)
		m.changeWindowSizeState = true
	case GotoMsg:
		// 新获取文件列表
		files, err := m.fileHandler.GetFiles(msg.Dir, 1)
		if err != nil {
			return false, tea.Quit
		}
		m.Dir = msg.Dir
		m.SetFiles(files)
	case RunTaskMsg:
		// 运行任务
		t := msg.Task

		switch t.Type {
		case TypeDelete:
			_, err := m.fileHandler.DeleteFile(t.File.Path)
			cmds = append(cmds, m.DoneTask(t, err))
			if err == nil {
				// 删除成功后刷新目录
				cmds = append(cmds, m.SendGoto(m.Dir))
			}

		}
	case ChangeFilesMsg:
		// 异步加载文件列表
		m.SetFiles(msg.Files)
	case RefreshPanMsg:
		// 刷新网盘信息
		pan, err := m.authHandler.GetPan()
		if err != nil {
			return flag, tea.Quit
		}
		m.pan = pan
	case RefreshUserMsg:
		// 刷新用户信息
		user, err := m.authHandler.GetUser()
		if err != nil {
			return flag, tea.Quit
		}
		m.user = user
	case ChangePanMsg:
		// 异步加载 pan 信息
		m.pan = msg.Pan
	case ChangeUserMsg:
		// 异步加载 user 信息
		m.user = msg.User
	// case ChangeMessageMsg:
	// 接收信息
	// m.SetMessage(msg.Message)
	case MessageTimeoutMsg:
		// 消息过期
		m.message = ""
		if m.messageTimer != nil {
			m.messageTimer.Stop()
		}

	}
	return flag, tea.Batch(cmds...)
}

func (m *BDPan) ListenKeyMsg(msg tea.Msg) (bool, tea.Cmd) {
	flag := true
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Exit):
			// 退出程序
			return true, tea.Quit
		case m.fileListModel.Focused():
			// 光标聚焦在文件列表中
			// 先做原始修改操作
			if m.FileListModelIsNotNil() {
				m.fileListModel, cmd = m.fileListModel.Update(msg)
				cmds = append(cmds, cmd)
			}
			switch {
			case key.Matches(msg, m.KeyMap.Delete):
				// 删除
				m.fileListModel.Blur()
				if m.FileListModelIsNotNil() {
					f, err := m.GetSelectFile()
					if err != nil {
						return true, tea.Quit
					}
					task := m.AddDeleteTask(f)
					m.confirmModel = wtea.NewConfirm("确认删除？").
						Width(m.GetRightWidth()).
						SetData(task).
						Focus()
				}

			case key.Matches(msg, m.KeyMap.Refresh):
				// 刷新目录
				cmds = append(cmds, m.SendGoto(m.Dir))
			case key.Matches(msg, m.KeyMap.Back):
				// 返回目录
				cmds = append(cmds, m.Goto(filepath.Dir(m.Dir)))
			case key.Matches(msg, m.KeyMap.Enter):
				selectFile, err := m.fileListModel.GetSelectFile()
				if err != nil {
					return true, tea.Quit
				}
				if selectFile.IsDir() {
					cmds = append(cmds, m.Goto(selectFile.Path))
				}
			}
		case m.confirmModel.Focused():
			// 光标聚焦在确认框中
			if m.confirmModel != nil {
				m.confirmModel, cmd = m.confirmModel.Update(msg)
				cmds = append(cmds, cmd)
				if !m.confirmModel.Focused() {
					m.fileListModel.Focus()
				}

				// 执行任务
				if m.confirmModel.GetValue() {
					t := m.confirmModel.GetData().(*Task)
					cmds = append(cmds, m.SendRunTask(t))
				}
			}
		}
	default:
		flag = false
	}
	return flag, tea.Batch(cmds...)
}

func (m *BDPan) ListenCombKeyMsg(msg tea.Msg) (bool, tea.Cmd) {
	// 监听两个键位的组合键位
	var cmds []tea.Cmd
	var cmd tea.Cmd
	flag := true
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.MatcheKeys(msg, m.KeyMap.GetCombKeys()...) {
			return false, cmd
		}
		switch {
		case m.MatcheKeys(msg, m.KeyMap.GetCopyKeys()...):
			// 监听复制键位
			var copyText string
			switch {
			case m.MatcheKeys(msg, m.KeyMap.CopyPath):
				logger.Infoln(m.KeyMap.CopyPath.Help().Desc)
				selectFile, err := m.GetSelectFile()
				if !m.IsLoadingFileList() && err == nil {
					copyText = selectFile.Path
				}
			case m.MatcheKeys(msg, m.KeyMap.CopyDir):
				logger.Infoln(m.KeyMap.CopyDir.Help().Desc)
				copyText = m.Dir
			case m.MatcheKeys(msg, m.KeyMap.CopyFilename):
				logger.Infoln(m.KeyMap.CopyFilename.Help().Desc)
				selectFile, err := m.GetSelectFile()
				if !m.IsLoadingFileList() && err == nil {
					copyText = selectFile.GetFilename()
				}
			case m.MatcheKeys(msg, m.KeyMap.CopyFilenameWithoutExt):
				logger.Infoln(m.KeyMap.CopyFilenameWithoutExt.Help().Desc)
				selectFile, err := m.GetSelectFile()
				if !m.IsLoadingFileList() && err == nil {
					filename := selectFile.GetFilename()
					// 获取文件名（包含扩展名）
					baseName := filepath.Base(filename)
					// 获取扩展名
					ext := filepath.Ext(filename)
					// 获取文件名（不包含扩展名）
					copyText = baseName[:len(baseName)-len(ext)]
				}
			}
			if copyText != "" {
				// m.SetClipboardMessage(copyText)
				clipboard.WriteAll(copyText)
				cmds = append(cmds, m.SendClipboardMessage(copyText))
			}
			// if m.IsLoadingFileList() {
			// m.SetLoadingMessage()
			// }
		case m.MatcheKeys(msg, m.KeyMap.GetGotoKeys()...):
			// 监听 Goto 键位
			switch {
			case m.MatcheKeys(msg, m.KeyMap.GotoRoot):
				cmds = append(cmds, m.SendGoto("/"))
			}

		}
	default:
		flag = false
	}
	return flag, tea.Batch(cmds...)
}

func (m *BDPan) View() string {
	begin := time.Now()
	logger.Infof("BDPan View begin ===========================")
	logger.Infof("Window size: %dx%d", m.width, m.height)

	midView := m.GetMidView()

	statusView := m.GetStatusView()

	messageView := m.GetMessageView()

	views := []string{
		midView,
	}

	views = append(
		views,
		statusView,
		messageView,
	)

	logger.Infof("BDPan View time used %v ====================", time.Now().Sub(begin))
	view := lipgloss.JoinVertical(
		lipgloss.Top,
		views...,
	)
	return view
}

func (m *BDPan) GetFileListView() string {
	if !m.IsLoadingFileList() && m.FileListModelIsNotNil() {
		// 尺寸改变重新加载
		if m.changeWindowSizeState {
			m.fileListModel = m.NewFileList(m.files)
			m.changeWindowSizeState = false
		}
		fileListView := m.fileListModel.View()
		logger.Infof("FileListView height %d", lipgloss.Height(fileListView))
		return fileListView
	} else {
		return m.NewFileList(nil).View()
	}
}

func (m *BDPan) GetConfirmView() string {
	// return baseStyle.Width(60).Render(m.confirmModel.View())
	return m.confirmModel.View()
}

func (m *BDPan) GetMidView() string {

	// filelist
	centerView := m.GetFileListView()

	// fileinfo
	fileinfoView := m.GetFileInfoView(nil)
	if m.FileListModelIsNotNil() {
		f, err := m.fileListModel.GetSelectFile()
		if err != nil {
			tea.Quit()
			return ""
		}
		fileinfoView = m.GetFileInfoView(f)
	}

	rightViews := []string{fileinfoView}
	if m.ConfirmFocused() {
		rightViews = append(rightViews, m.GetConfirmView())
	}

	rightView := lipgloss.JoinVertical(
		lipgloss.Top,
		rightViews...,
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		centerView,
		rightView,
	)
}

func (m *BDPan) GetMessageView() string {
	commonStyle := lipgloss.
		NewStyle().
		Width(m.GetWidth())
		// Width(m.GetWidth() / 2)
	// confirmView := commonStyle.
	// Align(lipgloss.Left).
	// Render("确认")
	messageView := commonStyle.
		Align(lipgloss.Left).
		Render(m.message)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		// confirmView,
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
	if m.ConfirmFocused() {
		lastH -= lipgloss.Height(m.GetConfirmView())
	}

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

func (m *BDPan) GetRightWidth() int {
	return 60
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

func (m *BDPan) ConfirmFocused() bool {
	return m.confirmModel != nil && m.confirmModel.Focused()
}

func (m *BDPan) AddDeleteTask(f *model.File) *Task {
	task := NewTask(TypeDelete, f)
	_, exists := m.taskMap[task.ID]
	if exists {
		m.SetSomeTaskMessage()
	} else {
		m.taskMap[task.ID] = task
	}
	return task
}

func (m *BDPan) GetConfirmTasks() []*Task {
	tasks := make([]*Task, 0)
	for _, t := range m.taskMap {
		if t.IsConfirm {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

func (m *BDPan) DoneTask(t *Task, err error) tea.Cmd {
	t.Status = StatusSuccess
	if err != nil {
		t.Status = StatusFailed
		t.err = err
	}
	delete(m.taskMap, t.ID)
	// m.SetMessage(t.String())
	return m.SendMessage(t.String())
}

// 改变显示的目录
func (m *BDPan) Goto(dir string) tea.Cmd {
	if m.IsLoadingFileList() {
		// m.SetLoadingMessage()
		return m.SendLoadingMessage()
	}

	m.Dir = dir
	files := m.getFiles(m.Dir)
	if files == nil {
		// 没有缓存时打开 Loading
		m.EnableLoadingFileList()
		return m.SendGoto(dir)
	} else {
		m.files = files
		m.fileListModel = m.NewFileList(m.files)
	}
	return nil
}

// 设置消息
func (m *BDPan) SetMessage(msg string, args ...interface{}) {
	if len(args) == 0 {
		m.message = msg
	} else {
		m.message = fmt.Sprintf(msg, args...)
	}
	logger.Infoln(m.message)
}

func (m *BDPan) SetSomeTaskMessage() {
	m.SetMessage("相同任务已添加")
}

func (m *BDPan) MessageIsNotNil() bool {
	return m.message != ""
}

func (m *BDPan) SendLoadingMessage() tea.Cmd {
	return m.SendMessage("数据加载中，稍后再试...")
}

func (m *BDPan) SendClipboardMessage(msg string) tea.Cmd {
	return m.SendMessage("'%s' 复制到剪切板中", msg)
}

func (m *BDPan) SendMessage(msg string, a ...any) tea.Cmd {
	s := msg
	if len(a) > 0 {
		s = fmt.Sprintf(msg, a...)
	}
	m.message = s
	if m.messageTimer != nil {
		m.messageTimer.Stop()
	}
	m.messageTimer = time.NewTimer(m.messageLifetime)
	return func() tea.Msg {
		<-m.messageTimer.C
		return MessageTimeoutMsg{}
	}
}

func (m *BDPan) SendRunTask(t *Task) tea.Cmd {
	t.Status = StatusRunning
	return tea.Batch(
		func() tea.Msg {
			return RunTaskMsg{
				Task: t,
			}
		},
		m.SendMessage(t.String()),
	)
}

func (m *BDPan) SendGoto(dir string) tea.Cmd {
	m.EnableLoadingFileList()
	return func() tea.Msg {
		return GotoMsg{
			Dir: dir,
		}
	}
}

func (m *BDPan) SendRefreshPan() tea.Cmd {
	return func() tea.Msg {
		return RefreshPanMsg{}
	}
}

func (m *BDPan) SendRefreshUser() tea.Cmd {
	return func() tea.Msg {
		return RefreshUserMsg{}
	}
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
