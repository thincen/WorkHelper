package core

import (
	"strconv"
	"strings"
)

type param struct {
	keyRow  int // 指标行号
	keyCol  int // 指标开始列号
	keyLen  int // 指标数量
	nameCol int // 点位名称
	dataRow int // 数据开始行号
}

// Level 使用等级评价
type Level uint8
type tag uint8
type limit string // 标准值、检测值
type errCode uint8
type warnKeys []string
type warn map[errCode]warnKeys

const (
	errLevel      Level = 10
	noHandleLevel Level = 11
)
const (
	// Riv 断面
	Riv tag = 1 << iota
	// Lak 湖库
	Lak
	// Wel 地下水
	Wel
)

const (
	// L1 leve 1
	L1 Level = iota
	// L2 leve 2
	L2
	// L3 leve 3
	L3
	// L4 level 4
	L4
	// L5 level 5
	L5
	// L6 地表水劣5类
	L6
)

const (
	errNoHandle   errCode = 1 << iota // 上报表格中转换float64错误的值
	errNoExistKey                     // 上报表格中存在，标准中不存在的指标
	errNDvalue                        // 未检出,检出限大于Ⅰ类标准值
	errStdValue                       // 标准文件中转换float64错误的值
	errLogicValue                     // TN<NH3
	errAvg                            // 计算均值错误
)

func (ec errCode) Error() string {
	switch ec {
	case errNoHandle:
		return "指标未处理,转换float64错误"
	case errNoExistKey:
		return "未查找到指标名称"
	case errNDvalue:
		return "检出限大于Ⅰ类标准值"
	case errStdValue:
		return "标准限值转换float64错误"
	case errLogicValue:
		return "逻辑错误,TN<NH3"
	case errAvg:
		return "计算均值错误"
	}
	return ""
}

// level 0-5 return "Ⅰ", "Ⅱ", "Ⅲ", "Ⅳ", "Ⅴ", "劣Ⅴ"
func (l Level) string() string {
	if l > L6 {
		return "/"
	}
	str := [...]string{"Ⅰ", "Ⅱ", "Ⅲ", "Ⅳ", "Ⅴ", "劣Ⅴ"}
	return str[l]
}

// string 2 float64
func (l limit) float() (float64, error) {
	if strings.Contains(string(l), ">") {
		return 1 << 32, nil
	}
	if strings.Compare(string(l), "不得检出") == 0 {
		return 0, nil
	}
	return strconv.ParseFloat(string(l), 64)
}

func (w *warn) string() string {
	var warnBuilder strings.Builder
	for e, v := range *w {
		warnBuilder.WriteString(e.Error())
		warnBuilder.WriteString("(")
		warnBuilder.WriteString(v.join(","))
		warnBuilder.WriteString(");\n")
	}
	return warnBuilder.String()
}

func (ks *warnKeys) join(sep string) string {
	return strings.Join(*ks, sep)
}

// pH 发生错误时,返回Level为10-errLevel
func pH(phValue string, tag tag) (Level, error) {
	switch tag {
	case Lak, Riv:
		f, e := limit(phValue).float()
		if e != nil {
			return errLevel, errNoHandle
		}
		if f < 6 || f > 9 {
			return L6, nil
		}
		return L1, nil
	case Wel:
		f, e := limit(phValue).float()
		if e != nil {
			return errLevel, errNoHandle
		}
		if f >= 6.5 && f <= 8.5 {
			return L1, nil
		} else if (f >= 5.5 && f < 6.5) || (f > 8.5 && f <= 9) {
			return L4, nil
		}
		return L5, nil
	}
	return noHandleLevel, errNoHandle
}

func o2(o2value string, limits []string) (Level, error) {
	if strings.Contains(o2value, "L") {
		return L6, nil
	}
	lenLimits := len(limits)
	fvalue, e := limit(o2value).float()
	if e != nil {
		return errLevel, errNoHandle
	}
	var i int
	for i = 0; i < lenLimits; i++ {
		flimit, e := limit(limits[i]).float()
		if e != nil {
			return errLevel, errStdValue
		}
		if fvalue >= flimit {
			break
		}
	}
	return Level(i), nil
}

// 检测值不为未检出；不为特殊指标
func nomalKey(value string, limits []string) (Level, error) {
	lenLimits := len(limits)
	fvalue, e := limit(value).float()
	if e != nil {
		return errLevel, errNoHandle
	}
	var i int
	for i = 0; i < lenLimits; i++ {
		flimit, e := limit(limits[i]).float()
		if e != nil {
			return errLevel, errStdValue
		}
		if fvalue <= flimit {
			break
		}
	}
	return Level(i), nil
}

// 超标等级和倍数
// func genDes(l Level,key,info string) string {

// }
