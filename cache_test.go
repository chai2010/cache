// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

// See https://github.com/google/leveldb/blob/master/util/cache_test.cc

import (
	"strconv"
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
	defer h.Close()
}

func (p *TCache) Lookup(key int) int {
	if v, h, ok := p.LRUCache.Lookup(strconv.Itoa(key)); ok {
		h.Close()
		return v.(int)
	}
	return -1
}

func (p *TCache) Erase(key int) {
	p.LRUCache.Erase(strconv.Itoa(key))
}

func TestLRUCache_hitAndMiss(t *testing.T) {
	c := tNewTCache(tCacheSize)
	defer c.Close()

	tAssertEQ(t, -1, c.Lookup(100))

	c.Insert(100, 101)
	tAssertEQ(t, 101, c.Lookup(100))
	tAssertEQ(t, -1, c.Lookup(200))
	tAssertEQ(t, -1, c.Lookup(200))

	c.Insert(200, 201)
	tAssertEQ(t, 101, c.Lookup(100))
	tAssertEQ(t, 201, c.Lookup(200))
	tAssertEQ(t, -1, c.Lookup(300))

	c.Insert(100, 102)
	tAssertEQ(t, 102, c.Lookup(100))
	tAssertEQ(t, 201, c.Lookup(200))
	tAssertEQ(t, -1, c.Lookup(300))

	tAssertEQ(t, 1, len(c.deleted_keys_))
	tAssertEQ(t, 100, c.deleted_keys_[0])
	tAssertEQ(t, 101, c.deleted_values_[0])
}

func TestLRUCache_erase(t *testing.T) {
	c := tNewTCache(tCacheSize)
	defer c.Close()

	c.Erase(200)
	tAssertEQ(t, 0, len(c.deleted_keys_))

	c.Insert(100, 101)
	c.Insert(200, 201)
	c.Erase(100)
	tAssertEQ(t, -1, c.Lookup(100))
	tAssertEQ(t, 201, c.Lookup(200))
	tAssertEQ(t, 1, len(c.deleted_keys_))
	tAssertEQ(t, 100, c.deleted_keys_[0])
	tAssertEQ(t, 101, c.deleted_values_[0])

	c.Erase(100)
	tAssertEQ(t, -1, c.Lookup(100))
	tAssertEQ(t, 201, c.Lookup(200))
	tAssertEQ(t, 1, len(c.deleted_keys_))
}

func TestLRUCache_entriesArePinned(t *testing.T) {
	c := tNewTCache(tCacheSize)
	defer c.Close()

	c.Insert(100, 101)
	v, h1, ok := c.LRUCache.Lookup(strconv.Itoa(100))
	tAssertTrue(t, ok)
	tAssertEQ(t, 101, v.(int))

	c.Insert(100, 102)
	v, h2, ok := c.LRUCache.Lookup(strconv.Itoa(100))
	tAssertTrue(t, ok)
	tAssertEQ(t, 102, v.(int))
	tAssertEQ(t, 0, len(c.deleted_keys_))

	h1.Close()
	tAssertEQ(t, 1, len(c.deleted_keys_))
	tAssertEQ(t, 100, c.deleted_keys_[0])
	tAssertEQ(t, 101, c.deleted_values_[0])

	c.Erase(100)
	tAssertEQ(t, -1, c.Lookup(100))
	tAssertEQ(t, 1, len(c.deleted_keys_))

	h2.Close()
	tAssertEQ(t, 2, len(c.deleted_keys_))
	tAssertEQ(t, 100, c.deleted_keys_[1])
	tAssertEQ(t, 102, c.deleted_values_[1])
}

func TestLRUCache_evictionPolicy(t *testing.T) {
	c := tNewTCache(tCacheSize)
	defer c.Close()

	c.Insert(100, 101)
	c.Insert(200, 201)

	// Frequently used entry must be kept around
	for i := 0; i < tCacheSize+100; i++ {
		c.Insert(1000+i, 2000+i)
		tAssertEQ(t, 2000+i, c.Lookup(1000+i))
		tAssertEQ(t, 101, c.Lookup(100))
	}
	tAssertEQ(t, 101, c.Lookup(100))
	tAssertEQ(t, -1, c.Lookup(200))
}

func TestLRUCache_heavyEntries(t *testing.T) {
	c := tNewTCache(tCacheSize)
	defer c.Close()

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
		c.Insert(index, 1000+index, weight)
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
		var r = c.Lookup(i)
		if r >= 0 {
			cached_weight += weight
			tAssertEQ(t, 1000+i, r)
		}
	}
	tAssertLE(t, cached_weight, tCacheSize+tCacheSize/10)
}

func TestLRUCache_NewId(t *testing.T) {
	c := tNewTCache(tCacheSize)
	defer c.Close()

	a := c.NewId()
	b := c.NewId()
	tAssertNE(t, a, b)
}

func tAtoi(s string, defaultV int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultV
}
