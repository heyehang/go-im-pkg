package tcontainer

import (
	"sync"
)

type SafeMap struct {
	mp  map[interface{}]interface{}
	mux sync.RWMutex
}

func NewSafeMap() *SafeMap {
	mp := new(SafeMap)
	mp.mp = make(map[interface{}]interface{}, 10)
	return mp
}

func (smp *SafeMap) Get(key interface{}) (interface{}, bool) {
	smp.mux.RLock()
	defer smp.mux.RUnlock()
	value, ok := smp.mp[key]
	return value, ok
}

func (smp *SafeMap) Set(key, value interface{}) {
	smp.mux.Lock()
	defer smp.mux.Unlock()
	smp.mp[key] = value
}

func (smp *SafeMap) Len() int {
	return len(smp.mp)
}

// Range 遍历map ,如果 返回false 则终止遍历
func (smp *SafeMap) Range(fn func(key, value interface{}) bool) {
	smp.mux.RLock()
	defer smp.mux.RUnlock()
	for k, v := range smp.mp {
		isStop := fn(k, v)
		if !isStop {
			break
		}
	}
}

func (smp *SafeMap) Del(Key interface{}) {
	smp.mux.Lock()
	defer smp.mux.Unlock()
	delete(smp.mp, Key)
}
