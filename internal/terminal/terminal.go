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

func (t *Terminal) Send(second int, send func() any) {
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

	logger.Infof("View base left size %d", baseStyle.GetBorderLeftSize())

	p := tea.NewProgram(m, tea.WithAltScreen())
	t.p = p
	t.m = m

	if _, err := p.Run(); err != nil {
		return err
	}
	logger.Infof("BDPan Terminal end   ================================")
	return nil
}
