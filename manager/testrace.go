package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

var mut sync.Mutex

func main() {

	threadCount := 10

	wgr := new(sync.WaitGroup)
	for i := threadCount; i > 0; i-- {
		wgr.Add(1)
		go bl(i, wgr)
	}
	wgr.Wait()
        time.Sleep(1)
}

func bl(threadCount int, wgr *sync.WaitGroup) {
	defer wgr.Done()
	mut.Lock()
	a := threadCount
	a = a + 1
	mut.Unlock()
	//time.Sleep(time.Duration(a) * time.Second)
	fmt.Println("i am thread " + strconv.Itoa(threadCount) + " my value is " + strconv.Itoa(a))
}
