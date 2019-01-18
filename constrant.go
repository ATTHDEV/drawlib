package drawlib

import (
	"path"
	"runtime"
)

var (
	ANGSANA  string
	ARIAL    string
	BROWALIA string
	CORDIA   string
	TAHOMA   string
)

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	ANGSANA = path.Dir(filename) + "/resource/ANGSANA.TTC"
	ARIAL = path.Dir(filename) + "/resource/ARIAL.TTF"
	BROWALIA = path.Dir(filename) + "/resource/BROWALIA.TTC"
	CORDIA = path.Dir(filename) + "/resource/CORDIA.TTC"
	TAHOMA = path.Dir(filename) + "/resource/TAHOMA.TTF"
}
