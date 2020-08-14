package workhelper

import (
	"fmt"
	"math"
	"strings"
)

// 计算达标率，返回string 100.0%
func percent(total, pass int) string {
	var res float64 = float64(pass) / float64(total) * 100
	return fmt.Sprintf("%.1f%%", res)
}

// 计算超标倍数
func calcMul(value, std float64) string {
	return fmt.Sprintf("%.2f", math.Abs(value-std)/std)
}

// 地表水 格式化总磷/总氮
// tag=河流 总磷-->总磷(河流) 总氮-->总氮(河流)
// tag=湖库 总磷-->总磷(湖库)
func formatTpTn(key, tag string) string {
	if key == "总磷" {
		return key + "(" + tag + ")"
	}
	if tag == "河流" && key == "总氮" {
		return key + "(" + tag + ")"
	}
	return key
}

// 格式化指标名称
// 硒（四价）-->硒
func formatKey(key string) string {
	var i int
	i = strings.Index(key, "（")
	if i > -1 {
		key = key[:i]
	} else {
		i = strings.Index(key, "(")
		if i > -1 {
			key = key[:i]
		}
	}
	key = strings.TrimSpace(key)
	return key
}

func levelToString(level int) string {
	var lvs = []string{"Ⅰ", "Ⅱ", "Ⅲ", "Ⅳ", "Ⅴ", "劣Ⅴ"}
	return lvs[level-1]
}

func maxLevel(levels []int) int {
	level := 0
	for i := 0; i < len(levels); i++ {
		if level < levels[i] {
			level = levels[i]
		}
	}
	return level
}

// 检查地表水氨氮和总氮逻辑
func checkNH3TN(cNH3, cTN float64) string {
	if cNH3 > cTN {
		return "氨氮>总氮;"
	}
	return ""
}
func err(keys []string, strPre string, strSuf string) string {
	if len(keys) > 0 {
		strKeys := strings.Join(keys, "、")
		return "\"" + strPre + strKeys + strSuf + "\""
	}
	return ""
}