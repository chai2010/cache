// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const (
	tCacheSize = 1000
)

type TCache struct {
	*LRUCache
	deleted_keys_   []int
	deleted_values_ []int
}

func tNewTCache(capacity int64) *TCache {
	return &TCache{
		LRUCache: NewLRUCache(capacity),
	}
}

func (p *TCache) onDeleter(key string, value interface{}) {
	p.deleted_keys_ = append(p.deleted_keys_, tAtoi(key, -1))
	p.deleted_values_ = append(p.deleted_values_, value.(int))
}

func (p *TCache) Insert(key, value int, size ...int) {
	if len(size) == 0 {
		size = []int{1}
	}
	h := p.LRUCache.Insert(strconv.Itoa(key), value, size[0], p.onDeleter)
	defer h.Release()
}

func (p *TCache) Lookup(key int) int {
	h, ok := p.LRUCache.Lookup(strconv.Itoa(key))
	if !ok {
		return -1
	}
	defer h.Release()
	return h.Value().(int)
}

func (p *TCache) Erase(key int) {
	p.LRUCache.Erase(strconv.Itoa(key))
}

func TestLRUCache_hitAndMiss(t *testing.T) {
	cache := tNewTCache(tCacheSize)
	defer cache.Close()

	tAssertEQ(t, -1, cache.Lookup(100))

	cache.Insert(100, 101)
	tAssertEQ(t, 101, cache.Lookup(100))
	tAssertEQ(t, -1, cache.Lookup(200))
	tAssertEQ(t, -1, cache.Lookup(200))

	cache.Insert(200, 201)
	tAssertEQ(t, 101, cache.Lookup(100))
	tAssertEQ(t, 201, cache.Lookup(200))
	tAssertEQ(t, -1, cache.Lookup(300))

	cache.Insert(100, 102)
	tAssertEQ(t, 102, cache.Lookup(100))
	tAssertEQ(t, 201, cache.Lookup(200))
	tAssertEQ(t, -1, cache.Lookup(300))

	tAssertEQ(t, 1, len(cache.deleted_keys_))
	tAssertEQ(t, 100, cache.deleted_keys_[0])
	tAssertEQ(t, 101, cache.deleted_values_[0])
}

func TestLRUCache_erase(t *testing.T) {
	cache := tNewTCache(tCacheSize)
	defer cache.Close()

	cache.Erase(200)
	tAssertEQ(t, 0, len(cache.deleted_keys_))

	cache.Insert(100, 101)
	cache.Insert(200, 201)
	cache.Erase(100)
	tAssertEQ(t, -1, cache.Lookup(100))
	tAssertEQ(t, 201, cache.Lookup(200))
	tAssertEQ(t, 1, len(cache.deleted_keys_))
	tAssertEQ(t, 100, cache.deleted_keys_[0])
	tAssertEQ(t, 101, cache.deleted_values_[0])

	cache.Erase(100)
	tAssertEQ(t, -1, cache.Lookup(100))
	tAssertEQ(t, 201, cache.Lookup(200))
	tAssertEQ(t, 1, len(cache.deleted_keys_))
}

func TestLRUCache_entriesArePinned(t *testing.T) {
	cache := tNewTCache(tCacheSize)
	defer cache.Close()

	cache.Insert(100, 101)
	h1, _ := cache.LRUCache.Lookup(strconv.Itoa(100))
	tAssertEQ(t, 101, h1.Value().(int))

	cache.Insert(100, 102)
	h2, _ := cache.LRUCache.Lookup(strconv.Itoa(100))
	tAssertEQ(t, 102, h2.Value().(int))
	tAssertEQ(t, 0, len(cache.deleted_keys_))

	h1.Release()
	tAssertEQ(t, 1, len(cache.deleted_keys_))
	tAssertEQ(t, 100, cache.deleted_keys_[0])
	tAssertEQ(t, 101, cache.deleted_values_[0])

	cache.Erase(100)
	tAssertEQ(t, -1, cache.Lookup(100))
	tAssertEQ(t, 1, len(cache.deleted_keys_))

	h2.Release()
	tAssertEQ(t, 2, len(cache.deleted_keys_))
	tAssertEQ(t, 100, cache.deleted_keys_[1])
	tAssertEQ(t, 102, cache.deleted_values_[1])
}

