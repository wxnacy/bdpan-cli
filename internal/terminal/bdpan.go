package terminal

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
	"github.com/wxnacy/go-bdpan"
)

func NewInitMsg(
	files []*model.File,
	pan *model.Pan,
	user *model.User,
) *InitMsg {
	return &InitMsg{
		files: files,
		pan:   pan,
		user:  user,
	}
}

type InitMsg struct {
	files []*model.File
	pan   *model.Pan
	user  *model.User
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
		fileHandler: handler.GetFileHandler(),
		authHandler: handler.GetAuthHandler(),
		KeyMap:      DefaultKeyMap(),
	}, nil
}

type BDPan struct {
	Dir string

	// Data
	// useCache bool
	files     []*model.File
	filesMap  map[string][]*model.File
	pan       *model.Pan
	userInfo  *bdpan.GetUserInfoRes
	KeyMap    KeyMap
	viewState bool

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

func (m *BDPan) getFileList(dir string) ([]*model.File, error) {
	files, ok := m.filesMap[dir]
	if ok {
		logger.Infof("从缓存中读取文件列表")
		return files, nil
	} else {
		logger.Infof("从网络中读取文件列表")
		files, err := m.fileHandler.GetFiles(m.Dir, 1)
		if err != nil {
			return nil, err
		}
		m.filesMap[dir] = files
		return files, nil
	}
}

func (m *BDPan) initFileList(dir string) tea.Cmd {
	var err error
	m.Dir = dir
	m.files, err = m.getFileList(m.Dir)
	if err != nil {
		return tea.Quit
	}

	m.fileListModel, err = NewFileList(m.files, m.GetWidth()/2, m.GetMidHeight())
	if err != nil {
		return tea.Quit
	}
	return nil
}

func (m *BDPan) Init() tea.Cmd {
	begin := time.Now()
	logger.Infof("BDPan Init begin ============================")
	logger.Infof("Window size: %dx%d", m.width, m.height)

	var err error
	m.pan, err = m.authHandler.GetPan()
	if err != nil {
		return tea.Quit
	}

	m.userInfo, err = m.authHandler.GetUserInfo()
	if err != nil {
		return tea.Quit
	}

	cmd := m.initFileList(m.Dir)
	if cmd != nil {
		return cmd
	}
	m.viewState = true
	logger.Infof("BDPan Init time used %v ====================", time.Now().Sub(begin))
	return tea.SetWindowTitle("bdpan")
}

func (m *BDPan) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger.Infof("BDPan Update =================================")
	var cmd tea.Cmd
	var err error
	logger.Infof("Update by msg: %v", msg)

	// 先做原始修改操作
	m.fileListModel, cmd = m.fileListModel.Update(msg)

	switch msg := msg.(type) {
	// case *bdpan.GetPanInfoRes:
	// m.panInfo = msg
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// 更改尺寸后,重新获取模型
		m.fileListModel, err = NewFileList(m.files, m.GetWidth()/2, m.GetMidHeight())
		if err != nil {
			tea.Quit()
			return m, tea.Quit
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Back):
			_cmd := m.initFileList(filepath.Dir(m.Dir))
			if _cmd != nil {
				return m, _cmd
			}
		case key.Matches(msg, m.KeyMap.Enter):
			selectFile, err := m.fileListModel.GetSelectFile()
			if err != nil {
				tea.Quit()
				return m, tea.Quit
			}
			if selectFile.IsDir() {
				_cmd := m.initFileList(selectFile.Path)
				if _cmd != nil {
					return m, _cmd
				}
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

	if !m.viewState {
		return "界面初始化..."
	}

	midView := m.GetMidView()

	statusView := m.GetStatusView()
	return lipgloss.JoinVertical(
		lipgloss.Top,
		midView,
		statusView,
	)
}

func (m *BDPan) GetMidCenterView() string {
	fileListView := m.fileListModel.View()
	logger.Infof("FileListView height %d", lipgloss.Height(fileListView))
	return fileListView
}

func (m *BDPan) GetMidView() string {

	// filelist
	centerView := m.fileListModel.View()
	logger.Infof("Mid center view height %d", lipgloss.Height(centerView))

	// fileinfo
	f, err := m.fileListModel.GetSelectFile()
	if err != nil {
		tea.Quit()
		return ""
	}
	rightView := m.GetFileInfoView(f)

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

	statusKey := statusStyle.Render(fmt.Sprintf(
		"容量 %s/%s",
		m.pan.GetUsedStr(),
		m.pan.GetTotalStr(),
	))
	encoding := encodingStyle.Render(m.userInfo.GetNetdiskName())
	fishCake := fishCakeStyle.Render(m.userInfo.GetVipName())

	f, err := m.fileListModel.GetSelectFile()
	if err != nil {
		tea.Quit()
		return ""
	}
	fileLineText := fmt.Sprintf("%s", f.Path)
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

func encodeImage(path string) string {
	// 读取图片文件内容
	file, _ := os.Open(path)
	defer file.Close()
	img, _, _ := image.Decode(file)

	// 2. 定义字符映射表（按灰度排序）
	chars := []string{"@", "#", "S", "%", "?", "*", "+", ";", ":", ",", "."}

	// 3. 遍历像素生成字符画
	bounds := img.Bounds()
	res := ""
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 { // 纵向降采样
		var line string
		for x := bounds.Min.X; x < bounds.Max.X; x += 1 {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			index := int(gray / 255 * float64(len(chars)-1))
			line += chars[index]
		}
		res += line
	}
	return res
}

func (m *BDPan) GetMidHeight() int {
	height := m.GetHeight() - 2
	logger.Infof("GetMidHeight %d", height)
	return height
}

type KeyMap struct {
	Enter key.Binding
	Back  key.Binding
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
	}
}
