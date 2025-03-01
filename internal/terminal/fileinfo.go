package terminal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
)

type FileInfo struct {
	Model table.Model
}

func (m FileInfo) Init() tea.Cmd { return nil }

func (m *FileInfo) Update(msg tea.Msg) (*FileInfo, tea.Cmd) {
	var cmd tea.Cmd
	return m, cmd
}

func (m FileInfo) View() string {
	return baseStyle.Render(m.Model.View()) + "\n"
}

func NewFileInfo(f *model.File, height int) (*FileInfo, error) {
	valueW := 50
	columns := []table.Column{
		{Title: "字段", Width: 10},
		{Title: "详情", Width: valueW},
	}

	rows := make([]table.Row, 0)
	rows = append(rows, table.Row{
		"FSID",
		fmt.Sprintf("%d", f.FSID),
	})
	filename := fmt.Sprintf("%s %s", f.GetFileTypeEmoji(), f.GetFilename()) + "\n"
	// newfilename := ""
	// for _, s := range filename {
	// w, _ := lipgloss.Size(newfilename)
	// }
	nameW, nameH := lipgloss.Size(filename)
	logger.Infof("名称尺寸 %dx%d", nameW, nameH)
	rows = append(rows, table.Row{
		"文件名",
		// fmt.Sprintf("%s %s", f.GetFileTypeEmoji(), f.GetFilename()),
		"你好\nss",
	})
	rows = append(rows, table.Row{
		"大小",
		f.GetSize(),
	})
	rows = append(rows, table.Row{
		"类型",
		f.GetFileType(),
	})
	rows = append(rows, table.Row{
		"地址",
		f.Path,
	})
	rows = append(rows, table.Row{
		"创建时间",
		f.GetServerCTime(),
	})
	rows = append(rows, table.Row{
		"修改时间",
		f.GetServerMTime(),
	})

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	// s.Selected = s.Selected.
	// Foreground(lipgloss.Color("229")).
	// Background(lipgloss.Color("57")).
	// Bold(false)
	t.SetStyles(s)

	return &FileInfo{t}, nil
}
