package workhelper

import (
	"errors"
	"reflect"

	"github.com/tealeg/xlsx"
)

// Task 要执行的任务
type Task struct {
	Module  Tag    // 功能模块
	Input   string // 处理上报表格
	Output  string // 保存结果的csv文件
	StdFile string // 外部标准文件
	Detail  string // 处理完成的汇总结果
}

// NewTask 创建新的分析表格任务,使用内部默认标准,返回*Task
func NewTask(input, output string, mod Tag,) *Task {
	return &Task{
		Module: mod,
		Input:  input,
		Output: output,
	}
}

// Run 执行任务
// 处理input 生成结果文件
func (task *Task) Run(a ...interface{}) error {
	var (
		webView bool = false 
	)
	for _, v := range a {
		switch reflect.ValueOf(v).Kind() {
		case reflect.Bool:
			webView = reflect.ValueOf(v).Bool()
		}
	}

	f, err := xlsx.OpenFile(task.Input)
	if err != nil {
		return errors.New("打开上报表格错误\n" + err.Error())
	}
	return task.handle(f, webView)
}

func (task *Task) handle(f *xlsx.File, webView bool) error {
	switch task.Module {
	case Sur:
		return handleSur(f, task)
	case Wel:
		return handleWel(f, task)
	case Riv:
		return handleRiv(f, task.Input, task.Output, webView)
	default:
		return nil
	}
}
