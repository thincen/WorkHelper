# WorkHelper

基于golang实现的`个人`工作助手, 快速处理检测点位等级评价和超标统计.支持乡镇集中式饮用水(地表水型/地下水型)、河流断面的数据分析统计.

## 评价标准

```golang
   task:=&workhelper.Task{
       ...,
       stdFile:"path/to/std/file",
   }
```

`Task.StdFile`指定标准文件时,优先加载外部标准文件:

- 模板文件夹位置: `stdTpl/`
- 地表水: `stdTpl/sur.csv`
- 地下水: `stdTpl/wel.csv`  

内部默认标准不适用时 **可手工修改** 模板文件内容保证正常运行

当前默认使用标准:

- 地表水质量标准(GBT3838-2002)
- 地下水质量标准(GBT14848-2017)

## 分析指标

### 地表水

地表水环境质量标准(GB3838-2002)表1的基本项目

- 综合评价: 除水温、总氮、粪大肠菌群外所有指标均参与评价
- 参考指标单独评价: 河流型水源地仅评价粪大肠菌群;湖库性水源地评价总氮和粪大肠菌群

### 地下水

地下水质量标准(GB/T 14848-2017)表1中感官性状及一般化学指标、微生物指标、毒理学指标、放射性指标共39项

## 上报模板

### 饮用水

- 地表水型: `report-tpl/sur.xlsx`
- 地下水型: `report-tpl/wel.xlsx`
- 交界断面: `report-tpl/riv.xlsx`

## 结果保存

默认为选定上报表格所在路径下 `res.xlsx` 文件

## 使用示例

### 内部默认标准

```golang
package main

import (
	"fmt"
	wh "github.com/thincen/WorkHelper"
)

func main() {
	inputFile := "path/to/input/file"
	outputFile := "path/to/output/file"
	task := wh.NewTask(inputFile, outputFile, wh.Sur)
	if err := task.Run(); err != nil {
		fmt.Println(err)
	}
}
```