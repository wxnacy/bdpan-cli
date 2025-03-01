package terminal

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
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
	logger.Infof("BDPan state %v", m.viewState)
	go func() {
		for {
			logger.Infof("开始异步刷新信息...")
			if !m.viewState {
				files, err := t.fileHandler.GetFiles(m.Dir, 1)
				if err != nil {
					panic(err)
				}
				pan, err := t.authHandler.GetPan()
				if err != nil {
					panic(err)
				}
				userInfo, err := t.authHandler.GetUserInfo()
				if err != nil {
					panic(err)
				}
				user := model.NewUser(userInfo)
				p.Send(NewInitMsg(
					files,
					pan,
					user,
				))
				continue
			}
			time.Sleep(time.Duration(10) * time.Second)
			logger.Infof("BDPan state %v", m.viewState)

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
