package terminal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// 超出部分使用符号省略
func OmitString(text string, maxWidth int) string {
	return OmitStringMid(text, maxWidth)
}

// 将字符串右侧超出部分隐藏
func OmitStringRight(text string, maxWidth int) string {
	// fmt.Printf("%s %d\n", text, runewidth.StringWidth(text))
	textW := runewidth.StringWidth(text)
	if textW > maxWidth {
		textMaxWidth := maxWidth - 3
		var result string
		for _, s := range text {
			splitS := string(s)
			if runewidth.StringWidth(result+splitS) > textMaxWidth {
				break
			}
			result += splitS
		}
		// 文字长度需要在减去 ... 的宽度
		return result + "..."
	}
	return text
}

// 将字符串中间超出部分隐藏
func OmitStringMid(text string, maxWidth int) string {
	// fmt.Printf("%s %d\n", text, runewidth.StringWidth(text))
	textW := runewidth.StringWidth(text)
	if textW > maxWidth {
		textMaxWidth := maxWidth - 3
		var begin, end string
		var arr = make([]string, 0)
		for _, s := range text {
			arr = append(arr, string(s))
		}
		for i, s := range arr {
			if runewidth.StringWidth(begin+end+s) > textMaxWidth {
				break
			}
			begin += s
			if runewidth.StringWidth(begin+end+s) > textMaxWidth {
				break
			}
			end = arr[len(arr)-i-1] + end
		}
		// 文字长度需要在减去 ... 的宽度
		return begin + "..." + end
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
	// Log.Debugf("DrawLine StartX: %d StartY: %d Text: %s", StartX, StartY, text)
	s.SetContent(StartX, StartY, runes[0], runes[1:], style)
}
