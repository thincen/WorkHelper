// 地表水

package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/tealeg/xlsx"
	"github.com/thincen/workHelper/std"
)

type surParam struct {
	*param
	useCol     int //= 15 // 取水量列号
	surTypeCol int //= 11 // 地表水类型列号
	townCol    int //= 6  // 乡镇名/测站名称
}
type surNode struct {
	*node
	town string // 乡镇名称
	use  string // 取水量
	tag  string // 河流/湖库
	ps   string // 地表水不参与评价的信息
}
type surResult struct {
	*result
	lake      int
	lakepass  int
	river     int
	riverpass int
}

func handleSur(f *xlsx.File, task *Task) error {
	// handle data
	var (
		p   = initSurParam()
		res = &surResult{result: &result{}} // 统计结果
		sh  = f.Sheets[0]
		ch  = make(chan *surNode, sh.MaxRow)
		wg  = new(sync.WaitGroup)
		err error
	)

	// 导入标准
	// 优先使用 std 文件夹下的csv外置标准文件
	if len(task.StdFile) > 0 {
		stdlimits, err = std.LimitsGen(task.StdFile)
		if err != nil {
			return err
		}
	} else {
		stdlimits = std.ParseStd(std.Sur)
	}

	wg.Add(sh.MaxRow - p.dataRow)

	for row := p.dataRow; row < sh.MaxRow; row++ {
		data := sh.Rows[row]
		go p.handleRow(data, ch, wg, nil)
	}

	chErr := make(chan error)
	go func() {
		err := surSave(task.Output, ch, res)
		task.Detail = res.detail()
		chErr <- err
	}()

	wg.Wait()
	close(ch)

	return <-chErr
}

// 处理每行数据
func (p *surParam) handleRow(row *xlsx.Row, chNode chan<- *surNode, wg *sync.WaitGroup, surnode *surNode) {
	defer wg.Done()
	var oneNode *surNode
	if surnode == nil {
		oneNode = &surNode{node: &node{}}
		oneNode.use = row.Cells[p.useCol].String()
		oneNode.tag = row.Cells[p.surTypeCol].String() // 湖库/河流
		oneNode.name = row.Cells[p.nameCol].String()
		oneNode.town = row.Cells[p.townCol].String()
	} else {
		oneNode = surnode
	}

	var (
		ltLimitKeys  = make([]string, 0)
		noHandleKeys = make([]string, 0)
		noExistKeys  = make([]string, 0)
		errStdValues = make([]string, 0)
		nh3          float64
		tn           float64
	)

	var (
		levels = make([]int, p.keyLen)
	)
	for i := p.keyCol; i < p.keyCol+p.keyLen; i++ {
		oneKey := p.handleSinKey(row, i, oneNode.tag)
		if strings.HasPrefix(oneKey.key, "总氮") {
			tn = oneKey.fvalue
		}
		if strings.HasPrefix(oneKey.key, "氨氮") {
			nh3 = oneKey.fvalue
		}
		if strings.Contains("总氮 粪大肠菌群", oneKey.key) {
			oneNode.ps = oneNode.ps + oneKey.info
		} else {
			levels = append(levels, oneKey.lv.level)
			oneNode.info = oneNode.info + oneKey.info
		}
		switch oneKey.err {
		case noExistKey:
			noExistKeys = append(noExistKeys, oneKey.key)
		case noHandle:
			noHandleKeys = append(noHandleKeys, oneKey.errMsg)
		case errNDvalue:
			ltLimitKeys = append(ltLimitKeys, oneKey.key)
		case errStdValue:
			errStdValues = append(errStdValues, oneKey.errMsg)
		}
	}
	oneNode.level = maxLevel(levels)
	oneNode.err = oneNode.err + checkNH3TN(nh3, tn)
	oneNode.err = oneNode.err + err(ltLimitKeys, "指标：", "检出限大于一类标准限值;")
	oneNode.err = oneNode.err + err(noHandleKeys, "指标：", "未处理(转换float失败),检查是否录入错误;")
	oneNode.err = oneNode.err + err(noExistKeys, "指标：", "未处理(未查询到标准值,检查指标名称是否正确);")
	oneNode.err = oneNode.err + err(errStdValues, "指标：", "转换float失败,结果评价可能有错;")
	chNode <- oneNode
}

func initSurParam() *surParam {
	return &surParam{
		param: &param{
			keyRow:  0,
			keyCol:  18,
			keyLen:  28,
			nameCol: 8,
			dataRow: 3,
		},
		useCol:     15,
		surTypeCol: 11,
		townCol:    6,
	}
}

