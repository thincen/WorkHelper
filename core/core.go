package core

import (
	"errors"

	"github.com/tealeg/xlsx"
	"github.com/thincen/workHelper/std"
)

// Option 设置参数
type Option func(*defaultHandler) error

// DataHandler & return
type DataHandler interface {
	Handle() *xlsx.File
	Marshal(s ...string) interface{} // for web
}

type defaultHandler struct {
	xlsx     *xlsx.File
	setLevel Level    // 使用等级评价
	banKeys  []string // 不评价的指标
	keys     map[string]string
}

func newDefaultHandler() *defaultHandler {
	return &defaultHandler{
		setLevel: L3,
		banKeys:  []string{"水温"},
		keys:     make(map[string]string),
	}
}

// NewDataHandler - interface
// setLevel default L3
func NewDataHandler(tag string, options ...Option) (DataHandler, error) {
	d := newDefaultHandler()

	for _, option := range options {
		if e := option(d); e != nil {
			return nil, e
		}
	}
	switch tag {
	case "well":
		return &well{d}, nil
	case "surf":
		return &surf{d}, nil
	case "riv":
		d.banKeys = append(d.banKeys, "总氮(河流)")
		return &riv{
			defaultHandler: d,
			p:              newDefaultRivParam(),
			limits:         std.ParseStd(std.Sur),
			res:            make(map[int]*result),
		}, nil
	}
	return nil, errors.New("错误的功能模块")
}

// WithLevel 设置使用的等级评价
func WithLevel(l string) Option {
	var setLevel Level
	switch l {
	case "l1":
		setLevel = L1
	case "l2":
		setLevel = L2
	case "l3":
		setLevel = L3
	case "l4":
		setLevel = L4
	case "l5":
		setLevel = L5
	default:
		setLevel = L3
	}
	return func(d *defaultHandler) error {
		d.setLevel = setLevel
		return nil
	}
}

// WithBanKeys 设置不参与评价的指标
func WithBanKeys(key ...string) Option {
	return func(d *defaultHandler) error {
		d.banKeys = append(d.banKeys, key...)
		return nil
	}
}

// WithBytes 加载 byte
func WithBytes(b []byte) Option {
	return func(d *defaultHandler) error {
		xlsx, err := byte2Xlsx(b)
		if err != nil {
			return errors.New("解析上传文件失败")
		}
		d.xlsx = xlsx
		return nil
	}
}

// WithFile 加载 本地文件
func WithFile(f string) Option {
	return func(d *defaultHandler) error {
		xlsx, err := xlsx.OpenFile(f)
		if err != nil {
			return err
		}
		d.xlsx = xlsx
		return nil
	}
}
