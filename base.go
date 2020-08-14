// 2020.8.7
// 提供处理 地表水/地下水 逻辑

package workhelper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/thincen/WorkHelper/std"
)

// Tag 处理模块标识
type Tag uint8
type errCode uint8

const (
	// Sur 地表水模块标识
	Sur Tag = iota
	// Wel 地表水模块标识
	Wel
	// Riv 地表水模块标识
	Riv
	// WebView bool 提供网页显示结果
	WebView     bool    = true
	noHandle    errCode = iota + 100 // 上报表格中转换float64错误的值
	noExistKey                       // 上报表格中存在，标准中不存在的指标
	errNDvalue                       // 未检出,检出限大于Ⅰ类标准值
	errStdValue                      // 标准文件中转换float64错误的值
)

var stdlimits *std.Limits

// := std.ParseStd(std.Sur)
// 表格内部位置
type param struct {
	keyRow  int //= 0  // 指标行号
	keyCol  int //= 18 // 指标开始列号
	keyLen  int //= 28 // 指标数
	nameCol int //= 8  // 点位名称列号
	dataRow int //= 3  // 数据开始行号
}
type level struct {
	level    int    // 1-6
	multiple string // 超标倍数
}

type sinKey struct {
	key    string
	fvalue float64
	lv     *level
	info   string
	err    errCode
	errMsg string // errCode的附加说明
}

type node struct {
	name string // 点位名称
	// level string // "Ⅰ", "Ⅱ", "Ⅲ", "Ⅳ", "Ⅴ", "劣Ⅴ"
	level int
	info  string // 超标指标(等级,超标倍数)
	err   string // 错误信息
}
type result struct {
	total int
	pass  int
}

func (lv *level) info(key string) string {
	// return key+"("+lv.string()+",超标"
	return fmt.Sprintf("%s(%s,超标%s倍);", key, levelToString(lv.level), lv.multiple)
}

// 未检出处理
func (sk *sinKey) handleND(limit1 string) {
	flimit, e := strconv.ParseFloat(limit1, 64)
	if e != nil {
		sk.err = errStdValue
		sk.errMsg = sk.key + "(" + limit1 + ")"
		return
	}
	if sk.fvalue > flimit {
		sk.err = errNDvalue
		return
	}
	sk.lv.level = 1
}

// 指标检测值转为float64
func (sk *sinKey) parseFloat(value string) error {
	value = strings.ReplaceAll(value, "<", "")
	value = strings.ReplaceAll(value, "L", "")

	fvalue, e := strconv.ParseFloat(value, 64)
	if e != nil {
		sk.err = noHandle
		sk.errMsg = sk.key + "(" + value + ")"
		return e
	}
	sk.fvalue = fvalue
	return nil
}

func (sk *sinKey) normal(limits []string) *sinKey {
	var (
		lenLimits = len(limits)
		flimit3   float64
		i         = 0
	)
	for i = 0; i < lenLimits; i++ {
		flimit, e := strconv.ParseFloat(limits[i], 64)
		if e != nil {
			if strings.HasPrefix(limits[i], ">") {
				sk.lv.level = i + 1
				if i > 2 {
					sk.lv.multiple = calcMul(sk.fvalue, flimit3)
					sk.info = sk.lv.info(sk.key)
					return sk
				}
			}
			// handle Wel 阴离子表面活性剂
			if limits[i] == "不得检出" {
				continue
			}
			if len(sk.errMsg) == 0 {
				sk.errMsg = sk.key + "(" + limits[i] + ")"
			} else {
				sk.errMsg = sk.errMsg[:len(sk.errMsg)-1] + ";" + limits[i] + ")"
			}
			continue
		}

		if i == 2 {
			flimit3 = flimit
		}
		if sk.fvalue <= flimit {
			break
		}
	}
	sk.lv.level = i + 1
	if sk.lv.level > 3 {
		sk.lv.multiple = calcMul(sk.fvalue, flimit3)
		sk.info = sk.lv.info(sk.key)
	}
	return sk
}
