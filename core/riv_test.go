package core

import (
	"testing"
)

func Test_riv_Marshal(t *testing.T) {
	d, e := NewDataHandler("riv", WithFile("/home/lgx/github/thincen/WorkHelper/core/test/riv-app0.xlsx"))
	if e != nil {
		t.Fatal(e)
	}
	d.Handle()
	res := d.Marshal()
	if r, ok := res.([]rivRes); ok {
		t.Log("len: ", len(r))
		t.Log(r[0].Up["cb"])
	}
	// t.Log([]rivRes(res)[0].Up)
}
