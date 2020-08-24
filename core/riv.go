// 交界断面

package core

import (
	"errors"
	"strings"
	"sync"

	"github.com/shopspring/decimal"
	"github.com/tealeg/xlsx"
	"github.com/thincen/workHelper/std"
)

type rivParam struct {
	*param
	city int // 监测站名称
}
type rivRow map[string]string
type rivNode struct {
	up   *rivRow // 上游
	down *rivRow // 下游
	ave  *rivRow // 均值
}

func handleRiv(f *xlsx.File, task *Task, webView bool) error {
	var (
		p   = initRivParam()
		sh  = f.Sheets[0]
		ch  = make(chan *surNode, sh.MaxRow)
		wg  = new(sync.WaitGroup)
		err error
	)
	// 计算均值
	maxRow := sh.MaxRow
	var (
		up     = sh.Rows[p.dataRow]
		down   = sh.Rows[p.dataRow]
		avgRow = sh.Rows[p.dataRow]
		avg    string
	)
	for i := p.dataRow; i < maxRow-2; i++ {
		if (i-p.dataRow)%3 == 0 {
			up = sh.Rows[i]
			down = sh.Rows[i+1]
			avgRow = sh.Rows[i+2]
		} else {
			continue
		}
		// keyCol=8 keyLen=23
		for col := p.keyCol - 1; col < p.keyCol+p.keyLen; col++ {
			upvalue := up.Cells[col].String()
			downvalue := down.Cells[col].String()
			if upvalue == "" {
				upvalue = "-1"
			}
			if downvalue == "" {
				downvalue = "-1"
			}
			// up.Cells[col].Value = upvalue
			up.Cells[col].SetString(upvalue)
			down.Cells[col].SetString(downvalue)
			avg, err = average(upvalue, downvalue)
			if err != nil {
				return errors.New("计算均值错误: " + up.Cells[p.nameCol].String() + "\n" + err.Error())
			}
			avgRow.Cells[col].Value = avg
		}
	}
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
		oneNode := &surNode{node: &node{}}
		oneNode.name = sh.Rows[row-(row-p.dataRow)%3].Cells[p.nameCol].String() // xx乡入境断面（中）
		oneNode.town = data.Cells[p.townCol].String()                           // xx监测站
		oneNode.tag = "河流"
		go p.handleRow(data, ch, wg, oneNode)
	}

	chErr := make(chan error)
	go func() {
		err := rivSave(task.Output, ch, f)
		chErr <- err
	}()

	wg.Wait()
	close(ch)

	return <-chErr

}

// 计算均值 四舍六入
func average(up string, down string) (string, error) {

	if up == "-1" && down == "-1" {
		return "-1", nil
	} else if up == "-1" && down != "-1" {
		return down, nil
	} else if up != "-1" && down == "-1" {
		return up, nil
	}
	var (
		upnd   = strings.Contains(up, "L")
		downnd = strings.Contains(down, "L")
		ndot   = countDec(down)
		dup    decimal.Decimal
		ddown  decimal.Decimal
		err    error
	)
	if upnd && downnd { //均未检出
		return down, nil
	} else if upnd && !downnd { //上游未检出
		return averageND(strings.ReplaceAll(up, "L", ""), down, ndot)
	} else if !upnd && downnd { //下游未检出
		return averageND(strings.ReplaceAll(down, "L", ""), up, ndot)
	}
	if dup, err = decimal.NewFromString(up); err != nil {
		return "", errors.New(up + "\n" + err.Error())
	}
	if ddown, err = decimal.NewFromString(down); err != nil {
		return "", errors.New(up + "\n" + err.Error())
	}
	return decimal.Avg(dup, ddown).StringFixedBank(int32(ndot)), nil
}

// 计算单站测量值未检出均值
// ndT:未检出 ndF:检出
func averageND(ndT, ndF string, ndot int) (string, error) {
	var (
		dndT decimal.Decimal
		dndF decimal.Decimal
		err  error
	)
	if dndT, err = decimal.NewFromString(ndT); err != nil {
		return "", err
	}
	dndT = dndT.Div(decimal.NewFromInt(2))
	if dndF, err = decimal.NewFromString(ndF); err != nil {
		return "", err
	}
	return decimal.Avg(dndT, dndF).StringFixedBank(int32(ndot)), nil
}

// 计算小数位数 以下游为准
func countDec(down string) int {
	var (
		ndot int
	)
	down = strings.ReplaceAll(down, "L", "")
	if ndot = strings.Index(down, "."); ndot == -1 {
		ndot = len(down) - 1
	}
	return len(down) - ndot - 1 //0.23 4-1-1=2;16 2-1-1=0
}

func rivSave(file string, nodes <-chan *surNode, f *xlsx.File) error {
	var (
		sh  *xlsx.Sheet
		err error
	)
	sh = f.Sheets[0]
	for node := range nodes {
		for row := 2; row < sh.MaxRow; row++ {
			name := sh.Rows[row-(row-2)%3].Cells[1].String()
			town := sh.Rows[row].Cells[3].String()
			if name == node.name && town == node.town {
				sh.Rows[row].Cells[32].SetString(levelToString(node.level))
				if node.info == "" {
					node.info = "/"
				}
				if node.ps == "" {
					node.ps = "/"
				}
				sh.Rows[row].Cells[33].SetString(node.info)
				sh.Rows[row].Cells[34].SetString(node.ps)
				sh.Rows[row].Cells[35].SetString(node.err)
			}
		}

	}
	err = f.Save(file)
	if err != nil {
		return errors.New("保存结果文件:\n" + file + "\n失败\n" + err.Error())
	}
	return nil
}

func initRivParam() *surParam {
	return &surParam{
		param: &param{
			keyRow:  1,
			keyCol:  8, // 忽略 水温
			keyLen:  23,
			nameCol: 1, // xx乡入境断面（中）
			dataRow: 2,
		},
		townCol:    3,
		surTypeCol: 31,
	}
}
