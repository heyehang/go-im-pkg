package util

import (
	"hash/crc32"
	"math"
	"sort"
	"strconv"
	"sync"
)

const (
	// 默认虚拟节点
	DefaultVirualSpots = 100
)

type node struct {
	nodeKey   string
	spotValue uint32
}

type nodesArray []node

func (p nodesArray) Len() int           { return len(p) }
func (p nodesArray) Less(i, j int) bool { return p[i].spotValue < p[j].spotValue }
func (p nodesArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p nodesArray) Sort()              { sort.Sort(p) }

//HashRing 结构体
type HashRing struct {
	// 虚拟槽节点数
	virualSpots int
	// 节点数组
	nodes nodesArray
	// 权重map
	weights map[string]int
	mu      sync.RWMutex
}

//创建一个hashring
// 建议值  虚拟节点数 =  已知总节点数 * 100 如果知道总节点数为6 个，那么虚拟节点数值 = 6 *100     差值最小，分布最均匀,差值在千分之5以内
func NewHashRing(spots int) *HashRing {
	if spots == 0 {
		spots = DefaultVirualSpots
	}
	h := &HashRing{
		virualSpots: spots,
		weights:     make(map[string]int),
	}
	return h
}

//AddNodes 添加一个节点到换上
func (h *HashRing) AddNodes(nodeWeight map[string]int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	//// 动态调整虚拟节点数，根据添加节点进行平均
	h.virualSpots = h.virualSpots + h.virualSpots*len(nodeWeight)

	for nodeKey, w := range nodeWeight {
		h.weights[nodeKey] = w
	}
	h.generate()
}

//AddNode 添加一个节点 ， weight 权重
func (h *HashRing) AddNode(nodeKey string, weight int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// 动态调整虚拟节点数，根据添加节点进行平均
	h.virualSpots = h.virualSpots + h.virualSpots*1

	h.weights[nodeKey] = weight
	h.generate()
}

//RemoveNode 移除一个节点
func (h *HashRing) RemoveNode(nodeKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.weights) > 0 {
		// 动态调整虚拟节点数
		h.virualSpots = h.virualSpots - h.virualSpots/len(h.weights)
	}
	delete(h.weights, nodeKey)
	h.generate()
}

//UpdateNode 更新一个节点的权重
func (h *HashRing) UpdateNode(nodeKey string, weight int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.weights[nodeKey] = weight
	h.generate()
}

func (h *HashRing) generate() {
	var totalW int
	for _, w := range h.weights {
		totalW += w
	}

	// 总虚拟节点 = 环当前设置虚拟节点数 * 节点数
	totalVirtualSpots := h.virualSpots * len(h.weights)
	h.nodes = nodesArray{}

	for nodeKey, w := range h.weights {
		// 计算 单个节点的虚拟节点数  = 当前节点权重值 / 总权重值  * 总虚拟节点数
		spots := int(math.Floor(float64(w) / float64(totalW) * float64(totalVirtualSpots)))
		for i := 1; i <= spots; i++ {
			spValue := crc32.Checksum([]byte(nodeKey+":"+strconv.Itoa(i)), crc32.IEEETable)
			n := node{
				nodeKey:   nodeKey,
				spotValue: spValue,
			}
			// 创建一个节点
			h.nodes = append(h.nodes, n)
		}
	}
	h.nodes.Sort()
}

//GetNode 根据一个key 获取一个节点主机
func (h *HashRing) GetNode(key string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.nodes) == 0 {
		return ""
	}
	v := crc32.Checksum([]byte(key), crc32.IEEETable)
	i := sort.Search(len(h.nodes), func(i int) bool { return h.nodes[i].spotValue >= v })
	if i == len(h.nodes) {
		i = 0
	}
	return h.nodes[i].nodeKey
}
