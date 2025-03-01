package terminal

import (
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
}

func (t *Terminal) Run() error {
	logger.Infof("")
	logger.Infof("BDPan Terminal begin ================================")
	m, err := NewBDPan(t.Path)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	// logger.Infof("BDPan state %v", m.viewState)
	go func() {
		for {
			// logger.Infof("开始异步刷新信息...")
			// if !m.viewState {
			// files, err := t.fileHandler.GetFiles(m.Dir, 1)
			// if err != nil {
			// panic(err)
			// }
			// pan, err := t.authHandler.GetPan()
			// if err != nil {
			// panic(err)
			// }
			// user, err := t.authHandler.GetUser()
			// if err != nil {
			// panic(err)
			// }
			// p.Send(NewInitMsg(
			// files,
			// pan,
			// user,
			// ))
			// continue
			// }
			// 刷新文件列表
			if m.IsLoadingFileList() || m.FileListModelIsNil() {
				files, err := t.fileHandler.GetFiles(m.Dir, 1)
				if err != nil {
					panic(err)
				}
				p.Send(ChangeFilesMsg{
					Files: files,
				})
				continue
			}
			// 初始化 pan 信息
			if m.PanIsNil() {
				pan, err := t.authHandler.GetPan()
				if err != nil {
					panic(err)
				}
				p.Send(ChangePanMsg{Pan: pan})
				continue
			}
			// 初始化 user 信息
			if m.UserIsNil() {
				user, err := t.authHandler.GetUser()
				if err != nil {
					panic(err)
				}
				p.Send(ChangeUserMsg{User: user})
				continue
			}
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
