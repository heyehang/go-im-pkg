package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestHashRing_GetNode(t *testing.T) {
	ring := NewHashRing(2000)

	host1 := "192.168.0.1:8080"
	host2 := "10.19.2.34:9982"
	host3 := "192.168.0.1:3333"
	host4 := "10.19.2.25:1111"
	host5 := "192.168.0.113:8888"
	host6 := "10.19.2.14:1124"

	ring.AddNode(host1, 1)
	ring.AddNode(host2, 1)
	ring.AddNode(host3, 1)
	ring.AddNode(host4, 1)
	ring.AddNode(host5, 1)
	ring.AddNode(host6, 1)

	fmt.Println(ring.GetNode("124"))
	fmt.Println(ring.GetNode("124"))
	fmt.Println(ring.GetNode("124"))

	rand.Seed(time.Now().UnixNano())

	cnt := 1000000
	fmt.Println(" 总数 ", cnt)

	var host1Cnt, host2Cnt, host3Cnt, host4Cnt, host5Cnt, host6Cnt = 0, 0, 0, 0, 0, 0

	getnodeEmptyCnt := 0
	mp := make(map[string]string, 10)

	for i := 0; i < cnt; i++ {
		rnd := rand.Int31n(int32(cnt))
		v := strconv.Itoa(int(rnd)) + strconv.Itoa(i)
		node := ring.GetNode(v)
		mp[v] = node
		switch node {
		case host1:
			host1Cnt += 1
		case host2:
			host2Cnt += 1
		case host3:
			host3Cnt += 1
		case host4:
			host4Cnt += 1
		case host5:
			host5Cnt += 1
		case host6:
			host6Cnt += 1
		default:
			getnodeEmptyCnt += 1
		}
	}

	fmt.Println("mp len ", len(mp))

	fmt.Println("h6 ", float64(host6Cnt)/float64(cnt))
	fmt.Println("h5 ", float64(host5Cnt)/float64(cnt))
	fmt.Println("h4 ", float64(host4Cnt)/float64(cnt))
	fmt.Println("h3 ", float64(host3Cnt)/float64(cnt))
	fmt.Println("h2 ", float64(host2Cnt)/float64(cnt))
	fmt.Println("h1 ", float64(host1Cnt)/float64(cnt))

	total := host6Cnt + host5Cnt + host4Cnt + host3Cnt + host2Cnt + host1Cnt + getnodeEmptyCnt
	fmt.Printf("get host1cnt = %d\n , host2cnt = %d\n , host3cnt = %d\n , host4cnt = %d\n , host5cnt = %d\n , host6cnt = %d\n , defalutcnt = %d \n , h1-defcnt = %d \n , srccnt = %d \n",
		host1Cnt, host2Cnt, host3Cnt, host4Cnt, host5Cnt, host6Cnt, getnodeEmptyCnt, total, cnt)

}

func BenchmarkHashRing_GetNode(b *testing.B) {
	ring := NewHashRing(100)

	host1 := "192.168.0.1:8080"
	host2 := "10.19.2.34:9982"
	host3 := "192.168.0.1:3333"
	host4 := "10.19.2.25:1111"
	host5 := "192.168.0.113:8888"
	host6 := "10.19.2.14:1124"

	ring.AddNode(host1, 1)
	ring.AddNode(host2, 1)
	ring.AddNode(host3, 1)
	ring.AddNode(host4, 1)
	ring.AddNode(host5, 1)
	ring.AddNode(host6, 1)

	rand.Seed(time.Now().UnixNano())

	cnt := 1000000
	fmt.Println(" 总数 ", cnt)

	var host1Cnt, host2Cnt, host3Cnt, host4Cnt, host5Cnt, host6Cnt = 0, 0, 0, 0, 0, 0

	getnodeEmptyCnt := 0

	for i := 0; i < cnt; i++ {
		rnd := rand.Int31n(int32(cnt))
		v := strconv.Itoa(int(rnd))
		node := ring.GetNode(v)
		switch node {
		case host1:
			host1Cnt += 1
		case host2:
			host2Cnt += 1
		case host3:
			host3Cnt += 1
		case host4:
			host4Cnt += 1
		case host5:
			host5Cnt += 1
		case host6:
			host6Cnt += 1
		default:
			getnodeEmptyCnt += 1
		}
	}

	total := host6Cnt + host5Cnt + host4Cnt + host3Cnt + host2Cnt + host1Cnt + getnodeEmptyCnt
	fmt.Printf("get host1cnt = %d\n , host2cnt = %d\n , host3cnt = %d\n , host4cnt = %d\n , host5cnt = %d\n , host6cnt = %d\n , defalutcnt = %d \n , h1-defcnt = %d \n , srccnt = %d \n",
		host1Cnt, host2Cnt, host3Cnt, host4Cnt, host5Cnt, host6Cnt, getnodeEmptyCnt, total, cnt)

}