func TestLRUCache_evictionPolicy(t *testing.T) {
	cache := tNewTCache(tCacheSize)
	defer cache.Close()

	cache.Insert(100, 101)
	cache.Insert(200, 201)

	// Frequently used entry must be kept around
	for i := 0; i < tCacheSize+100; i++ {
		cache.Insert(1000+i, 2000+i)
		tAssertEQ(t, 2000+i, cache.Lookup(1000+i))
		tAssertEQ(t, 101, cache.Lookup(100))
	}
	tAssertEQ(t, 101, cache.Lookup(100))
	tAssertEQ(t, -1, cache.Lookup(200))
}

func TestLRUCache_heavyEntries(t *testing.T) {
	cache := tNewTCache(tCacheSize)
	defer cache.Close()

	// Add a bunch of light and heavy entries and then count the combined
	// size of items still in the cache, which must be approximately the
	// same as the total capacity.
	const kLight = 1
	const kHeavy = 10
	var added = 0
	var index = 0
	for added < 2*tCacheSize {
		var weight int
		if (index & 1) != 0 {
			weight = kLight
		} else {
			weight = kHeavy
		}
		cache.Insert(index, 1000+index, weight)
		added += weight
		index++
	}

	var cached_weight = 0
	for i := 0; i < index; i++ {
		var weight int
		if (i & 1) != 0 {
			weight = kLight
		} else {
			weight = kHeavy
		}
		var r = cache.Lookup(i)
		if r >= 0 {
			cached_weight += weight
			tAssertEQ(t, 1000+i, r)
		}
	}
	tAssertLE(t, cached_weight, tCacheSize+tCacheSize/10)
}

func TestLRUCache_NewId(t *testing.T) {
	cache := tNewTCache(tCacheSize)
	defer cache.Close()

	a := cache.NewId()
	b := cache.NewId()
	tAssertNE(t, a, b)
}

func tAssert(t *testing.T, v bool, a ...interface{}) {
	if !v {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprint(a...); msg != "" {
			t.Fatalf("%s:%d: tAssert failed, %s!", file, line, msg)
		} else {
			t.Fatalf("%s:%d: tAssert failed!", file, line)
		}
	}
}

func tAssertf(t *testing.T, v bool, format string, a ...interface{}) {
	if !v {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprintf(format, a...); msg != "" {
			t.Fatalf("%s:%d: tAssert failed, %s!", file, line, msg)
		} else {
			t.Fatalf("%s:%d: tAssert failed!", file, line)
		}
	}
}

func tAssertEQ(t *testing.T, a, b interface{}, s ...interface{}) {
	if !reflect.DeepEqual(a, b) {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprint(s...); msg != "" {
			t.Fatalf("%s:%d: tAssertEQ failed, %v != %v, %s!", file, line, a, b, msg)
		} else {
			t.Fatalf("%s:%d: tAssertEQ failed, %v != %v!", file, line, a, b)
		}
	}
}

func tAssertNE(t *testing.T, a, b interface{}, s ...interface{}) {
	if reflect.DeepEqual(a, b) {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprint(s...); msg != "" {
			t.Fatalf("%s:%d: tAssertNE failed, %v == %v, %s!", file, line, a, b, msg)
		} else {
			t.Fatalf("%s:%d: tAssertNE failed, %v == %v!", file, line, a, b)
		}
	}
}

func tAssertLE(t *testing.T, a, b int, s ...interface{}) {
	if !(a <= b) {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprint(s...); msg != "" {
			t.Fatalf("%s:%d: tAssertNE failed, %v > %v, %s!", file, line, a, b, msg)
		} else {
			t.Fatalf("%s:%d: tAssertNE failed, %v > %v!", file, line, a, b)
		}
	}
}

func tCallerFileLine(skip int) (file string, line int) {
	_, file, line, ok := runtime.Caller(skip + 1)
	if ok {
		// Truncate file name at last file name separator.
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
	} else {
		file = "???"
		line = 1
	}
	return
}

func tAtoi(s string, defaultV int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultV
}
