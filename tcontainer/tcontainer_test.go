package tcontainer

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

var (
	tmap = NewSafeMap() //NewTreeMap() //NewSafeMap()
	sp   = new(sync.Map)
)

func initContainer() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 1000000; i++ {
		tmap.Set(i, rand.Int31n(1+int32(i)))
		sp.Store(i, rand.Int31n(1+int32(i)))
	}
}

// BenchmarkTreeMap_Del-10    	13885704	        85.18 ns/op	      32 B/op	       2 allocs/op
// BenchmarkSafeMap_Del-10    	20725195	        55.16 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSafeMap_Del-10    	64683204	        18.56 ns/op	       0 B/op	       0 allocs/op
func BenchmarkSafeMap_Del(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tmap.Del(i)
	}

}

// BenchmarkSync_map-16    	 9340906	       121.4 ns/op	       0 B/op	       0 allocs/op
func BenchmarkSync_map(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sp.Delete(i)
	}
}

// BenchmarkSafeMap_add-10    	 4940218	       325.0 ns/op	      91 B/op	       4 allocs/op
func BenchmarkSafeMap_add(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tmap.Set(i, i)
	}
}

// BenchmarkSync_add-10    	 6139191	       355.2 ns/op	     115 B/op	       4 allocs/op
func BenchmarkSync_add(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sp.Store(i, i)
	}
}

// BenchmarkSafeMap_range-16    	      78	  15054213 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSafeMap_range-16    	      26	  41799357 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSafeMap_range-10    	     142	   8347134 ns/op	       0 B/op	       0 allocs/op
//BenchmarkSafeMap_range-10    	       1	9594401500 ns/op	      24 B/op	       3 allocs/op

func BenchmarkSafeMap_range(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < 1000; i++ {
		tmap.Range(func(key, value interface{}) bool {
			return true
		})
	}
}

// BenchmarkSync_range-16    	      12	  97929962 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSync_range-16    	       9	 119385416 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSync_range-16    	       9	 120863920 ns/op	       0 B/op	       0 allocs/op
//BenchmarkSafeMap_range-10    	       1	9594401500 ns/op	      24 B/op	       3 allocs/op
//BenchmarkSafeMap_range-10    	       1	8914126125 ns/op	      24 B/op	       3 allocs/op
//BenchmarkSync_range-10    	       1	35468923292 ns/op	      40 B/op	       4 allocs/op

func BenchmarkSync_range(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < 1000; i++ {
		sp.Range(func(key, value interface{}) bool {
			// fmt.Println(key, value)
			return true
		})
	}
}
