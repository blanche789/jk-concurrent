package mutex

import (
	"sync"
	"testing"
)

func Test_Count(t *testing.T) {
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
