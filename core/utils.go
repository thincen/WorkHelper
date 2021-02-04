package core

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tealeg/xlsx"
)

// 格式化指标名称
// 硒（四价）-->硒
func formatKey(key string, tag tag) string {
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
	key = strings.ReplaceAll(key, " ", "")
	if key == "总磷" || key == "总氮" {
		key = formatTpTn(key, tag)
	}
	return key
}

// 地表水 格式化总磷/总氮
// tag=河流 总磷-->总磷(河流) 总氮-->总氮(河流)
// tag=湖库 总磷-->总磷(湖库)
func formatTpTn(key string, tag tag) string {
	var strTag string
	if tag == Riv {
		strTag = "河流"
	}
	if tag == Lak {
		strTag = "湖库"
	}
	if key == "总磷" {
		return key + "(" + strTag + ")"
	}
	if tag == Riv && key == "总氮" {
		return key + "(" + strTag + ")"
	}
	return key
}

// 计算达标率，返回string 100.0%
func percent(total, pass int) string {
	var res float64 = float64(pass) / float64(total) * 100
	return fmt.Sprintf("%.1f%%", res)
}

// 计算超标倍数 银行家计算
func calcMul(value, std string) string {
	// 前面已经转换过，不会出错
	v, _ := decimal.NewFromString(value)
	limit, _ := decimal.NewFromString(std)
	return v.Sub(limit).Abs().Div(limit).StringFixedBank(2)
}

func maxLevel(levels []Level) Level {
	var level Level = 0
	for i := 0; i < len(levels); i++ {
		if level < levels[i] {
			level = levels[i]
		}
	}
	return level
}

// 检查地表水氨氮和总氮逻辑
func checkNH3TN(cNH3, cTN float64) string {
	if cTN == -1 || cNH3 <= cTN {
		return ""
	}
	return "氨氮>总氮;"
}

// 消除"未检出"和未填写
func formatValue(v string) string {
	v = strings.TrimSpace(v)
	// 未填写
	if len(v) == 0 {
		return "-1"
	}
	v = strings.ReplaceAll(v, " ", "")
	// 未检出
	if strings.Compare(v, "未检出") == 0 {
		return "检出限+L"
	}
	// 非法数据
	if strings.HasPrefix(v, "<") {
		v = appendString(v[1:], "L")
	}
	// 科学计数

	return v
}

func byte2Xlsx(b []byte) (*xlsx.File, error) {
	return xlsx.OpenBinary(b)
}

// 未检出时判断是否小于1类标准限值
func checkND(value string, l1 limit) error {
	value = strings.ReplaceAll(value, "L", "")
	fl1, err := l1.float()
	if err != nil {
		return errStdValue
	}
	fv, err := limit(value).float()
	if err != nil {
		return errNoHandle
	}
	if fv > fl1 && fl1 != 0 {
		return errNDvalue
	}
	return nil
}

//  指标是否在不处理范围内
func isBanned(key string, banKeys []string) bool {
	return strings.Contains(strings.Join(banKeys, ";"), key)
}

// 使用 strings.Builder 拼接
func appendString(strs ...string) string {
	var strBuilder = new(strings.Builder)
	for _, v := range strs {
		strBuilder.WriteString(v)
	}
	return strBuilder.String()
}
