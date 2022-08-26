package tlog

import (
	"github.com/pkg/errors"
	"os"
	"testing"
)

func BenchmarkInfo(b *testing.B) {
	v := 0
	b.StartTimer()
	defer b.StopTimer()
	b.ReportAllocs()
	for idx := 0; idx < b.N; idx++ {
		v = v + 1
		Info("12344", "zhangsa", "zhangsa", "zhangsa", 123, "cndkcd", "ncjdncd", 12345432)
		//Debug("BenchmarkInfo","zhangsa","njcdcd","index",v)
		//Error("BenchmarkInfo",err,"cjdncjd","njcdcd","index",v)
		//Warn("BenchmarkInfo","zhangsa","njcdcd","index",v)
	}
	defer Closed()
}

func TestDebug(t *testing.T) {
	pwd, _ := os.Getwd()
	ll := NewTBaseLog(WithLogDir(pwd+"/"), WithPrefixName("defalut_log"), WithProjectName("TLog"), WithTimeLayout(LogDateFmt), WithLogLeval(logLevalWarn))
	ll.Info("12345", 222)
	ll.Error(errors.New("test1"), "jdcjdjc", "ddd")
	defer Closed()
}
