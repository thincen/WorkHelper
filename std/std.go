package std

import "strings"

// Limits 标准限值map
type Limits map[string][]string

// ParseStd 使用包内默认标准生成 stdLimit
func ParseStd(strStd string) *Limits {
	std := make(Limits)
	lines := strings.Split(strStd, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		strArr := strings.Split(line, ",")
		std[strArr[0]] = strArr[1:]
	}
	return &std
}
