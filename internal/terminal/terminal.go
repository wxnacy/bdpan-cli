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
	p           *tea.Program
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

func (t *Terminal) Run() error {
	logger.Infof("")
	logger.Infof("BDPan Terminal begin ================================")
	m, err := NewBDPan(t.Path)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	t.p = p
	// logger.Infof("BDPan state %v", m.viewState)
	go func() {
		for {
			// 刷新文件列表
			if m.IsLoadingFileList() || m.FileListModelIsNil() {
				files, err := t.fileHandler.GetFiles(m.Dir, 1)
				if err != nil {
					panic(err)
				}
				p.Send(ChangeFilesMsg{
					Files: files,
				})
			}
			// 初始化 pan 信息
			if m.PanIsNil() {
				pan, err := t.authHandler.GetPan()
				if err != nil {
					panic(err)
				}
				p.Send(ChangePanMsg{Pan: pan})
			}
			// 初始化 user 信息
			if m.UserIsNil() {
				user, err := t.authHandler.GetUser()
				if err != nil {
					panic(err)
				}
				p.Send(ChangeUserMsg{User: user})
			}

			// 2 秒执行一次
			// if time.Now().Second()%2 == 0 {
			// if m.MessageIsNotNil() {
			// p.Send(ChangeMessageMsg{Message: ""})
			// }
			// }
			t.Send(5, func() interface{} {
				if m.MessageIsNotNil() {
					return ChangeMessageMsg{Message: ""}
				}
				return nil
			})
			// time.Sleep(time.Duration(10) * time.Second)
			// logger.Infof("BDPan state %v", m.viewState)

			// panInfo, _ := t.authHandler.GetPanInfo()
			// p.Send(panInfo)
		}
	}()

	if _, err := p.Run(); err != nil {
		return err
	}
	logger.Infof("BDPan Terminal end   ================================")
	return nil
}
