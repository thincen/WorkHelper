package core

import (
	"testing"
)

func Test_InitParam(t *testing.T) {
	s := initSurParam()
	t.Logf("%v", s.keyRow)
}

func Test_HandleSur(t *testing.T) {
	task := &Task{
		Module: Sur,
		Input:  "../test/sur.xlsx",
		Output: "../test/surTest.xlsx",
	}
	t.Log(task.Run())
}
