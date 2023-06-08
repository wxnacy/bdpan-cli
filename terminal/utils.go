package terminal

import (
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
