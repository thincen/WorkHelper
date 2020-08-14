// 转换 string 至 map[string][]string 供使用
// 加载外部标准文件 格式为: 标准名,1类限值,2类限值...\n

package std

import (
	"errors"
	"io/ioutil"
	"strings"
)

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

// LimitsGen 读取标准限值文件 生成Limits
func LimitsGen(stdFile string) (*Limits, error) {
	buf, err := ioutil.ReadFile(stdFile)
	if err != nil {
		return nil, errors.New("加载外部标准文件错误\n" + err.Error())
	}
	return ParseStd(string(buf)), nil
}
