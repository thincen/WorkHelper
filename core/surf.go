package core

import "github.com/tealeg/xlsx"

type surf struct {
	*defaultHandler
}

// Handle surface
func (d *surf) Handle() *xlsx.File {

	return nil
}

// Marshal web template
func (d *surf) Marshal(s ...string) interface{} {
	return nil
}