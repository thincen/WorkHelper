package core

import "testing"

func Test_Avg(t *testing.T) {
	task := &Task{
		Module: Riv,
		Input:  "../test/riv202008.xlsx",
		Output: "../test/rivTest.xlsx",
	}
	t.Log(task.Run())
}
