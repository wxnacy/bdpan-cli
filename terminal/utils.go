package terminal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// 超出部分使用符号省略
func OmitString(text string, maxWidth int) string {
	textW := runewidth.StringWidth(text)
	if textW > maxWidth {
		// 文字长度需要在减去 ... 的宽度
		return text[0:maxWidth-3] + "..."
	}
	return text
}

// 不够的部分使用空格填充
func FillString(text string, width int) string {
	text = runewidth.FillRight(text, width)
	// text = strings.ReplaceAll(text, " ", "-")
	return text
}

// 绘制当行
func DrawLine(s tcell.Screen, StartX, StartY int, style tcell.Style, text string) {
	runes := []rune(text)
	Log.Debugf("DrawLine StartX: %d StartY: %d Text: %s", StartX, StartY, text)
	s.SetContent(StartX, StartY, runes[0], runes[1:], style)
}
