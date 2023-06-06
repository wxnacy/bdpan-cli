package terminal

import (
	"github.com/mattn/go-runewidth"
)

// 超出部分使用符号省略
func OmitString(text string, maxWidth int) string {
	textW := runewidth.StringWidth(text)
	if textW > maxWidth {
		// 宽度需要减去边框的，所以是 len("...") + 1
		return text[0:maxWidth] + "..."
	}
	return text
}

// 不够的部分使用空格填充
func FillString(text string, width int) string {
	text = runewidth.FillRight(text, width)
	// text = strings.ReplaceAll(text, " ", "-")
	return text
}
