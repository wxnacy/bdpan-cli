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
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
	wtea "github.com/wxnacy/bdpan-cli/pkg/whitetea"
)

type RunTaskMsg struct {
	Task *Task
}
type ShowConfirmMsg struct {
	Title string
	Data  wtea.ExtData
	Model wtea.Model
}

type GotoMsg struct {
	Dir string
}

type RefreshPanMsg struct{}
type RefreshUserMsg struct{}
type RefreshQuickMsg struct{}
type DeleteQuickMsg struct {
	Quick *model.Quick
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

type MessageTimeoutMsg struct{}

func NewBDPan(dir string) (*BDPan, error) {
	begin := time.Now()
	quicks := []*model.Quick{
		{
			Filename: "我的应用数据",
			Path:     "/apps",
			Key:      "1",
		},
	}

	item := &BDPan{
		Dir:             dir,
		files:           make([]*model.File, 0),
		filesMap:        make(map[string][]*model.File, 0),
		fileCursorMap:   make(map[string]int, 0),
		taskMap:         make(map[int]*Task, 0),
		quicks:          quicks,
		messageLifetime: time.Second,
		fileHandler:     handler.GetFileHandler(),
		authHandler:     handler.GetAuthHandler(),
		KeyMap:          DefaultKeyMap(),
	}
	logger.Infof("NewDBPan time used %v", time.Now().Sub(begin))
	return item, nil
}

type BDPan struct {
	Dir string

	// Data
	taskMap map[int]*Task
	pan     *model.Pan
	user    *model.User

	// message
	message         string
	messageTimer    *time.Timer
	messageLifetime time.Duration // 消息生命周期

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
	files             []*model.File
	filesMap          map[string][]*model.File // 文件列表缓存
	fileCursorMap     map[string]int           // 文件光标选择记录
	fileListModel     *FileList
	fileListViewState bool

	// quick
	quicks     []*model.Quick
	quickKeys  []key.Binding
	quickModel *Quick

	// confirm
	confirmModel *wtea.Confirm
}

func (m *BDPan) GetWidth() int {
	var w = 100
	if m.width > 0 {
		w = m.width
	}
	// logger.Infof("View Window Width %d", w)
	return w
}

func (m *BDPan) GetHeight() int {
	var h = 20
	if m.height > 0 {
		h = m.height
	}
	// logger.Infof("View Window Height %d", h)
	return h
}

func (m *BDPan) getFiles(dir string) []*model.File {
	files, _ := m.filesMap[dir]
	return files
}

func (m *BDPan) SetFiles(files []*model.File) *BDPan {
	m.files = files
	m.filesMap[m.Dir] = files
	m.fileListModel = m.NewFileList(m.files)
	m.DisableLoadingFileList()
	m.RefreshQuickSelect()
	// 设置光标位置
	m.fileListModel.Cursor(m.GetFileListCursor())
	return m
}

func (m *BDPan) GetQuickKeys() []key.Binding {
	return m.quickKeys
}
func (m *BDPan) GetQuickByKeyStr(k string) *model.Quick {
	for _, v := range m.quicks {
		if k == v.Key {
			return v
		}
	}
	return nil
}
func (m *BDPan) GetQuickByPath(p string) *model.Quick {
	for _, v := range m.quicks {
		if p == v.Path {
			return v
		}
	}
	return nil
}
func (m *BDPan) SetQuicks(q []*model.Quick) {
	if q != nil && len(q) > 0 {
		m.quicks = q
	}
	keys := make([]key.Binding, 0)
	for _, quick := range q {
		if quick.Key != "" {
			k := "g" + quick.Key
			keys = append(keys, key.NewBinding(
				key.WithKeys(k),
				key.WithHelp(k, fmt.Sprintf("Go to the %s", quick.Path)),
			))
		}
	}
	m.quickKeys = keys

	var focused bool
	if m.quickModel != nil {
		focused = m.quickModel.Focused()
	}
	m.quickModel = NewQuick("快速访问", m.quicks, baseStyle).
		Width(m.GetLeftWidth()).
		Height(m.GetMidHeight())
	if focused {
		m.quickModel.Focus()
	}
}
func (m *BDPan) RefreshQuickSelect() *BDPan {
	// 定位快速访问
	for i, v := range m.quicks {
		if v.Path == m.Dir {
			m.quickModel.Select(i)
		}
	}
	return m
}

func (m *BDPan) Init() tea.Cmd {
	begin := time.Now()
	logger.Infof("BDPan Init begin ============================")

	m.SetQuicks(m.quicks)

	m.confirmModel = wtea.NewConfirm("", baseFocusStyle)

	logger.Infof("BDPan Init time used %v ====================", time.Now().Sub(begin))
	return tea.Batch(
		m.SendRefreshQuick(),
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
	var err error
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// 更改尺寸后,重新获取模型
		// m.ChangeDir(m.Dir)
		m.changeWindowSizeState = true
		// 修改quick size
		m.quickModel.
			Width(m.GetLeftWidth()).
			Height(m.GetMidHeight())
		m.quickModel, cmd = m.quickModel.Update(msg)
		cmds = append(cmds, cmd)
	case GotoMsg:
		// 新获取文件列表
		files, err := m.fileHandler.GetFilesAndSave(msg.Dir, 1)
		if err != nil {
			return false, tea.Quit
		}
		m.Dir = msg.Dir
		m.SetFiles(files)
	case ShowConfirmMsg:
		// 展示确认框
		m.confirmModel.
			Title(msg.Title).
			Width(m.GetRightWidth()).
			Value(false).
			Data(msg.Data).
			FromModel(msg.Model).
			Focus()
	case DeleteQuickMsg:
		// 删除快捷方式
		model.DeleteById[model.Quick](int(msg.Quick.ID))
		cmds = append(cmds, m.SendRefreshQuick(), m.SendMessage("删除快速访问 %s", msg.Quick.Path))
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
		case TypeDownload:
			req := dto.NewDownloadReq()
			req.Path = t.File.Path
			err = m.fileHandler.CmdDownload(req)
			cmds = append(cmds, m.DoneTask(t, err))
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
	case RefreshQuickMsg:
		// 刷新快速访问
		quicks := model.FindItems[model.Quick]()
		m.SetQuicks(quicks)
	case ChangePanMsg:
		// 异步加载 pan 信息
		m.pan = msg.Pan
	case ChangeUserMsg:
		// 异步加载 user 信息
		m.user = msg.User
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
		case key.Matches(msg, m.KeyMap.MovePaneLeft):
			// 向左移动面板
			if m.fileListModel.Focused() {
				m.quickModel.Focus()
				m.fileListModel.Blur()
			}
		case key.Matches(msg, m.KeyMap.MovePaneRight):
			// 向右移动面板
			if m.quickModel.Focused() {
				m.FileListFocus()
			}
		case m.fileListModel.Focused():
			// 光标聚焦在文件列表中
			if m.FileListModelIsNotNil() {
				// 先做原始修改操作
				m.fileListModel, cmd = m.fileListModel.Update(msg)
				cmds = append(cmds, cmd)

				// 记录光标位置
				m.fileCursorMap[m.Dir] = m.fileListModel.GetCursor()
			}
			switch {
			case key.Matches(msg, m.KeyMap.Delete):
				// 删除
				// m.fileListModel.Blur()
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
			case key.Matches(msg, m.KeyMap.Refresh):
				// 刷新目录
				// 盘信息
				cmds = append(cmds, m.SendGoto(m.Dir), m.SendRefreshPan())
			case key.Matches(msg, m.KeyMap.Back):
				// 返回目录
				cmds = append(cmds, m.Goto(filepath.Dir(m.Dir)))
			case key.Matches(msg, m.KeyMap.Enter):
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
			case key.Matches(msg, m.fileListModel.GetKeyMap().AddQuick):
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
							q := f.ToQuick()
							model.Save(q)
							cmds = append(cmds, m.SendRefreshQuick(), m.SendMessage("%s 已添加快速访问", f.Path))
						}

					} else {
						cmds = append(cmds, m.SendMessage("文件不支持添加快速访问"))
					}
				}
			}

		case m.ConfirmFocused():
			// 光标聚焦在确认框中
			if m.confirmModel != nil {
				m.confirmModel, cmd = m.confirmModel.Update(msg)
				cmds = append(cmds, cmd)
				if !m.confirmModel.Focused() {
					fromModel := m.confirmModel.GetFromModel()
					if fromModel != nil {
						fromModel.Focus()
					}
				}

				// 执行任务
				if m.confirmModel.GetValue() {
					switch d := m.confirmModel.GetData().(type) {
					case *Task:
						cmds = append(cmds, m.SendRunTask(d))
					case tea.Cmd:
						cmds = append(cmds, d)
					}
				}
			}
		case m.quickModel.Focused():
			// 聚焦在快速访问
			m.quickModel, cmd = m.quickModel.Update(msg)
			switch {
			case key.Matches(msg, m.quickModel.GetKeyMap().Enter):
				m.quickModel.Blur()
				m.fileListModel.Focus()
				q := m.quickModel.GetSelect()
				cmds = append(cmds, m.SendGoto(q.Path))
			case key.Matches(msg, m.quickModel.GetKeyMap().Delete):
				q := m.quickModel.GetSelect()
				cmd = m.SendShowConfirm(
					fmt.Sprintf("确认删除快速访问 %s?", q.Path),
					m.SendDeleteQuick(q),
					m.quickModel,
				)
				cmds = append(cmds, cmd)
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
		if !m.MatcheKeys(msg, m.KeyMap.GetCombKeys(m.GetQuickKeys())...) {
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
				clipboard.WriteAll(copyText)
				cmds = append(cmds, m.SendClipboardMessage(copyText))
			}
			if m.IsLoadingFileList() {
				cmds = append(cmds, m.SendLoadingMessage())
			}
		case m.MatcheKeys(msg, m.KeyMap.GetGotoKeys(m.GetQuickKeys())...):
			// 监听 Goto 键位
			var gotoDir string
			switch {
			case m.MatcheKeys(msg, m.KeyMap.GotoRoot):
				gotoDir = "/"
			default:
				q := m.GetQuickByKeyStr(msg.String())
				if q != nil {
					gotoDir = q.Path
				}
			}
			if gotoDir != "" {
				cmds = append(
					cmds,
					m.Goto(gotoDir),
					m.SendMessage("快速跳转 %s", gotoDir),
				)
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

// 获取文件光标
func (m *BDPan) GetFileListCursor() int {
	i, exists := m.fileCursorMap[m.Dir]
	if exists {
		return i
	} else {
		return 0
	}
}

func (m *BDPan) GetFileListView() string {
	if !m.IsLoadingFileList() && m.FileListModelIsNotNil() {
		// 尺寸改变重新加载
		if m.changeWindowSizeState {
			m.fileListModel = m.NewFileList(m.files)
			m.changeWindowSizeState = false
		}
		return m.fileListModel.View()
	} else {
		return m.NewFileList(nil).View()
	}
}

func (m *BDPan) GetDirView() string {
	return baseStyle.Width(m.GetMidWidth()-2).Render("当前目录", m.Dir)
}

func (m *BDPan) GetCenterView() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.GetDirView(),
		m.GetFileListView(),
	)
}

