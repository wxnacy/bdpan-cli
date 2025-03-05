package terminal

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/logger"
)

func NewTerminal(path string) *Terminal {
	return &Terminal{
		Path:        path,
		fileHandler: handler.GetFileHandler(),
		authHandler: handler.GetAuthHandler(),
	}
}

type Terminal struct {
	Path        string
	fileHandler *handler.FileHandler
	authHandler *handler.AuthHandler

	p *tea.Program
	m *BDPan
}

func (t *Terminal) Send(second int, send func() interface{}) {
	now := time.Now().Second()
	if now%second == 0 {
		// logger.Infof("监听时间并执行发送消息任务 %d", now)
		m := send()
		if m != nil {
			t.p.Send(m)
		}
	}
}

func (t *Terminal) refreshPan() {
	// 初始化 pan 信息
	if t.m.PanIsNil() {
		pan, err := t.authHandler.GetPan()
		if err != nil {
			panic(err)
		}
		t.p.Send(ChangePanMsg{Pan: pan})
	}
}

func (t *Terminal) refreshFiles() {
	// 刷新文件列表
	if t.m.IsLoadingFileList() || t.m.FileListModelIsNil() {
		files, err := t.fileHandler.GetFiles(t.m.Dir, 1)
		if err != nil {
			panic(err)
		}
		t.p.Send(ChangeFilesMsg{
			Files: files,
		})
	}
}

func (t *Terminal) listenRefreshFiles() {
	// 监听刷新文件列表
	for {
		t.refreshFiles()
	}
}

// func (t *Terminal) listenTasks() {
// // 监听任务
// for {
// time.Sleep(time.Duration(1) * time.Second)
// for _, task := range t.m.GetConfirmTasks() {
// logger.Infof("监听到任务 %v", task)
// // t.m.DoneTask(task)
// }
// }
// }

// func (t *Terminal) clearMessage() {
// // 清理消息
// for {
// time.Sleep(time.Duration(2) * time.Second)
// if t.m.MessageIsNotNil() {
// t.p.Send(ChangeMessageMsg{Message: ""})
// }
// }
// }

func (t *Terminal) refreshUser() {
	// 初始化 user 信息
	if t.m.UserIsNil() {
		user, err := t.authHandler.GetUser()
		if err != nil {
			panic(err)
		}
		t.p.Send(ChangeUserMsg{User: user})
	}
}

func (t *Terminal) Run() error {
	logger.Infof("")
	logger.Infof("BDPan Terminal begin ================================")
	m, err := NewBDPan(t.Path)
	if err != nil {
		return err
	}

	logger.Infof("View base left size %d", baseStyle.GetBorderLeftSize())

	p := tea.NewProgram(m, tea.WithAltScreen())
	t.p = p
	t.m = m

	// go t.refreshFiles()
	// go t.refreshPan()
	// go t.refreshUser()

	// go t.listenRefreshFiles()

	// go t.listenTasks()

	// go t.clearMessage()

	if _, err := p.Run(); err != nil {
		return err
	}
	logger.Infof("BDPan Terminal end   ================================")
	return nil
}
