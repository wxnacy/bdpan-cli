package downloader

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/go-tools"
)

// ProgressModel 下载进度条 Model
type ProgressModel struct {
	filename   string
	totalSize  int64
	downloaded int64
	progress   progress.Model
	startTime  time.Time
	quitting   bool
	err        error
	cancelFunc context.CancelFunc // 取消下载的函数
	status     string             // 状态行（例如当前文件）
}

// progressMsg 进度更新消息
type progressMsg struct {
	downloaded int64
	total      int64
}

// completedMsg 下载完成消息
type completedMsg struct{}

// errorMsg 错误消息
type errorMsg struct{ err error }

// statusMsg 状态更新消息
type statusMsg struct{ text string }

// NewProgressModel 创建进度条 Model
func NewProgressModel(filename string, totalSize int64, cancelFunc context.CancelFunc) ProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	
	return ProgressModel{
		filename:   filename,
		totalSize:  totalSize,
		downloaded: 0,
		progress:   p,
		startTime:  time.Now(),
		cancelFunc: cancelFunc,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return nil
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			// 取消下载
			if m.cancelFunc != nil {
				m.cancelFunc()
			}
			return m, tea.Quit
		}

	case progressMsg:
		m.downloaded = msg.downloaded
		if m.downloaded >= m.totalSize {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case completedMsg:
		m.quitting = true
		return m, tea.Quit

	case errorMsg:
		m.err = msg.err
		m.quitting = true
		return m, tea.Quit

	case statusMsg:
		m.status = msg.text
		return m, nil

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - 20
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}
		return m, nil

	default:
		return m, nil
	}

	return m, nil
}

func (m ProgressModel) View() string {
	if m.err != nil {
		// 让外层控制错误/取消的显示
		return ""
	}

	if m.quitting {
		// 让外层控制完成或取消的提示
		return ""
	}

	var b strings.Builder

	// 文件名（左对齐，标题单独一行）
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("下载: %s", m.filename)))
	b.WriteString("\n")

	// 状态行（可选）
	if m.status != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		b.WriteString(statusStyle.Render(m.status))
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n")
	}

	// 进度条（确保从行首开始）
	percent := float64(m.downloaded) / float64(m.totalSize)
	bar := m.progress.ViewAs(percent)
	bar = strings.TrimLeft(bar, " ")
	b.WriteString(bar)
	b.WriteString("\n\n")

	// 进度信息
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	downloadedStr := tools.FormatSize(m.downloaded)
	totalStr := tools.FormatSize(m.totalSize)
	percentStr := fmt.Sprintf("%.1f%%", percent*100)
	
	// 计算速度和剩余时间
	elapsed := time.Since(m.startTime).Seconds()
	if elapsed > 0 {
		speed := float64(m.downloaded) / elapsed
		speedStr := formatSpeed(int64(speed))
		
		var etaStr string
		if speed > 0 {
			remaining := float64(m.totalSize-m.downloaded) / speed
			etaStr = formatDuration(time.Duration(remaining) * time.Second)
		} else {
			etaStr = "--:--"
		}
		
		info := fmt.Sprintf("%s / %s  %s  速度: %s  剩余: %s",
			downloadedStr, totalStr, percentStr, speedStr, etaStr)
		b.WriteString(infoStyle.Render(info))
	} else {
		info := fmt.Sprintf("%s / %s  %s",
			downloadedStr, totalStr, percentStr)
		b.WriteString(infoStyle.Render(info))
	}

	b.WriteString("\n\n")
	
	// 提示信息
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(helpStyle.Render("按 q 或 Ctrl+C 取消"))
	b.WriteString("\n")

	return b.String()
}

// formatSpeed 格式化速度
func formatSpeed(bytesPerSecond int64) string {
	if bytesPerSecond < 1024 {
		return fmt.Sprintf("%d B/s", bytesPerSecond)
	} else if bytesPerSecond < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", float64(bytesPerSecond)/1024)
	} else if bytesPerSecond < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB/s", float64(bytesPerSecond)/1024/1024)
	}
	return fmt.Sprintf("%.1f GB/s", float64(bytesPerSecond)/1024/1024/1024)
}

// formatDuration 格式化时间
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// ProgressWriter 用于在下载时更新进度的 Writer
type ProgressWriter struct {
	program    *tea.Program
	downloaded int64
	total      int64
}

// NewProgressWriter 创建进度写入器
func NewProgressWriter(program *tea.Program, total int64) *ProgressWriter {
	return &ProgressWriter{
		program: program,
		total:   total,
	}
}

// UpdateProgress 更新进度
func (pw *ProgressWriter) UpdateProgress(downloaded, total int64) {
	pw.downloaded = downloaded
	pw.total = total
	if pw.program != nil {
		pw.program.Send(progressMsg{
			downloaded: downloaded,
			total:      total,
		})
	}
}

// UpdateStatus 更新状态行
func (pw *ProgressWriter) UpdateStatus(text string) {
	if pw.program != nil {
		pw.program.Send(statusMsg{text: text})
	}
}

// Complete 标记完成
func (pw *ProgressWriter) Complete() {
	if pw.program != nil {
		pw.program.Send(completedMsg{})
	}
}

// Error 发送错误
func (pw *ProgressWriter) Error(err error) {
	if pw.program != nil {
		pw.program.Send(errorMsg{err: err})
	}
}
