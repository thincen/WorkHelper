package core

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tealeg/xlsx"
	"github.com/thincen/whserver/std"
)

// web data
type (
	sitRes map[string]string // 每行
	rivRes struct {          // 每个断面
		Up, Down, Avg sitRes
		Ltd           []string
	}
)

type riv struct {
	*defaultHandler
	limits *std.Limits
	p      *rivParam
	res    map[int]*result
}
type rivParam struct {
	*param
	originCol int // 数据来源所在列
	errCol    int // 错误信息所在列
	levelCol  int // 实测类别列
	detailCol int // 超标信息列
	coliform  int // 粪大肠菌群超标信息列
}

// result 每行评价结果
type result struct {
	isHandle Level    // 是否检测
	maxLevel Level    // 实测类别
	detail   []string // 超标信息
	coliform string   // 粪大肠菌群超标信息
	err      warn     // 错误信息
}

// Handle river
func (d *riv) Handle() *xlsx.File {
	// 1.格式化指标名称和测试值
	d.format()
	// 2. 计算均值栏
	d.xlsx = d.calcAverage()
	// 3. 评价和错误信息
	for row := d.p.dataRow; row < d.xlsx.Sheets[0].MaxRow; row++ {
		d.xlsx = d.handleRow(row)
	}
	return d.xlsx
}

// Marshal web template return []rivRes
func (d *riv) Marshal(ltds ...string) interface{} {
	var n = (d.xlsx.Sheets[0].MaxRow - d.p.dataRow + 1) / 3 // 断面数
	var tpl = make([]rivRes, n)
	var wrap = func(data *xlsx.Row) sitRes {
		var sit = make(sitRes)
		for col := 0; col < d.xlsx.Sheets[0].MaxCol-3; col++ {
			key := data.Sheet.Row(d.p.keyRow).Cells[col].String()
			switch {
			case strings.EqualFold("pH", key):
				key = "pH"
			case strings.HasPrefix(key, "粪大肠"):
				key = "粪大肠菌群"
			}
			sit[key] = data.Cells[col].String()
		}
		sit["cb"] = data.Cells[d.p.detailCol].String()
		sit["dc"] = data.Cells[d.p.coliform].String()
		sit["warn"] = data.Cells[d.p.errCol].String()
		return sit
	}
	for r := 0; r < n; r++ {
		rivRes := new(rivRes)
		for row := r*3 + d.p.dataRow; row < (r+1)*3+d.p.dataRow; row++ {
			switch (row - d.p.dataRow) % 3 {
			case 0:
				rivRes.Up = wrap(d.xlsx.Sheets[0].Row(row))
			case 1:
				rivRes.Down = wrap(d.xlsx.Sheets[0].Row(row))
			case 2:
				rivRes.Avg = wrap(d.xlsx.Sheets[0].Row(row))
			}
		}
		if len(ltds) > 0 {
			for _, ltd := range ltds {
				if strings.Contains(rivRes.Avg["测站名称"], string([]rune(ltd)[:2])) {
					rivRes.Ltd = append(rivRes.Ltd, ltd)
				}
			}
		}
		tpl[r] = *rivRes
	}
	return tpl
}

// 评价
func (d *riv) handleRow(row int) *xlsx.File {
	var sh *xlsx.Sheet
	// 只处理第一张表
	sh = d.xlsx.Sheets[0]
	r, ok := d.res[row]
	if !ok {
		r = &result{
			detail:   make([]string, 0),
			err:      make(warn),
			isHandle: noHandleLevel,
		}
	}
	r.coliform = "/"
	var key, value string
	data := sh.Rows[row]
	var (
		starCol = d.p.keyCol + 1 // skip temp
		endCol  = d.p.keyLen + starCol
	)
	for i := starCol; i < endCol; i++ {
		// 清空结果信息
		setResultCell(data, d.p.levelCol-1, d.setLevel.string())
		setResultCell(data, d.p.levelCol, "")
		setResultCell(data, d.p.detailCol, "/")
		setResultCell(data, d.p.coliform, "/")
		setResultCell(data, d.p.errCol, "")
		key = d.keys[sh.Rows[d.p.keyRow].Cells[i].String()]
		value = data.Cells[i].String()
		d.res[row] = d.handleKey(key, value, r)
	}
	if d.res[row].isHandle < noHandleLevel{
		// 设置此行实测类别
		setResultCell(data, d.p.levelCol, d.res[row].maxLevel.string())
	}else{
		setResultCell(data, d.p.levelCol, d.res[row].isHandle.string())
	}
	// 设置超标评价信息
	setResultCell(data, d.p.detailCol, d.res[row].detail...)
	// 设置细菌评价信息
	data.Cells[d.p.coliform].SetString(d.res[row].coliform)
	// 设置错误
	var resErr = make([]string, 0)
	for k, v := range d.res[row].err {
		resErr = append(resErr, appendString(k.Error(), "(", strings.Join(v, ","), ")"))
	}
	setResultCell(data, d.p.errCol, resErr...)
	return d.xlsx
}

