package workhelper

import "testing"

func Test_InitParam(t *testing.T) {
	s := initSurParam()
	t.Logf("%v", s.keyRow)
}
