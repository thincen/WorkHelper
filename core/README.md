# packge core

学习golang基础，写个帮助工作的小玩具练习入门

## 使用自定义外部标准文件

内部默认标准不适用双不方便修改源码时 **手工修改** 模板文件内容保证正确处理数据.

修改`../bin/template/stdTpl`文件夹内csv文件

场景: 地表水国家标准中溶解指标的Ⅰ类标准限值修改7.5为8

|指标名称|Ⅰ类|Ⅱ类|Ⅲ类|Ⅳ类|Ⅴ类|
|---|---|---|---|---|---|
|溶解氧|7.5|6|5|3|2|
|...|...|...|...|...|...|

`../bin/template/stdTpl/sur.csv`中修改

```diff
...
-溶解氧,7.5,6,5,3,2
+溶解氧,8,6,5,3,2
...
```

```golang
   task:=&workhelper.Task{
       ...,
       stdFile:"path/to/std/file",
   }
```

`Task.StdFile`指定标准文件时,优先加载外部标准文件.

## 内部默认标准

```golang
package main

import (
	"fmt"
	"github.com/thincen/workHelper/core"
)

func main() {
	inputFile := "path/to/input/file"
	outputFile := "path/to/output/file"
	task := core.NewTask(inputFile, outputFile, core.Sur)
	if err := task.Run(); err != nil {
		fmt.Println(err)
	}
}
```
