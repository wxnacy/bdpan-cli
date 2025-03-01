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
}

func (t *Terminal) Run() error {
	m, err := NewBDPan(t.Path)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	go func() {
		for {
			time.Sleep(time.Duration(10) * time.Second)
			logger.Infof("开始异步刷新信息...")

			panInfo, _ := t.authHandler.GetPanInfo()
			p.Send(panInfo)
		}
	}()

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
