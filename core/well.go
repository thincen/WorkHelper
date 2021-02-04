package core

import "github.com/tealeg/xlsx"

type well struct {
	*defaultHandler
}

// Handle well
func (d *well) Handle() *xlsx.File {

	return nil
}

// Marshal web template
func (d *well) Marshal(s ...string) interface{} {
	return nil
}
