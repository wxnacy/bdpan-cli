package terminal

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type FileList struct {
	model table.Model
	files []*model.File
}

func (m *FileList) GetSelectFile() (*model.File, error) {
	id := m.model.SelectedRow()[0]
	fsid, err := strconv.Atoi(id)
	logger.Infof("select file fsid %d", fsid)
	if err != nil {
		return nil, err
	}

	for _, v := range m.files {
		if v.FSID == uint64(fsid) {
			return v, nil
		}
	}
	return nil, fmt.Errorf("%d not found", fsid)
}

func (m *FileList) Init() tea.Cmd { return nil }

func (m *FileList) Update(msg tea.Msg) (*FileList, tea.Cmd) {
	var cmd tea.Cmd
	// 先做原始修改操作
	m.model, cmd = m.model.Update(msg)
	logger.Infof("光标移动后的对象: %v", m.model.SelectedRow())

	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	// switch msg.String() {
	// case "esc":
	// if m.model.Focused() {
	// m.model.Blur()
	// } else {
	// m.model.Focus()
	// }
	// }
	// }
	return m, cmd
}

func (m FileList) View() string {
	return baseStyle.Render(m.model.View())
}

func (m *FileList) Focus() *FileList {
	m.model.Focus()
	return m
}

func (m *FileList) Blur() *FileList {
	m.model.Blur()
	return m
}

func (m FileList) Focused() bool {
	return m.model.Focused()
}

func NewFileList(files []*model.File, width, height int) *FileList {
	columns := []table.Column{
		{Title: "FSID", Width: 0},
		{Title: "文件名", Width: width - 10 - 10 - 20},
		{Title: "大小", Width: 10},
		{Title: "类型", Width: 10},
		{Title: "修改时间", Width: 20},
	}

	rows := make([]table.Row, 0)
	if files == nil {
		rows = append(rows, table.Row{
			"",
			"数据加载中...",
			"",
			"",
			"",
		})
	} else {
		for _, f := range files {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", f.FSID),
				fmt.Sprintf("%s %s", f.GetFileTypeEmoji(), f.GetFilename()),
				f.GetSize(),
				f.GetFileType(),
				f.GetServerMTime(),
			})
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height-4),
	)
	logger.Infof("Table height %d", t.Height())

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return &FileList{
		model: t,
		files: files,
	}
}