func (d *riv) handleKey(key, value string, r *result) *result {
	// skip banKeys
	if isBanned(key, d.banKeys) {
		return r
	}
	if value == "-1" {
		return r
	}
	var wrap = func(l Level, e error) *result {
		if e != nil {
			if ec, ok := e.(errCode); ok {
				r.err[ec] = append(r.err[ec], key)
			}
			return r
		}
		// 排除errLevel
		if l <= L6 {
			r.isHandle = l
			if l > r.maxLevel && !strings.Contains(key, "大肠") {
				r.maxLevel = l
			}
			if l > d.setLevel {
				var detail string
				switch {
				case strings.HasPrefix(strings.ToLower(key), "ph"):
					detail = appendString("pH(", l.string(), ")")
					r.detail = append(r.detail, detail)
				case strings.Contains(key, "大肠"):
					mul := calcMul(value, (*d.limits)[key][d.setLevel])
					r.coliform = appendString(key, "(", l.string(), ",超标", mul, "倍)")
				default:
					mul := calcMul(value, (*d.limits)[key][d.setLevel])
					detail := appendString(key, "(", l.string(), ",超标", mul, "倍)")
					r.detail = append(r.detail, detail)
				}
			}
		}
		return r
	}
	// pH
	if strings.HasPrefix(strings.ToLower(key), "ph") {
		return wrap(pH(value, Riv))
	}
	limits, ok := (*d.limits)[key]
	// 标准中未找到key
	// 总氮(河流) 已加入banKeys
	if !ok {
		r.err[errNoExistKey] = append(r.err[errNoExistKey], key)
		return r
	}
	if strings.HasPrefix(key, "溶解氧") {
		return wrap(o2(value, limits))
	}
	// 未检出
	if strings.Contains(value, "L") {
		return wrap(noHandleLevel, checkND(value, limit(limits[0])))
	}
	return wrap(nomalKey(value, limits))
}

// 规范指标名称、计算均值、未填写-->-1 检查录入数据
func (d *riv) format() {
	var sh *xlsx.Sheet
	// 只处理第一张表
	sh = d.xlsx.Sheets[0]
	var (
		starCol = d.p.keyCol
		endCol  = d.p.keyLen + d.p.keyCol + 1
	)
	keys := sh.Rows[d.p.keyRow].Cells[starCol:endCol]
	for _, key := range keys {
		d.keys[key.String()] = formatKey(key.String(), Riv)
	}
	var sliceValue []*xlsx.Cell
	// sh.MaxRow-2 排除最后一行均值
	for i := d.p.dataRow; i < sh.MaxRow-1; i++ {
		sliceValue = sh.Rows[i].Cells[starCol:endCol]
		for id, value := range sliceValue {
			// if value.Row.Sheet.Row(d.p.keyRow).Cells[]
			if sh.Row(d.p.keyRow).Cells[id+starCol].String() == "水温" && value.String() == "" {
				sliceValue[id].SetString("/")
			}
			sliceValue[id].SetString(formatValue(value.String()))
		}
	}
}

// newDefaultRivParam 默认的参数
func newDefaultRivParam() *rivParam {
	p := &rivParam{
		param: &param{
			keyRow:  1,
			dataRow: 2,
			keyCol:  7,  // 水温列开始
			keyLen:  23, // 不包含水温
			nameCol: 1,
		},
		originCol: 3,
		levelCol:  32,
		errCol:    35,
		detailCol: 33,
		coliform:  34,
	}
	return p
}

func (d *riv) calcAverage() *xlsx.File {
	var sh *xlsx.Sheet
	// 只处理第一张表
	sh = d.xlsx.Sheets[0]
	var row int
	var up, down, avgRow *xlsx.Row
	var (
		starCol = d.p.keyCol
		endCol  = d.p.keyLen + d.p.keyCol + 1
	)
	var avg string
	var err error
	for row = d.p.dataRow + 2; row < sh.MaxRow; row += 3 {
		r, ok := d.res[row]
		if !ok {
			r = &result{
				err: make(warn),
			}
		}
		up = sh.Rows[row-2]
		down = sh.Rows[row-1]
		avgRow = sh.Rows[row]
		for col := starCol; col < endCol; col++ {
			upvalue := up.Cells[col].String()
			downvalue := down.Cells[col].String()
			avg, err = average(upvalue, downvalue)
			if err != nil {
				r.err[errAvg] = append(r.err[errAvg], up.Sheet.Rows[d.p.keyRow].Cells[col].String())
			}
			avgRow.Cells[col].SetString(avg)
		}
		d.res[row] = r
	}
	return d.xlsx
}

// 计算均值 四舍六入
func average(up string, down string) (string, error) {
	// 水温未检测
	if up == "/" && down == "/" {
		return "/", nil
	} else if up == "/" && down != "/" {
		return down, nil
	} else if up != "/" && down == "/" {
		return up, nil
	}

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
		return "", errors.New(down + "\n" + err.Error())
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

func setResultCell(row *xlsx.Row, col int, s ...string) {
	if len(row.Cells)-1 < col {
		row.AddCell().SetString(strings.Join(s, ";"))
		return
	}
	if len(s) == 0 {
		row.Cells[col].SetString("/")
		return
	}
	row.Cells[col].SetString(strings.Join(s, ";"))
}
