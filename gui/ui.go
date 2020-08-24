package main

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/thincen/utils"
	"github.com/thincen/workHelper/core"
)

type mwin struct {
	*walk.MainWindow
	lb                 *walk.Label
	leInput            *walk.LineEdit
	leOutput           *walk.LineEdit
	cb                 *walk.ComboBox
	showAboutBoxAction *walk.Action
	width              int
	height             int
	checkbox           *walk.CheckBox
	stdFile            string
}

// 菜单栏
func (mw *mwin) menuInit() *[]MenuItem {
	return &[]MenuItem{
		Menu{
			Text: "&Help",
			Items: []MenuItem{
				Action{
					AssignTo:    &mw.showAboutBoxAction,
					Text:        "&About",
					OnTriggered: mw.showAboutBoxActionTriggered,
				},
			},
		},
	}
}

// 帮助信息
func (mw *mwin) usage() *Composite {
	usage := &Composite{
		Layout: VBox{
			MarginsZero: true,
			Alignment:   AlignHNearVNear,
		},
		Children: []Widget{
			Label{
				AssignTo: &mw.lb,
				Text: `1.修改template文件夹中模板文件点位及检测数据
【注意】不要修改模板原有格式
2. 点击Open/拖放 选择上报模板xlsx文件
3. 选择结果存放位置(默认为第1步选择文件所在路径下生成res.xlsx/RivRes.xlsx文件)
4. Run`,
			},
		},
	}
	return usage
}

// 选择模块
func (mw *mwin) selectFunc() *Composite {
	return &Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			Label{Text: "数据属性:"},
			// 下拉选项
			ComboBox{
				AssignTo:     &mw.cb,
				CurrentIndex: 2,
				Model:        []string{"地表水型饮用水源地", "地下水型饮用水源地", "交界断面"},
				OnCurrentIndexChanged: func() {
					tag = mw.cb.CurrentIndex()
				},
			},
			CheckBox{
				AssignTo:         &mw.checkbox,
				Text:             "自定义标准文件(慎用)",
				OnCheckedChanged: mw.checkboxOnChanged,
				ToolTipText:      "建议使用template中标准文件修改",
			},
			// 空格占满
			VSpacer{},
		},
	}
}

// 自定义外部标准文件
func (mw *mwin) checkboxOnChanged() {
	if mw.checkbox.CheckState() == win.BST_CHECKED {
		dlg := new(walk.FileDialog)
		dlg.Title = "选择文件"
		dlg.Filter = "标准文件 (*.csv)|*.csv|所有文件|*.*"
		accepted, err := dlg.ShowOpen(mw)
		if err != nil {
			walk.MsgBox(mw, "OpenDlgErr", err.Error(), walk.MsgBoxIconError)
			return
		}
		if !accepted {
			mw.checkbox.SetChecked(false)
		} else {
			mw.stdFile = dlg.FilePath
			mw.checkbox.SetToolTipText(dlg.FilePath)
		}
	} else {
		mw.stdFile = ""
		mw.checkbox.SetToolTipText("建议使用template中标准文件修改")
	}
}

// 选择路径
func (mw *mwin) selectPath() *Composite {
	return &Composite{
		Layout:  Grid{MarginsZero: true, Columns: 2},
		MaxSize: Size{Width: mw.width},
		Children: []Widget{
			Label{Text: "选择要处理的数据xlsx文件:"},
			VSplitter{},
			LineEdit{
				AssignTo:          &mw.leInput,
				OnEditingFinished: mw.inputChanged,
				MinSize:           Size{Width: 200},
				// OnTextChanged: mw.inputChanged,
			},
			PushButton{
				Text:      "Open",
				MaxSize:   Size{Height: 25},
				OnClicked: mw.selectFile,
			},
			Label{Text: "选择结果保存位置:"},
			VSplitter{},
			LineEdit{
				AssignTo: &mw.leOutput,
				// OnEditingFinished: mw.outputChanged,
			},
			PushButton{Text: "Save", MaxSize: Size{Height: 25}, OnClicked: mw.saveFile},
		},
	}
}

// 设置窗口属性
func (mw *mwin) show() {
	// 设置 ^win.WS_MAXIMIZEBOX 禁用最大化按钮
	// 设置 ^win.WS_THICKFRAME 禁用窗口大小改变
	win.SetWindowLong(mw.Handle(), win.GWL_STYLE, win.GetWindowLong(mw.Handle(), win.GWL_STYLE) & ^win.WS_MAXIMIZEBOX & ^win.WS_THICKFRAME)
	win.SetWindowPos(mw.Handle(), 0, (win.GetSystemMetrics(win.SM_CXSCREEN)-600)/2, (win.GetSystemMetrics(win.SM_CYSCREEN)-400)/2, 300, 200, win.SWP_FRAMECHANGED)
	mw.Run()
}

// 选择要处理的文件
func (mw *mwin) selectFile() {
	dlg := new(walk.FileDialog)
	dlg.Title = "选择文件"
	dlg.Filter = "上报表格 (*.xlsx)|*.xlsx"
	// mw.leInput.SetText("") //通过重定向变量设置TextEdit的Text
	accepted, err := dlg.ShowOpen(mw)
	if err != nil {
		walk.MsgBox(mw, "OpenDlgErr", err.Error(), walk.MsgBoxIconError)
		return
	}
	if accepted {
		mw.leInput.SetText(dlg.FilePath)
		mw.inputChanged()
	}
}

