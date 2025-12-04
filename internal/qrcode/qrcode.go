package qrcode

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/wxnacy/bdpan-cli/internal/logger"
)

func CreateQRCodeImage(text string, size int, filename string) error {
	scaleW := size
	scaleH := size
	// 生成二维码
	qrCode, err := qr.Encode(text, qr.M, qr.Auto)
	if err != nil {
		return err
	}

	// 可选：调整二维码的大小
	qrCode, err = barcode.Scale(qrCode, scaleW, scaleH)
	if err != nil {
		return err
	}

	// 创建图像文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 编码为PNG并保存到文件
	err = png.Encode(file, qrCode)
	if err != nil {
		return err
	}
	return nil
}

func ShowByUrl(uri string, timeout time.Duration) error {
	return ShowByUrlWithSize(uri, timeout, 100, 50)
}

// ShowByUrlWithSize 显示二维码，可以自定义显示大小
func ShowByUrlWithSize(uri string, timeout time.Duration, width, height int) error {
	var images []image.Image
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	image, _, err := image.Decode(resp.Body)
	if err != nil {
		return err
	}
	images = append(images, image)

	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()

	img := widgets.NewImage(nil)
	img.SetRect(0, 0, width, height)
	index := 0
	render := func() {
		img.Image = images[index]
		img.Monochrome = true
		img.Title = "BDPan"
		ui.Render(img)
	}
	render()

	// uiEvents := ui.PollEvents()
	for i := range int(timeout / time.Second) {
		deadline := int(timeout/time.Second) - i
		logger.Infof("二维码倒计时 %d", deadline)
		img.Title = fmt.Sprintf("BDPan %d", deadline)
		// e := <-uiEvents
		// switch e.ID {
		// case "q", "<C-c>":
		// return errors.New("Exit")
		// }
		render()
		time.Sleep(1 * time.Second)
	}
	return nil
}

// qrModel 是用于二维码展示的 bubbletea Model
//
// 设计说明：
// - 使用 bubbletea 的 Model-Update-View 架构实现响应式 UI
// - 支持实时倒计时、用户交互（q/Ctrl+C取消）、外部登录成功通知
// - 状态机设计：正常显示 -> 登录成功/用户取消/超时 -> 退出
//
// 字段说明：
// - qrCode: 二维码的字符串表示（使用半块字符渲染）
// - remaining: 剩余倒计时秒数
// - totalTime: 总倒计时时间（用于计算进度）
// - quitting: 是否正在退出（控制 View 渲染）
// - interrupted: 用户是否主动取消（按 q/Ctrl+C/Esc）
// - loginDone: 是否登录成功（外部通知）
// - timeout: 是否倒计时结束超时
type qrModel struct {
	qrCode      string
	remaining   int
	totalTime   int
	quitting    bool
	interrupted bool
	loginDone   bool
	timeout     bool
}

// tickMsg 每秒触发一次的倒计时消息
type tickMsg time.Time

// loginSuccessMsg 外部发送的登录成功通知消息
type loginSuccessMsg struct{}

// Init 初始化 bubbletea Model，启动倒计时
//
// 返回倒计时命令，每秒触发一次 tickMsg
func (m qrModel) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
	)
}

