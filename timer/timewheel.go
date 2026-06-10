package timer

import (
	"container/list"
	"context"
	"errors"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// TimeWheel 核心结构体
type TimeWheel struct {
	interval   time.Duration
	slots      []*list.List
	ticker     *time.Ticker
	currentPos int
	slotNums   int

	addTaskChannel    chan *Task
	removeTaskChannel chan *removeReq

	workerChan chan Job
	workerNum  int

	cancel    context.CancelFunc
	autoId    uint64
	startMu   sync.Mutex
	isRunning bool

	taskRecords *sync.Map
}

type removeReq struct {
	timerId uint64
	done    chan struct{}
}

// Job 需要执行的 Job 函数
type Job func()

// Task 时间轮盘上需要执行的任务
type Task struct {
	timerId     uint64
	interval    time.Duration
	createdTime time.Time
	pos         int
	circle      int
	job         Job
	times       int
	cancelled   int32
	done        chan struct{}
}

var tw *TimeWheel
var once sync.Once

// NewTimeWheel 用来实现 TimeWheel 的单例模式
func NewTimeWheel(interval time.Duration, slotNums, workerNum int) (*TimeWheel, error) {
	if interval <= 0 || slotNums <= 0 || workerNum <= 0 {
		return nil, errors.New("param err")
	}
	once.Do(func() {
		tw = New(interval, slotNums, workerNum)
	})
	return tw, nil
}

// New 初始化一个 TimeWheel 对象
func New(interval time.Duration, slotNums, workerNum int) *TimeWheel {
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNums),
		slotNums:          slotNums,
		addTaskChannel:    make(chan *Task),
		removeTaskChannel: make(chan *removeReq),
		workerChan:        make(chan Job, workerNum*64),
		workerNum:         workerNum,
		taskRecords:       &sync.Map{},
	}
	tw.initSlots()
	return tw
}

// Start 启动时间轮
func (tw *TimeWheel) Start() {
	tw.startMu.Lock()
	defer tw.startMu.Unlock()
	if tw.isRunning {
		return
	}

	tw.ticker = time.NewTicker(tw.interval)
	ctx, cancel := context.WithCancel(context.Background())
	tw.cancel = cancel
	go tw.start(ctx)
	for i := 0; i < tw.workerNum; i++ {
		go tw.workerRoutine(ctx)
	}
	tw.isRunning = true
}

// Stop 关闭时间轮
func (tw *TimeWheel) Stop() {
	tw.startMu.Lock()
	defer tw.startMu.Unlock()
	if !tw.isRunning || tw.cancel == nil {
		return
	}
	tw.cancel()
	tw.isRunning = false
}

// IsRunning 检查时间轮是否在运行
func (tw *TimeWheel) IsRunning() bool {
	tw.startMu.Lock()
	defer tw.startMu.Unlock()
	return tw.isRunning
}

// AddTimer 向时间轮盘添加任务，times 为 -1 表示永久循环
func (tw *TimeWheel) AddTimer(interval time.Duration, times int, job Job) uint64 {
	if interval <= 0 || times == 0 || job == nil || !tw.IsRunning() {
		return 0
	}

	task := &Task{
		timerId:     atomic.AddUint64(&tw.autoId, 1),
		interval:    interval,
		createdTime: time.Now(),
		job:         job,
		times:       times,
		done:        make(chan struct{}),
	}
	tw.addTaskChannel <- task
	<-task.done
	return task.timerId
}

// RemoveTimer 从时间轮盘删除任务
func (tw *TimeWheel) RemoveTimer(key uint64) {
	if key == 0 || !tw.IsRunning() {
		return
	}

	val, ok := tw.taskRecords.Load(key)
	if !ok || val == nil {
		return
	}

	task := val.(*list.Element).Value.(*Task)
	atomic.StoreInt32(&task.cancelled, 1)

	req := &removeReq{timerId: key, done: make(chan struct{})}
	tw.removeTaskChannel <- req
	<-req.done
}