func (m *BDPan) GetMidView() string {

	// quick
	leftView := m.quickModel.View()

	// filelist
	centerView := m.GetCenterView()

	// fileinfo
	fileinfoView := m.GetFileInfoView(nil)
	if m.FileListModelIsNotNil() && m.FilesIsNotNil() {
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
		leftView,
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
	if m.FileListModelIsNotNil() && m.FilesIsNotNil() {
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

	var leftW = 10
	// 2 是边框的长度
	var rightW = m.GetRightWidth() - leftW - 2

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
		Width(leftW)

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
		Width(rightW)

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
		// } else {
		// lines = append(lines, lipgloss.JoinHorizontal(
		// lipgloss.Top,
		// leftStyle.Render(""),
		// rightStyle.Render("数据加载中..."),
		// ))
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

	var view = lipgloss.JoinVertical(
		lipgloss.Top,
		lines...,
	)
	var viewW, viewH int
	viewW, viewH = lipgloss.Size(view)
	logger.Infof("FileInfoView Full Size %dx%d", viewW, viewH)
	return view

}

func (m *BDPan) GetMidWidth() int {
	w := m.GetWidth() / 2
	logger.Infof("View GetMidWidth %d", w)
	return w
}

func (m *BDPan) GetMidHeight() int {
	height := m.GetHeight() - 1 - lipgloss.Height(m.GetMessageView())
	logger.Infof("View GetMidHeight %d", height)
	return height
}

func (m *BDPan) GetLeftWidth() int {
	var w = int(float32(m.GetWidth()) * 0.16)
	logger.Infof("View GetLeftWidth %d", w)
	return w
}

func (m *BDPan) GetRightWidth() int {
	// 1 是因为 Left 会多出一个长度
	var w = m.GetWidth() - m.GetMidWidth() - m.GetLeftWidth() - 1
	logger.Infof("View GetRightWidth %d", w)
	return w
}

func (m *BDPan) NewFileList(files []*model.File) *FileList {
	var focused bool = true
	if m.FileListModelIsNotNil() {
		focused = m.fileListModel.Focused()
	}
	var h = m.GetMidHeight() - lipgloss.Height(m.GetDirView())
	model := NewFileList(files, m.GetMidWidth(), h)
	// 检查之前是否聚焦
	if !focused {
		model.Blur()
	}
	return model
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

func (m *BDPan) FilesIsNotNil() bool {
	return m.files != nil && len(m.files) > 0
}

func (m *BDPan) CanSelectFile() bool {
	return m.FilesIsNotNil() && m.FileListModelIsNotNil()
}

func (m *BDPan) AddFileTask(f *model.File, t TaskType) *Task {
	task := NewTask(t, f)
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
func (m *BDPan) ConfirmFocused() bool {
	return m.confirmModel != nil && m.confirmModel.Focused()
}
func (m *BDPan) GetConfirmView() string {
	return m.confirmModel.View()
}

func (m *BDPan) FileListFocus() *BDPan {
	m.fileListModel.Focus()
	m.quickModel.Blur()
	if m.confirmModel != nil {
		m.confirmModel.Blur()
	}
	return m
}

// func (m *BDPan) SetConfirmModel(title string, task *Task) *BDPan {
// m.confirmModel = wtea.NewConfirm(title, baseFocusStyle).
// Width(m.GetRightWidth()).
// Data(task).
// Focus()
// return m
// }

func (m *BDPan) DoneTask(t *Task, err error) tea.Cmd {
	t.Status = StatusSuccess
	if err != nil {
		t.Status = StatusFailed
		t.err = err
	}
	delete(m.taskMap, t.ID)
	return m.SendMessage(t.String())
}

// 改变显示的目录
func (m *BDPan) Goto(dir string) tea.Cmd {
	if m.IsLoadingFileList() {
		return m.SendLoadingMessage()
	}

	m.Dir = dir
	files := m.getFiles(m.Dir)
	if files == nil {
		// 没有缓存时打开 Loading
		m.EnableLoadingFileList()
		return m.SendGoto(dir)
	} else {
		m.SetFiles(files)
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

func (m *BDPan) SendMsg(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

// 发送显示确认框消息
// title 展示标题
// data 确认框携带的额外信息
// fromModel 从哪个模型跳转的，方便返回聚焦
func (m *BDPan) SendShowConfirm(
	title string,
	data wtea.ExtData,
	fromModel wtea.Model,
) tea.Cmd {
	m.fileListModel.Blur()
	m.quickModel.Blur()
	return func() tea.Msg {
		return ShowConfirmMsg{
			Title: title,
			Data:  data,
			Model: fromModel,
		}
	}
}

// 发送跳转目录的命令
// 实时获取最新结果
func (m *BDPan) SendGoto(dir string) tea.Cmd {
	m.EnableLoadingFileList()
	return func() tea.Msg {
		return GotoMsg{
			Dir: dir,
		}
	}
}

func (m *BDPan) SendDeleteQuick(q *model.Quick) tea.Cmd {
	return func() tea.Msg {
		return DeleteQuickMsg{Quick: q}
	}
}

func (m *BDPan) SendRefreshQuick() tea.Cmd {
	return func() tea.Msg {
		return RefreshQuickMsg{}
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
