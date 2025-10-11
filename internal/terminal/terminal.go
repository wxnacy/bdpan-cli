package terminal

import (
	"fmt"
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
	for {
		logger.Infof("BDPan Terminal begin ================================")
		m, err := NewBDPan(t.Path)
		if err != nil {
			return err
		}

		// If we have a previous model state from a loop, restore it.
		if t.m != nil {
			m.RestoreState(t.m)
		}

		p := tea.NewProgram(m, tea.WithAltScreen())
		t.p = p
		t.m = m // Store the current model for the next loop iteration

		if _, err := p.Run(); err != nil {
			// If Run returns an error, we should probably exit.
			return err
		}
		logger.Infof("BDPan Terminal end   =================================")

		// After Run() returns, check if we need to perform an action outside the TUI
		switch t.m.NextAction {
		case "batch-rename":
			files, ok := t.m.ActionPayload.([]*model.File)
			if !ok {
				// Set an error message for the next run
				t.m.message = "Error: Invalid payload for batch-rename"
				continue // Restart TUI
			}

			_, err := t.fileHandler.BatchRenameFiles(files)
			if err != nil {
				// Pass the error message to the next TUI run
				t.m.message = fmt.Sprintf("Rename failed: %v", err)
			} else {
				t.m.message = "Batch rename successful!"
			}
			// After the action, loop to restart the TUI
			continue
		default:
			// No special action, so exit the loop and the program
			return nil
		}
	}
}