// tickCmd 创建一个每秒触发的倒计时命令
//
// 实现细节：
// - 使用 tea.Tick 创建定时器
// - 每秒返回一个 tickMsg，触发 Update 方法处理倒计时逻辑
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update 处理各种消息并更新 Model 状态
//
// 消息处理逻辑：
//
// 1. tea.KeyMsg - 键盘输入
//   - q/Ctrl+C/Esc: 用户主动取消，设置 interrupted=true，退出程序
//   - 其他按键: 忽略
//
// 2. tickMsg - 每秒倒计时
//   - remaining 减 1
//   - 记录日志用于调试
//   - 如果 remaining <= 0: 设置 timeout=true，退出程序
//   - 否则: 返回新的 tickCmd 继续倒计时
//
// 3. loginSuccessMsg - 外部登录成功通知
//   - 设置 loginDone=true，立即退出程序
//   - 由外部 goroutine 通过 p.Send() 发送此消息
//
// 设计要点：
// - 三种退出状态互斥：interrupted、timeout、loginDone
// - 所有退出都设置 quitting=true，触发 View 显示相应提示
// - 返回 tea.Quit 命令会结束 bubbletea 程序
func (m qrModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			m.interrupted = true
			return m, tea.Quit
		}

	case tickMsg:
		m.remaining--
		logger.Infof("二维码倒计时 %d", m.remaining)
		if m.remaining <= 0 {
			m.quitting = true
			m.timeout = true
			return m, tea.Quit
		}
		return m, tickCmd()

	case loginSuccessMsg:
		m.loginDone = true
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View 渲染当前 UI 界面
//
// 渲染逻辑：
//
// 1. 如果正在退出（quitting=true），根据状态显示不同的退出提示：
//   - interrupted: 显示红色 "登录已取消"
//   - loginDone: 显示绿色 "✓ 登录成功！"
//   - timeout: 显示橙色 "登录超时，请重试"
//   - 其他: 返回空字符串（不应该发生）
//
// 2. 正常显示时，从上到下渲染：
//   - 标题: "BDPan 登录" (蓝色加粗，居中)
//   - 二维码: 带圆角边框和内边距 (蓝色边框)
//   - 提示文字: "请使用百度网盘APP扫描二维码登录 (倒计时: X秒)" (橙色)
//   - 帮助文字: "按 q 或 Ctrl+C 取消" (灰色)
//
// 样式说明：
// - 使用 lipgloss 库实现样式
// - 颜色方案: 主题色蓝色(63)、成功绿色(46)、警告橙色(214)、错误红色(196)
// - 布局: 所有元素居中对齐
func (m qrModel) View() string {
	if m.quitting {
		if m.interrupted {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Render("登录已取消\n")
		}
		if m.loginDone {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true).
				Render("✓ 登录成功！\n")
		}
		if m.timeout {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Render("登录超时，请重试\n")
		}
		return ""
	}

	// 定义样式
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(1)

	qrStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Align(lipgloss.Center)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Align(lipgloss.Center).
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		MarginTop(1)

	// 构建视图
	var view strings.Builder
	view.WriteString(titleStyle.Render("BDPan 登录"))
	view.WriteString("\n\n")
	view.WriteString(qrStyle.Render(m.qrCode))
	view.WriteString("\n")
	view.WriteString(hintStyle.Render(fmt.Sprintf("请使用百度网盘APP扫描二维码登录 (倒计时: %d秒)", m.remaining)))
	view.WriteString("\n")
	view.WriteString(helpStyle.Render("按 q 或 Ctrl+C 取消"))
	view.WriteString("\n")

	return view.String()
}

// ShowQRCodeWithLipgloss 使用lipgloss和bubbletea展示二维码（基于字符渲染）
//
// 参数：
// - imageURL: 二维码图片的 URL，会自动下载并转换为终端字符显示
// - timeout: 显示超时时间
//
// 返回：
// - error: 下载或显示错误，正常结束返回 nil
//
// 说明：这是简化版本，内部调用 ShowQRCodeWithCallback 但不接收任何通知
func ShowQRCodeWithLipgloss(imageURL string, timeout time.Duration) error {
	return ShowQRCodeWithCallback(imageURL, timeout, nil, nil, nil)
}

