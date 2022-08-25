package ttime

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkTimeWheel_AddTimer(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	tw := NewTimeWheel(WithInterval(time.Second*1), WithSlotNum(3600))
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		tw.AddTimer(time.Second*time.Duration(rand.Intn(10)), key, func(kv ...interface{}) {
			fmt.Println("args ", kv, " now ", time.Now().Unix())
		}, i)
	}
	tw.start()
	// defer tw.Stop()
}

func TestParse(t *testing.T) {
	t1, _ := Parse("2006-01-02 15:04:05", "2020-01-02 13:03:01")
	fmt.Println(t1)
}
