package mutex

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func Test_Count(t *testing.T) {
	wait := sync.WaitGroup{}
	//var mutex sync.Mutex
	var count int
	for i := 0; i < 10; i++ {
		wait.Add(1)
		go func() {
			for i := 0; i < 10000; i++ {
				count++
			}
			wait.Done()
		}()
	}
	wait.Wait()
	println(count)
}

func Test_Count_normal(t *testing.T) {
	wait := sync.WaitGroup{}
	var mutex sync.Mutex
	var count int
	for i := 0; i < 10; i++ {
		wait.Add(1)
		go func() {
			mutex.Lock()
			for i := 0; i < 10000; i++ {
				count++
			}
			mutex.Unlock()
			wait.Done()
		}()
	}
	wait.Wait()
	println(count)
}

const (
	locked = 1 << iota
	wait
	zero = iota
	one
)

func Test_itoa(t *testing.T) {
	// iota在const关键字出现时将被重置为0，const中每新增一行常量声明将使iota计数一次
	println(locked) // 1 1 << 0 = 1 iota:0
	println(wait)   // 2 1 << 1 = 2 iota:1
	println(zero)   // 2 			iota:2
	println(one)    // 3 			iota:3
}

// chapter3
func Test_ReentrantLock(t *testing.T) {

}

// chapter4 - TryLock
const (
	mutexLocked = 1 << iota
	mutexWoken
	mutexStarving
	mutexWaterShift = iota
)

type Mutex struct {
	sync.Mutex
}

func (m *Mutex) TryLock() bool {
	if atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), 0, mutexLocked) {
		return true
	}

	old := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
	if old&(mutexLocked|mutexStarving|mutexWoken) != 0 {
		return false
	}

	new := old | mutexLocked
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), old, new)
}

type i int64

func Test_Hacker(t *testing.T) {
	var mu Mutex
	go func() {
		mu.Lock()
		defer mu.Unlock()
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	}()

	time.Sleep(time.Second)
	ok := mu.TryLock()
	if ok { // 获取成功
		t.Log("got the lock")
		// do something
		mu.Unlock()
		return
	}

	t.Log("can't get the lock")
}

type Unsafe struct {
	Age    int
	Height int
	Name   string
}

func Test_unsafe(t *testing.T) {
	u := &Unsafe{
		Age:    21,
		Height: 172,
	}
	pointer := unsafe.Pointer(u)
	i := (*int32)(pointer) // 根据结构体field的定义，按顺序顺序赋值第一位给i，此处赋值age给i
	t.Log(i)
}

// chapter4 - monitor lock count
type MonitorMutex struct {
	sync.Mutex
}

func (m *MonitorMutex) Count() int32 {
	state := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
	waitCnt := state >> mutexWaterShift // 锁等待的goroutine数量
	waitCnt += state & mutexLocked      // 加上目前锁持有者的数量，0或1
	return waitCnt
}

func (m *MonitorMutex) isLocked() bool {
	state := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
	return state&mutexLocked == mutexLocked
}

func (m MonitorMutex) isWoken() bool {
	state := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
	return state&mutexWoken == mutexWoken
}

func (m MonitorMutex) isStarving() bool {
	state := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
	return state&mutexStarving == mutexStarving
}

func Test_MonitorMutex(t *testing.T) {
	mutex := MonitorMutex{}
	for i := 0; i < 1000; i++ {
		go func() {
			mutex.Lock()
			time.Sleep(time.Second)
			mutex.Unlock()
		}()

	}

	time.Sleep(1 * time.Second)
	t.Logf("watings:%d, islocked: %v, woken:%v, starving:%v", mutex.Count(), mutex.isLocked(), mutex.isWoken(), mutex.isStarving())
}

type SliceQueue struct {
	data []interface{}
	sync.Mutex
}

func NewSliceQueue(cap int) SliceQueue {
	return SliceQueue{
		data:  make([]interface{}, 0, cap),
		Mutex: sync.Mutex{},
	}
}

func (s *SliceQueue) Enqueue(data interface{}) {
	s.Lock()
	defer s.Unlock()
	s.data = append(s.data, data)
}

func (s *SliceQueue) Dequeue() interface{} {
	s.Lock()
	defer s.Unlock()
	if len(s.data) == 0 {
		return nil
	}
	data := s.data[0]
	s.data = s.data[1:]
	return data
}

func Test_Write_SliceQueue(t *testing.T) {
	queue := NewSliceQueue(1000)
	wait := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wait.Add(1)
		go func() {
			queue.Enqueue(i)
			wait.Done()
		}()
	}

	wait.Wait()
	t.Logf("queue len: %d", len(queue.data))
}

func Test_Read_SliceQueue(t *testing.T) {
	queue := NewSliceQueue(1000)

	for i := 0; i < 1000; i++ {
		queue.Enqueue(i)
	}

	wait := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wait.Add(1)
		go func() {
			queue.Dequeue()
			wait.Done()
		}()
	}

	wait.Wait()
	t.Logf("queue len: %d", len(queue.data))
}
