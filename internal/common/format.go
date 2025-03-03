package common

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatNumberWithHeadZeros(num int, totalDigits int) string {
	return fmt.Sprintf(fmt.Sprintf("%%%02dd", totalDigits), num)
}

// FormatNumberWithTrailingZeros 将数字格式化为指定位数，不足部分后面补零
func FormatNumberWithTrailingZeros(num int, totalDigits int) string {
	// 将数字转换为字符串
	numStr := strconv.Itoa(num)

	// 计算需要补充的零的数量
	zeroCount := totalDigits - len(numStr)

	// 如果已经满足或超过位数，直接返回原始字符串
	if zeroCount <= 0 {
		return numStr
	}

	// 使用 strings.Repeat 动态生成零
	zeros := strings.Repeat("0", zeroCount)
	return numStr + zeros
}
