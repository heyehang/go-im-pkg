package tlog

import "testing"

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

}
