// 饮用水-地下型

package core

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/tealeg/xlsx"
	"github.com/thincen/workHelper/std"
)

type welNode struct {
	*node
	use  string
	town string
}

type welParam struct {
	*param
	useCol  int //= 15 // 取水量列号
	townCol int //= 6  // 乡镇名
}

func handleWel(f *xlsx.File, task *Task) error {

	var (
		p   = initWelParam()
		res = &result{} // 统计结果
		sh  = f.Sheets[0]
		ch  = make(chan *welNode, sh.MaxRow)
		wg  = new(sync.WaitGroup)
		err error
	)

	// 导入标准
	// 优先使用指定文件夹下的csv外置标准文件
	if len(task.StdFile) > 0 {
		stdlimits, err = std.LimitsGen(task.StdFile)
		if err != nil {
			return err
		}
	} else {
		stdlimits = std.ParseStd(std.Wel)
	}

	wg.Add(sh.MaxRow - p.dataRow)
	for row := p.dataRow; row < sh.MaxRow; row++ {
		data := sh.Rows[row]
		go p.handleRow(data, ch, wg)
	}

	chErr := make(chan error)
	go func() {
		err := welSave(task.Output, ch, res)
		task.Detail = res.detail()
		chErr <- err
	}()

	wg.Wait()
	close(ch)

	return <-chErr
}

func initWelParam() *welParam {
	return &welParam{
		param: &param{
			keyRow:  0,
			keyCol:  15,
			keyLen:  39,
			nameCol: 8,
			dataRow: 3,
		},
		useCol:  13,
		townCol: 6,
	}
}

func (p *welParam) handleRow(row *xlsx.Row, chNode chan<- *welNode, wg *sync.WaitGroup) {
	defer wg.Done()
	var (
		levels       = make([]int, p.keyLen)
		ltLimitKeys  = make([]string, 0)
		noHandleKeys = make([]string, 0)
		noExistKeys  = make([]string, 0)
		errStdValues = make([]string, 0)
	)
	oneNode := &welNode{node: &node{}}
	oneNode.name = row.Cells[p.nameCol].String()
	oneNode.town = row.Cells[p.townCol].String()
	oneNode.use = row.Cells[p.useCol].String()

	for i := p.keyCol; i < p.keyCol+p.keyLen; i++ {
		oneKey := p.handleSinKey(row, i)
		levels = append(levels, oneKey.lv.level)
		oneNode.info = oneNode.info + oneKey.info
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
	oneNode.err = oneNode.err + err(ltLimitKeys, "指标：", "检出限大于一类标准限值;")
	oneNode.err = oneNode.err + err(noHandleKeys, "指标：", "未处理(转换float失败),检查是否录入错误;")
	oneNode.err = oneNode.err + err(noExistKeys, "指标：", "未处理(未查询到标准值,检查指标名称是否正确);")
	oneNode.err = oneNode.err + err(errStdValues, "指标：", "转换float失败,结果评价可能有错;")
	chNode <- oneNode
}

func welSave(file string, nodes <-chan *welNode, res *result) error {
	f, err := xlsx.OpenFile(file)
	if err != nil {
		f = xlsx.NewFile()
	}
	var (
		sh *xlsx.Sheet
	)
	for i, sh := range f.Sheets {
		if sh.Name == "地下水" {
			f.Sheets = append(f.Sheets[:i], f.Sheets[i+1:]...)
			delete(f.Sheet, "地下水")
			break
		}
	}
	sh, err = f.AddSheet("地下水")
	if err != nil {
		return errors.New("保存结果文件:\n" + file + "\n创建表格失败\n" + err.Error())
	}
	for node := range nodes {
		res.total++
		if node.level < 4 {
			res.pass++
		}
		row := sh.AddRow()
		row.AddCell().Value = node.town
		row.AddCell().Value = node.name
		row.AddCell().Value = node.use
		row.AddCell().Value = levelToString(node.level)
		row.AddCell().Value = node.info
		warnCell := row.AddCell()
		warnCell.Value = node.err
	}

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
func (res *result) detail() string {
	return fmt.Sprintf("共处理点位: %d 个,超标: %d 个,达标率: %s\n",
		res.total, res.total-res.pass, percent(res.total, res.pass))
}

func (p *welParam) handleSinKey(f *xlsx.Row, col int) *sinKey {
	var (
		sk    = &sinKey{lv: &level{}}
		key   = f.Sheet.Rows[p.keyRow].Cells[col].String()
		value = strings.TrimSpace(f.Cells[col].String())
		e     error
	)
	key = formatKey(key)
	sk.key = key
	// handle Wel 嗅和味
	// handle Wel 肉眼可见物
	if key == "嗅和味" || key == "肉眼可见物" {
		return sk.handleSmSo(value)
	}

	if e = sk.parseFloat(value); e != nil {
		return sk
	}
	// pH
	if strings.ToLower(key) == "ph" {
		sk.welPH(sk.fvalue)
		return sk
	}
	limits, ok := (*stdlimits)[key]
	// 检查指标是否在标准中的存在
	if !ok {
		sk.err = noExistKey
		return sk
	}
	// handle 未检出
	if value[len(value)-1:] == "L" {
		if key == "阴离子表面活性剂" {
			sk.lv.level = 1
			return sk
		}
		sk.handleND(limits[0])
		return sk
	}
	// 正常情况
	return sk.normal(limits)
}

func (sk *sinKey) welPH(fvalue float64) *sinKey {
	if fvalue >= 6.5 && fvalue <= 8.5 {
		sk.lv.level = 1
	} else if (fvalue >= 5.5 && fvalue < 6.5) || (fvalue > 8.5 && fvalue <= 9) {
		sk.lv.level = 4
		sk.info = "pH(Ⅳ);"
	} else {
		sk.lv.level = 5
		sk.info = "pH(Ⅴ);"
	}
	return sk
}
func (sk *sinKey) handleSmSo(value string) *sinKey {
	sk.lv.level = 1
	if value == "有" {
		sk.info = sk.key + "(Ⅴ);"
		sk.lv.level = 5
	}
	return sk
}