func (tw *TimeWheel) initSlots() {
	for i := 0; i < tw.slotNums; i++ {
		tw.slots[i] = list.New()
	}
}

func (tw *TimeWheel) start(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("timeWheel start recovered err: %v\n", r)
			log.Printf("timeWheel start recovered stack: %s\n", debug.Stack())
			if ctx.Err() == nil {
				go tw.start(ctx)
			}
		}
	}()

	for {
		select {
		case <-tw.ticker.C:
			tw.checkAndRunTask()
		case task := <-tw.addTaskChannel:
			tw.addTask(task, true)
			close(task.done)
		case req := <-tw.removeTaskChannel:
			tw.removeTaskByID(req.timerId)
			close(req.done)
		case <-ctx.Done():
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimeWheel) workerRoutine(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("timeWheel workerRoutine recovered err: %v\n", r)
			log.Printf("timeWheel workerRoutine recovered stack: %s\n", debug.Stack())
			if ctx.Err() == nil {
				go tw.workerRoutine(ctx)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case execFunc := <-tw.workerChan:
			if execFunc != nil {
				execFunc()
			}
		}
	}
}

func (tw *TimeWheel) checkAndRunTask() {
	currentList := tw.slots[tw.currentPos]
	if currentList != nil {
		for item := currentList.Front(); item != nil; {
			task := item.Value.(*Task)
			if task.circle > 0 {
				task.circle--
				item = item.Next()
				continue
			}

			next := item.Next()
			tw.taskRecords.Delete(task.timerId)
			currentList.Remove(item)

			if task.job != nil && task.times != 0 && atomic.LoadInt32(&task.cancelled) == 0 {
				tw.dispatchJob(task.job)
			}

			if atomic.LoadInt32(&task.cancelled) == 0 {
				if task.times == -1 {
					tw.addTask(task, true)
				} else {
					task.times--
					if task.times > 0 {
						tw.addTask(task, true)
					}
				}
			}

			item = next
		}
	}

	if tw.currentPos == tw.slotNums-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

func (tw *TimeWheel) dispatchJob(job Job) {
	select {
	case tw.workerChan <- job:
	default:
		go job()
	}
}

func (tw *TimeWheel) addTask(task *Task, byInterval bool) {
	var pos, circle int
	if byInterval {
		pos, circle = tw.getPosAndCircleByInterval(task.interval)
	} else {
		pos, circle = tw.getPosAndCircleByCreatedTime(task.createdTime, task.interval)
	}

	task.circle = circle
	task.pos = pos
	element := tw.slots[pos].PushBack(task)
	tw.taskRecords.Store(task.timerId, element)
}

func (tw *TimeWheel) removeTaskByID(timerId uint64) {
	val, ok := tw.taskRecords.LoadAndDelete(timerId)
	if !ok || val == nil {
		return
	}

	el := val.(*list.Element)
	task := el.Value.(*Task)
	atomic.StoreInt32(&task.cancelled, 1)
	tw.slots[task.pos].Remove(el)
}

func delayTicks(delay, interval time.Duration) int {
	if delay <= 0 {
		return 1
	}
	d := int64(delay)
	iv := int64(interval)
	return int((d + iv - 1) / iv)
}

func (tw *TimeWheel) getPosAndCircleByInterval(delay time.Duration) (int, int) {
	ticks := delayTicks(delay, tw.interval)
	circle := ticks / tw.slotNums
	pos := (tw.currentPos + ticks) % tw.slotNums
	return pos, circle
}

func (tw *TimeWheel) getPosAndCircleByCreatedTime(createdTime time.Time, delay time.Duration) (int, int) {
	remaining := delay - time.Since(createdTime)
	if remaining <= 0 {
		remaining = tw.interval
	}
	return tw.getPosAndCircleByInterval(remaining)
}