// ShowQRCodeWithCallback 使用 lipgloss 和 bubbletea 展示二维码，支持外部通知与交互
//
// 参数：
// - imageURL: 二维码图片的 URL，会通过 HTTP GET 下载
// - timeout: 显示超时时间，倒计时结束后自动退出
// - successChan: 接收登录成功通知，外部向此 channel 发送 true 会立即退出并显示成功提示
// - cancelChan: 发送用户取消通知，当用户按 q/Ctrl+C/Esc 时，向此 channel 发送 true
// - timeoutChan: 发送超时通知，当倒计时结束时，向此 channel 发送 true
//
// 返回：
// - error: 仅当下载或解码图片失败时返回错误，其他情况（取消/超时/成功）均返回 nil
//
// 实现逻辑：
//
// 1. 下载并转换二维码图片
//   - 从 imageURL 下载图片
//   - 将图片转换为终端字符串（使用半块字符 █▀▄ 渲染）
//
// 2. 创建 bubbletea Model
//   - 初始化 qrModel，设置二维码内容和倒计时
//   - 启动 bubbletea 程序
//
// 3. 监听外部登录成功通知
//   - 如果提供了 successChan，启动 goroutine 监听
//   - 接收到通知后，通过 p.Send() 发送 loginSuccessMsg
//   - 触发 Model 的 Update 方法，设置 loginDone=true 并退出
//
// 4. 等待程序结束并处理最终状态
//   - interrupted: 向 cancelChan 发送 true
//   - timeout: 向 timeoutChan 发送 true
//   - loginDone: 不发送任何通知，由外部已经知道
//
// 设计要点：
// - 所有 channel 均为可选，传 nil 表示不需要该通知
// - 用户取消和超时的提示信息已在 UI 中显示，无需返回错误，避免重复显示
// - bubbletea 程序是阻塞运行的，直到退出前不会返回
func ShowQRCodeWithCallback(imageURL string, timeout time.Duration, successChan <-chan bool, cancelChan chan<- bool, timeoutChan chan<- bool) error {
	// 从URL下载并转换二维码图片为字符串
	qrStr, err := downloadAndConvertQRCode(imageURL)
	if err != nil {
		return err
	}

	// 创建model
	m := qrModel{
		qrCode:    qrStr,
		remaining: int(timeout / time.Second),
		totalTime: int(timeout / time.Second),
	}

	// 运行bubbletea程序
	p := tea.NewProgram(m)

	// 如果提供了成功通知channel，启动监听goroutine
	if successChan != nil {
		go func() {
			<-successChan
			p.Send(loginSuccessMsg{})
		}()
	}

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// 检查最终状态
	if fm, ok := finalModel.(qrModel); ok {
		if fm.interrupted {
			// 通知外部用户取消了操作
			if cancelChan != nil {
				cancelChan <- true
			}
		} else if fm.timeout {
			// 通知外部倒计时结束
			if timeoutChan != nil {
				timeoutChan <- true
			}
		}
	}

	return nil
}

// downloadAndConvertQRCode 从 URL 下载二维码图片并转换为终端字符串
//
// 参数：
// - imageURL: 二维码图片的 HTTP URL
//
// 返回：
// - string: 转换后的字符串，使用半块字符渲染
// - error: 下载或解码失败时的错误
//
// 实现步骤：
// 1. 使用 http.Get 下载图片
// 2. 使用 image.Decode 解码图片（自动识别格式）
// 3. 调用 imageToString 将图片转换为字符串
func downloadAndConvertQRCode(imageURL string) (string, error) {
	// 下载图片
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("下载二维码图片失败: %w", err)
	}
	defer resp.Body.Close()

	// 解码图片
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("解码二维码图片失败: %w", err)
	}

	// 将图片转换为字符串
	return imageToString(img), nil
}

