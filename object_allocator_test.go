// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCheckObjectType(t *testing.T) {
	type validStruct struct {
		a, b int
		u    uintptr
		s    struct {
			arr [3]int
		}
	}
	type invalidStruct1 struct {
		a, b *int
	}
	type invalidStruct2 struct {
		a, b []int
	}
	type invalidStruct3 struct {
		s string
	}
	var i int
	var c complex128
	var arr = [3]int{}
	var arr2 = [3]string{}
	var slsl [][]int
	var m map[int]int

	assert.NoError(t, checkObject(i))
	assert.NoError(t, checkObject(c))
	assert.NoError(t, checkObject(arr))
	assert.NoError(t, checkObject(arr[:]))
	assert.NoError(t, checkObject(validStruct{}))
	assert.Error(t, checkObject(invalidStruct1{}))
	assert.Error(t, checkObject(invalidStruct2{}))
	assert.Error(t, checkObject(invalidStruct3{}))
	assert.Error(t, checkObject(arr2))
	assert.Error(t, checkObject(arr2[:]))
	assert.NoError(t, checkObject(sync.Mutex{}))
	assert.Error(t, checkObject(m))

	assert.Error(t, checkObject(slsl))
}

func TestAllocInt(t *testing.T) {
	var i int = 0x01027FFF
	data := make([]byte, unsafe.Sizeof(i))
	if !assert.NoError(t, alloc(data, i)) {
		return
	}
	ptr := (*int)(unsafe.Pointer(byteSliceToUintPtr(data)))
	assert.Equal(t, i, *ptr)
}

func TestAllocIntArray(t *testing.T) {
	i := [3]int{0x01, 0x7F, 0xFF}
	data := make([]byte, unsafe.Sizeof(i))
	if !assert.NoError(t, alloc(data, i)) {
		return
	}
	ptr := (*[3]int)(unsafe.Pointer(byteSliceToUintPtr(data)))
	assert.Equal(t, i, *ptr)
}

func TestAllocStruct(t *testing.T) {
	type internal struct {
		d complex128
		p uintptr
	}
	type s struct {
		a, b int
		ss   internal
	}
	obj := s{-1, 11, internal{complex(10, 11), uintptr(0)}}
	data := make([]byte, unsafe.Sizeof(obj))
	if !assert.NoError(t, alloc(data, obj)) {
		return
	}
	ptr := (*s)(unsafe.Pointer(byteSliceToUintPtr(data)))
	assert.Equal(t, obj, *ptr)
}

func TestAllocMutex(t *testing.T) {
	var obj sync.Mutex
	data := make([]byte, unsafe.Sizeof(obj))
	if !assert.NoError(t, alloc(data, obj)) {
		return
	}
	ptr := (*sync.Mutex)(unsafe.Pointer(byteSliceToUintPtr(data)))
	assert.Equal(t, obj, *ptr)
}

/* TODO(avd) - fix it
func TestAllocSlice(t *testing.T) {
	obj := make([]uint, 10)
	for i, _ := range obj {
		obj[i] = uint(i)
	}
	data := make([]byte, unsafe.Sizeof(uint(0))*10)
	if !assert.NoError(t, alloc(data, obj)) {
		return
	}
	ptr := (*[]byte)(unsafe.Pointer(byteSliceToUintPtr(data)))
	assert.Equal(t, obj, *ptr)
}*/
