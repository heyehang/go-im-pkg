package ttime

import "time"

// Parse 避免time.local 忘记，因为标准库 少了时区，会导致解析时间少了 8小时
func Parse(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, time.Local)
}
