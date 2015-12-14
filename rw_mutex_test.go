// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testRwMutexName = "rwm-test"

func TestLockRwMutex(t *testing.T) {
	if !assert.NoError(t, DestroyRwMutex(testRwMutexName)) {
		return
	}
	mut, err := NewRwMutex(testRwMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) || !assert.NotNil(t, mut) {
		return
	}
	defer mut.Destroy()
	var wg sync.WaitGroup
	sharedValue := 0
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			mut.Lock()
			for i := 0; i < 1000; i++ {
				sharedValue++
			}
			mut.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, 30000, sharedValue)
}

func TestRwMutexOpenMode(t *testing.T) {
	if !assert.NoError(t, DestroyRwMutex(testRwMutexName)) {
		return
	}
	mut, err := NewRwMutex(testRwMutexName, O_READWRITE, 0666)
	assert.Error(t, err)
	mut, err = NewRwMutex(testRwMutexName, O_CREATE_ONLY|O_READ_ONLY, 0666)
	assert.Error(t, err)
	mut, err = NewRwMutex(testRwMutexName, O_OPEN_OR_CREATE|O_WRITE_ONLY, 0666)
	assert.Error(t, err)
	mut, err = NewRwMutex(testRwMutexName, O_OPEN_ONLY|O_WRITE_ONLY, 0666)
	assert.Error(t, err)
	mut, err = NewRwMutex(testRwMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mut.Destroy()
	mut, err = NewRwMutex(testRwMutexName, O_OPEN_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
}

func TestRwMutexOpenMode2(t *testing.T) {
	if !assert.NoError(t, DestroyRwMutex(testRwMutexName)) {
		return
	}
	mut, err := NewRwMutex(testRwMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mut.Destroy()
	mut, err = NewRwMutex(testRwMutexName, O_OPEN_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	mut, err = NewRwMutex(testRwMutexName, O_CREATE_ONLY, 0666)
	if !assert.Error(t, err) {
		return
	}
}

func TestRwMutexOpenMode3(t *testing.T) {
	if !assert.NoError(t, DestroyRwMutex(testRwMutexName)) {
		return
	}
	_, err := NewRwMutex(testRwMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
}

func printer() chan int {
	ch := make(chan int, 100)
	var cur int32
	go func() {
		for value := range ch {
			atomic.StoreInt32(&cur, int32(value))
			/*if value > 0 {
				fmt.Printf("%d got the lock\n", value)
			} else {
				fmt.Printf("%d released the lock\n", -value)
			}*/
		}
	}()
	go func() {
		for {
			<-time.After(time.Millisecond * 250)
			value := atomic.LoadInt32(&cur)
			if value > 0 {
				fmt.Printf("%d got the lock\n", value)
			} else {
				fmt.Printf("%d released the lock\n", -value)
			}
		}
	}()
	return ch
}

func NoTestRwMutexMemory(t *testing.T) {
	if !assert.NoError(t, DestroyRwMutex(testRwMutexName)) {
		return
	}
	mut, err := NewRwMutex(testRwMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mut.Destroy()
	region, err := createMemoryRegionSimple(O_OPEN_OR_CREATE|O_READWRITE, SHM_READWRITE, 512, 0)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		region.Close()
		DestroyMemoryObject(defaultObjectName)
	}()
	data := region.Data()
	for i, _ := range data { // fill the data with correct values
		data[i] = byte(i)
	}
	args := argsForSyncTestCommand(testRwMutexName, "rwm", 1, defaultObjectName, 100, data, "/home/avd/sync.log")
	var wg sync.WaitGroup
	var flag int32 = 1
	const jobs = 4
	wg.Add(jobs)
	ch := printer()
	for i := 0; i < jobs; i++ {
		go func(i int) {
			for j := 0; atomic.LoadInt32(&flag) != 0 && j < 162100; j++ {
				mut.Lock()
				ch <- i
				// corrupt the data and then restore it.
				// as the entire operation is under mutex protection,
				// no one should see these changes.
				for i, _ := range data {
					data[i] = byte(0)
				}
				for i, _ := range data {
					data[i] = byte(i)
				}
				mut.Unlock()
				ch <- -i
			}
			wg.Done()
		}(i)
	}
	result := runTestApp(args, nil)
	atomic.StoreInt32(&flag, 0)
	wg.Wait()
	if !assert.NoError(t, result.err) {
		t.Logf("test app error. the output is: %s", result.output)
	}
}
