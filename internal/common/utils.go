package common

import "time"

const (
	FMT_DATETIME string = "2006-01-02 15:04:05"
)

// 是否为空时间
func IsNilTime(t time.Time) bool {
	nilTimeStr := "0001-01-01 00:00:00"
	nilTime, _ := time.Parse(FMT_DATETIME, nilTimeStr)
	return nilTime.Equal(t)
}
