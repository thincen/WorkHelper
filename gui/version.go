package main

import (
	"time"
)

var (
	// NAME app名称
	NAME string
	// VERSION 版本号
	VERSION string
	// BUILDDAY 构建日期
	BUILDDAY string
)

func init() {
	// NAME = "data-helper-win32"
	// VERSION = "0.1.0"
	BUILDDAY = time.Now().Format("2006-01-02 15:04")
}