func (p *surParam) handleSinKey(f *xlsx.Row, col int, tag string) *sinKey {
	var (
		sk    = &sinKey{lv: &level{}}
		key   = f.Sheet.Rows[p.keyRow].Cells[col].String()
		value = strings.TrimSpace(f.Cells[col].String())
		e     error
	)
	// if tagB {
	// 	tag = "河流"
	// } else {
	// 	tag = f.Cells[p.surTypeCol].String() // 湖库/河流
	// }
	key = formatKey(key)
	key = formatTpTn(key, tag)
	sk.key = key
	if value == "-1" {
		sk.fvalue = -1
		return sk
	}
	if e = sk.parseFloat(value); e != nil {
		return sk
	}

	// pH
	if strings.ToLower(key) == "ph" {
		sk.surPH(sk.fvalue)
		return sk
	}

	limits, ok := (*stdlimits)[key]
	// 检查指标是否在标准中的存在
	if !ok {
		if key != "总氮(河流)" {
			sk.err = noExistKey
		}
		return sk
	}

	// handle 未检出
	if value[len(value)-1:] == "L" {
		sk.handleND(limits[0])
		return sk
	}

	// 溶解氧
	if key == "溶解氧" {
		return sk.O2(limits)
	}
	// "溶解氧 硫酸盐 氯化物 硝酸盐 铁 锰"
	if strings.Contains("溶解氧 硫酸盐 氯化物 硝酸盐 铁 锰", key) {
		return sk.surPS(limits)
	}
	// 正常情况
	return sk.normal(limits)
}

func (sk *sinKey) surPH(fvalue float64) *sinKey {
	if fvalue < 6 || fvalue > 9 {
		sk.lv.level = 6
		sk.info = "pH(劣Ⅴ);"
		return sk
	}
	sk.lv.level = 1
	return sk
}

// handle 溶解氧
func (sk *sinKey) O2(limits []string) *sinKey {
	var (
		lenLimits = len(limits)
		i         = 0
		flimit3   float64 // Ⅲ类标准值
	)
	for i = 0; i < lenLimits; i++ {
		flimit, e := strconv.ParseFloat(limits[i], 64)
		if e != nil {
			sk.err = errStdValue
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
		if sk.fvalue >= flimit {
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

// "溶解氧 硫酸盐 氯化物 硝酸盐 铁 锰"
func (sk *sinKey) surPS(limits []string) *sinKey {
	flimit, e := strconv.ParseFloat(limits[0], 64)
	if e != nil {
		sk.err = errStdValue
		sk.errMsg = sk.key + "(" + limits[0] + ")"
		return sk
	}
	if flimit < sk.fvalue {
		sk.lv.level = 4 // 粗暴
		sk.lv.multiple = calcMul(sk.fvalue, flimit)
		sk.info = sk.lv.info(sk.key)
	} else {
		sk.lv.level = 1
	}
	return sk
}
func surSave(file string, nodes <-chan *surNode, res *surResult) error {
	f, err := xlsx.OpenFile(file)
	if err != nil {
		f = xlsx.NewFile()
	}
	var (
		sh *xlsx.Sheet
	)
	for i, sh := range f.Sheets {
		if sh.Name == "地表水" {
			f.Sheets = append(f.Sheets[:i], f.Sheets[i+1:]...)
			delete(f.Sheet, "地表水")
			break
		}
	}
	sh, err = f.AddSheet("地表水")
	if err != nil {
		return errors.New("保存结果文件:\n" + file + "\n创建表格失败\n" + err.Error())
	}
	for node := range nodes {

		switch {
		case strings.Contains(node.tag, "湖库"):
			res.lake++
			if node.level < 4 {
				res.lakepass++
			}
		case strings.Contains(node.tag, "河流"):
			res.river++
			if node.level < 4 {
				res.riverpass++
			}
		}
		row := sh.AddRow()
		row.AddCell().Value = node.town
		row.AddCell().Value = node.name
		row.AddCell().Value = node.use
		row.AddCell().Value = levelToString(node.level)
		row.AddCell().Value = node.info
		row.AddCell().Value = node.ps
		warnCell := row.AddCell()
		warnCell.Value = node.err
	}
	res.total = res.lake + res.river
	res.pass = res.lakepass + res.riverpass
	detailRow, err := sh.AddRowAtIndex(0)
	if err != nil {
		return errors.New("保存结果文件:\n" + file + "\n插入detail失败\n" + err.Error())
	}
	detailRow.AddCell().Value = res.detail()
	err = f.Save(file)
	if err != nil {
		return errors.New("保存结果文件:\n" + file + "\n失败\n" + err.Error())
	}
	return nil
}
func (res *surResult) detail() string {
	return fmt.Sprintf("共处理点位: %d 个,超标: %d 个,达标率: %s\n河流型: %d 个,超标: %d 个,达标率: %s\n湖库型: %d 个,超标: %d 个,达标率: %s\n",
		res.total, res.total-res.pass, percent(res.total, res.pass),
		res.river, res.river-res.riverpass, percent(res.river, res.riverpass),
		res.lake, res.lake-res.lakepass, percent(res.lake, res.lakepass))
}

/*
func countAndSave(file string, nodes <-chan *surNode, res *surResult) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return errors.New("保存结果文件:\n" + file + "\n失败\n" + err.Error())
	}
	defer f.Close()
	for node := range nodes {
		switch {
		case strings.Contains(node.tag, "湖库"):
			res.lake++
			if node.level < 4 {
				res.lakepass++
			}
		case strings.Contains(node.tag, "河流"):
			res.river++
			if node.level < 4 {
				res.riverpass++
			}
		}
		fmt.Fprintln(f, node.town+","+node.name+","+node.use+","+levelToString(node.level)+","+"\""+node.info+"\""+","+"\""+node.ps+"\""+","+node.err)
	}
	res.total = res.lake + res.river
	res.pass = res.lakepass + res.riverpass
	return nil
}
*/
