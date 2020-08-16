package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// 处理类型 0:地表水型 1:地下水型
var (
	tag int
)

func main() {
	mw := new(mwin)
	mw.width = 300
	mw.height = 200
	err := MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "个人工作助手",
		MinSize:  Size{Width: mw.width, Height: mw.height},
		// Size:     Size{Width: 300, Height: 200},
		Icon:   "logo48.ico",
		Layout: VBox{},
		Font: Font{
			PointSize: 9,
			Family:    "微软雅黑",
		},
		OnDropFiles: func(files []string) {
			mw.leInput.SetText(files[0])
		},
		MenuItems: *mw.menuInit(), // 菜单栏
		Children: []Widget{
			*mw.usage(),      // 使用帮助
			VSeparator{},     // 分割线
			*mw.selectFunc(), // 选择模块
			*mw.selectPath(), // 选择路径
			PushButton{
				Text:        "Run",
				MaxSize:     Size{Width: 100},
				OnClicked:   mw.handleData,
				ToolTipText: "处理数据",
			},
		},
	}.Create()
	if err != nil {
		walk.MsgBox(mw, "Create window Error", err.Error(), walk.MsgBoxIconError)
	}
	mw.show()
}
