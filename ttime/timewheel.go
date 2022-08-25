package ttime

import (
	"container/list"
	"fmt"
	"runtime"
	"time"
)

// JobFunc 延时任务回调函数
type JobFunc func(kv ...interface{})

type Option func(tw *TimeWheel)

// TimeWheel 时间轮
type TimeWheel struct {
	interval time.Duration // 指针每隔多久往前移动一格
	ticker   *time.Ticker
	slots    []*list.List // 时间轮槽
	// key: 定时器唯一标识 value: 定时器所在的槽, 主要用于删除定时器, 不会出现并发读写，不加锁直接访问
	timerKeys         map[interface{}]int
	currentPos        int              // 当前指针指向哪一个槽
	slotNum           int              // 槽数量
	addTaskChannel    chan Task        // 新增任务channel
	removeTaskChannel chan interface{} // 删除任务channel
	stopChannel       chan bool        // 停止定时器channel
}

// Task 延时任务
type Task struct {
	delay  time.Duration // 延迟时间
	circle int           // 时间轮需要转动几圈
	key    interface{}   // 定时器唯一标识, 用于删除定时器
	job    JobFunc       // 回调函数
	args   []interface{} // 回调函数参数
}

// NewTimeWheel New 创建时间轮
// 初始化时间轮
// interval 第一个参数为tick刻度, 即时间轮多久转动一次
// slotNum 第二个参数为时间轮槽slot数量
// interval 精确到秒 , 精度越高越耗性能 ，建议到秒级别
// 使用步骤 1、newtimewheel  2、start 3、addtimer 4 stop
func NewTimeWheel(opts ...Option) *TimeWheel {
	tw := &TimeWheel{
		timerKeys:         make(map[interface{}]int, 10),
		currentPos:        0,
		addTaskChannel:    make(chan Task, 10),
		removeTaskChannel: make(chan interface{}, 10),
		stopChannel:       make(chan bool, 1),
	}
	// 设置其他参数
	for i := 0; i < len(opts); i++ {
		opts[i](tw)
	}
	// 默认
	if len(opts) == 0 {
		tw.interval = time.Second
		tw.slotNum = 3600
		tw.slots = make([]*list.List, tw.slotNum)
	}
	if tw.slotNum <= 10 {
		tw.slotNum = 3600
	}
	if tw.interval <= 0 {
		tw.interval = time.Second
	}
	// 初始化槽，每个槽指向一个双向链表
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}

	return tw
}

// WithSlotNum 设置槽的数量
func WithSlotNum(slotNum int) Option {
	return func(tw *TimeWheel) {
		if slotNum < 10 {
			tw.slotNum = 3600
		} else {
			tw.slotNum = slotNum
		}
		tw.slots = make([]*list.List, tw.slotNum)
	}
}

func WithInterval(interval time.Duration) Option {
	return func(tw *TimeWheel) {
		if interval < time.Second {
			tw.interval = time.Second
		} else {
			tw.interval = interval
		}
	}
}

// Start 启动时间轮
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}

// Stop 停止时间轮
func (tw *TimeWheel) Stop() {
	tw.stopChannel <- true
}

// AddTimer 添加定时器 key为定时器唯一标识
func (tw *TimeWheel) AddTimer(delay time.Duration, key interface{}, callBack JobFunc, args ...interface{}) {
	if delay < 0 {
		return
	}
	tw.addTaskChannel <- Task{delay: delay, key: key, job: callBack, args: args}
}

// RemoveTimer 删除定时器 key为添加定时器时传递的定时器唯一标识
func (tw *TimeWheel) RemoveTimer(key interface{}) {
	if key == nil {
		return
	}
	tw.removeTaskChannel <- key
}

func (tw *TimeWheel) start() {
	defer func() {
		fmt.Println("stop wheel ", time.Now().String())
	}()
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(&task)
		case key := <-tw.removeTaskChannel:
			tw.removeTask(key)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
		runtime.Gosched()
	}
}

func (tw *TimeWheel) tickHandler() {
	l := tw.slots[tw.currentPos]
	tw.scanAndRunTask(l)
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

// 扫描链表中过期定时器, 并执行回调函数
func (tw *TimeWheel) scanAndRunTask(l *list.List) {
	if l.Len() <= 0 {
		return
	}
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}
		go task.job(task.args...)
		next := e.Next()
		l.Remove(e)
		if task.key != nil {
			delete(tw.timerKeys, task.key)
		}
		e = next
	}
}

// 新增任务到链表中
func (tw *TimeWheel) addTask(task *Task) {
	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle
	tw.slots[pos].PushBack(task)
	if task.key != nil {
		tw.timerKeys[task.key] = pos
	}
}

// 获取定时器在槽中的位置, 时间轮需要转动的圈数
func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	// 延迟多少秒
	delaySeconds := int(d.Milliseconds())
	// 频率
	intervalSeconds := int(tw.interval.Milliseconds())
	// 圈数 = 延迟 / 频率 / 槽的节点数
	circle = delaySeconds / intervalSeconds / tw.slotNum
	// 刻度 ，最终放到哪个 链表上
	pos = (tw.currentPos + delaySeconds/intervalSeconds) % tw.slotNum
	return
}

// 从链表中删除任务
func (tw *TimeWheel) removeTask(key interface{}) {
	// 获取定时器所在的槽
	position, ok := tw.timerKeys[key]
	if !ok {
		return
	}
	// 获取槽指向的链表
	l := tw.slots[position]
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.key == key {
			delete(tw.timerKeys, task.key)
			l.Remove(e)
		}
		e = e.Next()
	}
}
