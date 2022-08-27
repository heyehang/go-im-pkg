package tcontainer

import (
	"fmt"
	"sync"

	"github.com/emirpasic/gods/trees/redblacktree"
)

type TreeMap struct {
	sync.RWMutex
	tree *redblacktree.Tree
}

func NewTreeMap() *TreeMap {
	treeMp := new(TreeMap)
	treeMp.RWMutex = sync.RWMutex{}
	treeMp.tree = redblacktree.NewWithStringComparator()
	return treeMp
}

func (t *TreeMap) Get(key interface{}) (value interface{}, found bool) {
	t.RLock()
	defer t.RUnlock()
	return t.tree.Get(getKey(key))
}

func getKey(v interface{}) string {
	str, ok := v.(string)
	if ok {
		return str
	}
	return fmt.Sprintf("%+v", v)
}

func (t *TreeMap) Set(key, value interface{}) {
	t.Lock()
	defer t.Unlock()
	t.tree.Put(getKey(key), value)
}

func (t *TreeMap) Del(key interface{}) {
	t.Lock()
	defer t.Unlock()
	_, ok := t.tree.Get(getKey(key))
	if !ok {
		return
	}
	t.tree.Remove(getKey(key))
}

func (t *TreeMap) Len() int {
	t.RLock()
	defer t.RUnlock()
	return t.tree.Size()
}

func (t *TreeMap) Exists(key interface{}) bool {
	t.RLock()
	defer t.RUnlock()
	_, ok := t.tree.Get(getKey(key))
	return ok
}

// isStop == true 则会打断遍历
func (t *TreeMap) Range(fun func(k, v interface{}) (isStop bool)) {
	it := t.tree.Iterator()
	for i := 0; it.Next(); i++ {
		if !fun(it.Key(), it.Value()) {
			break
		}
	}
}

func (t *TreeMap) Clear() {
	t.Lock()
	defer t.Unlock()
	t.tree.Clear()
}

func (t *TreeMap) Keys() []interface{} {
	t.RLock()
	defer t.RUnlock()
	return t.tree.Keys()
}

func (t *TreeMap) Values() []interface{} {
	t.RLock()
	defer t.RUnlock()
	return t.tree.Values()
}

func (t *TreeMap) ToJSON() ([]byte, error) {
	t.Lock()
	defer t.Unlock()
	return t.tree.ToJSON()
}