// imageToString 将图片转换为终端字符串显示
//
// 参数：
// - img: 要转换的图片（image.Image 接口）
//
// 返回：
// - string: 使用终端字符渲染的图片字符串
//
// 转换算法：
//
// 1. 自动缩放
//   - 计算缩放比例，使二维码宽度适配终端（目标宽度 50 字符）
//   - scale = width / 50，最小为 1
//
// 2. 半块字符渲染
//   - 每 2 行像素合并为 1 行字符，提高显示密度
//   - 使用 RGBA() 获取像素颜色，计算灰度值 (0-255)
//   - 根据两个像素的灰度值选择字符：
//   - 两个都暗 (< 128): █ 全实心
//   - 上暗下亮: ▀ 上半实心
//   - 上亮下暗: ▄ 下半实心
//   - 两个都亮 (≥ 128): 空格
//
// 3. 灰度值计算
//   - RGBA() 返回 0-65535 范围的值
//   - gray = (r + g + b) / 3 / 257，转换为 0-255 范围
//
// 设计要点：
// - 使用半块字符可以将垂直分辨率提高 2 倍
// - 灰度阈值 128 是经验值，适合大多数二维码图片
// - 缩放算法保证二维码在终端中大小适中，易于扫描
func imageToString(img image.Image) string {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X

	// 计算缩放比例，使二维码适合终端显示（目标宽度约50字符）
	// 提高目标宽度，提升分辨率，便于手机识别
	// 同时保留半块字符渲染以尽量保持纵横比
	targetWidth := 100
	scale := width / targetWidth
	if scale < 1 {
		scale = 1
	}

	var qrStrBuilder strings.Builder

	// 添加顶部静区（quiet zone），2 行空白，增强识别稳定性
	for i := 0; i < 2; i++ {
		qrStrBuilder.WriteString("\n")
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += scale * 2 {
		// 左侧静区（适当留白）
		qrStrBuilder.WriteString("    ")

		for x := bounds.Min.X; x < bounds.Max.X; x += scale {
			// 使用半块字符来提高密度
			c1 := img.At(x, y)
			// y+scale 可能越界，进行边界保护
			y2 := y + scale
			if y2 >= bounds.Max.Y {
				y2 = bounds.Max.Y - 1
			}
			c2 := img.At(x, y2)

			// 转换为灰度值（0-255）
			r1, g1, b1, _ := c1.RGBA()
			r2, g2, b2, _ := c2.RGBA()
			gray1 := (r1 + g1 + b1) / 3 / 257
			gray2 := (r2 + g2 + b2) / 3 / 257

			// 更严格的二值化，尽量避免灰度过渡造成的锯齿
			const threshold uint32 = 140

			// 根据灰度值选择字符
			if gray1 < threshold && gray2 < threshold {
				qrStrBuilder.WriteString("█") // 全实心
			} else if gray1 < threshold && gray2 >= threshold {
				qrStrBuilder.WriteString("▀") // 上半
			} else if gray1 >= threshold && gray2 < threshold {
				qrStrBuilder.WriteString("▄") // 下半
			} else {
				qrStrBuilder.WriteString(" ") // 空白
			}
		}

		// 右侧静区
		qrStrBuilder.WriteString("    ")
		qrStrBuilder.WriteString("\n")
	}

	// 底部静区
	for i := 0; i < 2; i++ {
		qrStrBuilder.WriteString("\n")
	}

	return qrStrBuilder.String()
}

// generateQRCodeString 从文本生成二维码并转换为字符串
//
// 参数：
// - text: 要编码为二维码的文本内容
//
// 返回：
// - string: 生成的二维码字符串
// - error: 编码失败时的错误
//
// 说明：
// - 此函数用于从文本生成新的二维码
// - 与 downloadAndConvertQRCode 不同，后者是从 URL 下载现有的二维码图片
// - 当前在登录流程中使用 downloadAndConvertQRCode，此函数保留供未来扩展使用
// - 使用 qr.M 错误修正级别（中等），平衡容错率和数据容量
func generateQRCodeString(text string) (string, error) {
	// 生成二维码
	qrCode, err := qr.Encode(text, qr.M, qr.Auto)
	if err != nil {
		return "", err
	}

	return imageToString(qrCode), nil
}