func (mw *mwin) inputChanged() {
	var (
		inputfiledir = filepath.Dir(mw.leInput.Text())
		outputfile   string
	)
	if tag != 2 {
		outputfile = inputfiledir + "\\res.xlsx"
	} else {
		outputfile = inputfiledir + "\\RivRes.xlsx"
	}
	if mw.leOutput.Text() == "" {
		mw.leOutput.SetText(outputfile)
	} else if inputfiledir != filepath.Dir(mw.leOutput.Text()) {
		if walk.MsgBox(mw, "文件保存位置", "处理文件的路径已更改,是否生成默认保存位置?\n\n默认保存在要处理表格文件同级目录下", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == 6 {
			mw.leOutput.SetText(outputfile)
		}
	}
}

func (mw *mwin) saveFile() {
	dlg := new(walk.FileDialog)
	dlg.Title = "选择存放结果的文件夹"
	if _, err := dlg.ShowBrowseFolder(mw); err != nil {
		walk.MsgBox(mw, "SaveDlgErr", err.Error(), walk.MsgBoxIconError)
		return
	}
	mw.leOutput.SetText(dlg.FilePath + "\\res.xlsx")
}

// Run 处理数据
func (mw *mwin) handleData() {
	var (
		inputfile  = mw.leInput.Text()
		outputfile = mw.leOutput.Text()
	)
	if !mw.checkinput(inputfile) {
		return
	}
	// task := core.NewTask(inputfile, outputfile, core.Tag(tag))
	task := &core.Task{
		Input:   inputfile,
		Output:  outputfile,
		StdFile: mw.stdFile,
		Module:  core.Tag(tag),
	}
	if err := task.Run(); err != nil {
		walk.MsgBox(mw, "error", err.Error(), walk.MsgBoxIconError)
		return
	}
	res := "处理完毕\n\n" + task.Detail + "\n详细信息保存在文件：" + outputfile
	if walk.MsgBox(mw, "完成", res+"\n\n是否打开查看结果？", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == 6 {
		cmd := exec.Command("explorer", outputfile)
		err := cmd.Start()
		if err != nil {
			walk.MsgBox(mw, "打开失败", err.Error(), walk.MsgBoxIconError)
		}
	}
}

func (mw *mwin) showAboutBoxActionTriggered() {
	// 	walk.MsgBox(mw, "About",
	// 		NAME+`
	// 版本号: `+VERSION+`
	// 构建时间: `+BUILDDAY+`
	// 反馈: no_1seed@163.com`+`
	// 仓库地址: https://github.com/thincen/workHelper`,
	// 		0)
	Dialog{
		Title: "关于",
		Layout: VBox{
			Margins: Margins{
				Top:    10,
				Right:  30,
				Bottom: 30,
				Left:   30,
			},
			Alignment: AlignHNearVNear,
		},
		Font: Font{
			PointSize: 10,
			Family:    "微软雅黑",
			// StrikeOut: true,
		},
		// MinSize: Size{Width: 300, Height: 300},
		Children: []Widget{
			Label{
				Text: NAME,
				Font: Font{
					Family:    "微软雅黑",
					PointSize: 12,
					Bold:      true,
				},
			},
			Label{Text: "版本号: " + VERSION},
			Label{Text: "构建时间: " + BUILDDAY},
			LinkLabel{
				Text: `反馈: <a id="mail" href="mailto:no_1seed@163.com">no_1seed@163.com</a>`,
				OnLinkActivated: func(link *walk.LinkLabelLink) {
					cmd := exec.Command("explorer", "mailto:no_1seed@163.com")
					err := cmd.Start()
					if err != nil {
						walk.MsgBox(mw, "反馈", err.Error(), walk.MsgBoxIconError)
					}
				},
			},
			LinkLabel{
				Text: `仓库地址：<a href="https://github.com/thincen/workHelper">https://github.com/thincen/workHelper</a>`,
				OnLinkActivated: func(link *walk.LinkLabelLink) {
					cmd := exec.Command("explorer", "https://github.com/thincen/workHelper")
					err := cmd.Start()
					if err != nil {
						walk.MsgBox(mw, "打开项目仓库失败", err.Error(), walk.MsgBoxIconError)
					}
				},
			},
			VSpacer{},
		},
	}.Run(mw)
}

func (mw *mwin) checkinput(file string) bool {
	if len(file) == 0 {
		walk.MsgBox(mw, "选择文件错误", "要处理的文件不能为空", walk.MsgBoxIconError)
		return false
	}
	if !strings.HasSuffix(filepath.Base(file), ".xlsx") {
		walk.MsgBox(mw, "选择文件错误", "仅支持饮用水上报表格文件xlsx类型\n请重新选择要处理的文件", walk.MsgBoxIconError)
		// mw.leInput.SetToolTipText("仅支持饮用水上报表格文件xlsx类型\n请重新选择要处理的文件")
		mw.leInput.SetText("")
		return false
	}
	if !utils.IsExist(file) {
		walk.MsgBox(mw, "选择文件错误", "要处理的文件\n"+file+" 不存在", walk.MsgBoxIconError)
		return false
	}
	if tag == 0 && strings.Contains(file, "地下水") {
		r := walk.MsgBox(mw, "文件匹配确认", "当前为\"地表水\"处理模式\n处理文件可能为\"地下水\"数据,确认按照\"地表水\"处理?", walk.MsgBoxYesNo|walk.MsgBoxIconWarning)
		if r == 6 {
			// 用户确认不修改
		} else {
			mw.cb.SetCurrentIndex(1)
		}
		return true
	}
	if tag == 1 && strings.Contains(file, "地表水") {
		if walk.MsgBox(mw, "文件匹配确认", "当前为\"地下水\"处理模式\n处理文件可能为\"地表水\"数据,确认按照\"地下水\"处理?", walk.MsgBoxYesNo|walk.MsgBoxIconWarning) == 6 {
			// 用户确认不修改
		} else {
			mw.cb.SetCurrentIndex(0)
		}
		return true
	}
	return true
}
