// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

var (
	_ Handle = (*LRUHandle)(nil)
)

// LRUCache is a typical LRU cache implementation.  If the cache
// reaches the capacity, the least recently used item is deleted from
// the cache. Note the capacity is not the number of items, but the
// total sum of the Size() of each item.
type LRUCache struct {
	mu sync.Mutex

	// list & table of *LRUHandle objects
	list  *list.List
	table map[string]*list.Element

	// Our current size. Obviously a gross simplification and
	// low-grade approximation.
	size int64

	// How much we are limiting the cache to.
	capacity int64

	// for next id
	last_id uint64
}

type LRUHandle struct {
	key           string
	value         interface{}
	size          int64
	deleter       func(key string, value interface{})
	time_accessed time.Time
}

func (p *LRUHandle) Value() interface{} {
	return p.value
}

func (p *LRUHandle) Release() {
	// ref--
}

// NewLRUCache creates a new empty cache with the given capacity.
func NewLRUCache(capacity int64) *LRUCache {
	return &LRUCache{
		list:     list.New(),
		table:    make(map[string]*list.Element),
		capacity: capacity,
	}
}

func (p *LRUCache) NewId() uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	v := p.last_id
	p.last_id++
	return v
}

// Get returns a value from the cache, and marks the LRUHandle as most
// recently used.
func (p *LRUCache) Get(key string) (v interface{}, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil {
		return nil, false
	}
	p.moveToFront(element)
	return element.Value.(*LRUHandle).value, true
}

// Set sets a value in the cache.
func (p *LRUCache) Set(key string, value interface{}, size int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if element := p.table[key]; element != nil {
		p.updateInplace(element, value, size)
	} else {
		p.addNew(key, value, size)
	}
}

// Delete removes an LRUHandle from the cache, and returns if the LRUHandle existed.
func (p *LRUCache) Delete(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil {
		return false
	}

	p.list.Remove(element)
	delete(p.table, key)
	p.size -= element.Value.(*LRUHandle).size
	return true
}

// Clear will clear the entire cache.
func (p *LRUCache) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.list.Init()
	p.table = make(map[string]*list.Element)
	p.size = 0
}

// SetCapacity will set the capacity of the cache. If the capacity is
// smaller, and the current cache size exceed that capacity, the cache
// will be shrank.
func (p *LRUCache) SetCapacity(capacity int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.capacity = capacity
	p.checkCapacity()
}

// Stats returns a few stats on the cache.
func (p *LRUCache) Stats() (length, size, capacity int64, oldest time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if lastElem := p.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*LRUHandle).time_accessed
	}
	return int64(p.list.Len()), p.size, p.capacity, oldest
}

// StatsJSON returns stats as a JSON object in a string.
func (p *LRUCache) StatsJSON() string {
	if p == nil {
		return "{}"
	}
	l, s, c, o := p.Stats()
	return fmt.Sprintf("{\"Length\": %v, \"Size\": %v, \"Capacity\": %v, \"OldestAccess\": \"%v\"}", l, s, c, o)
}

// Length returns how many elements are in the cache
func (p *LRUCache) Length() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return int64(p.list.Len())
}

// Size returns the sum of the objects' Size() method.
func (p *LRUCache) Size() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.size
}

// Capacity returns the cache maximum capacity.
func (p *LRUCache) Capacity() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.capacity
}

// Oldest returns the insertion time of the oldest element in the cache,
// or a IsZero() time if cache is empty.
func (p *LRUCache) Oldest() (oldest time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if lastElem := p.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*LRUHandle).time_accessed
	}
	return
}

// Keys returns all the keys for the cache, ordered from most recently
// used to last recently used.
func (p *LRUCache) Keys() []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	keys := make([]string, 0, p.list.Len())
	for e := p.list.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*LRUHandle).key)
	}
	return keys
}

func (p *LRUCache) updateInplace(element *list.Element, value interface{}, size int) {
	valueSize := int64(size)
	sizeDiff := valueSize - element.Value.(*LRUHandle).size
	element.Value.(*LRUHandle).value = value
	element.Value.(*LRUHandle).size = valueSize
	p.size += sizeDiff
	p.moveToFront(element)
	p.checkCapacity()
}

func (p *LRUCache) moveToFront(element *list.Element) {
	p.list.MoveToFront(element)
	element.Value.(*LRUHandle).time_accessed = time.Now()
}

func (p *LRUCache) addNew(key string, value interface{}, size int) {
	newEntry := &LRUHandle{key, value, int64(size), func(key string, value interface{}) {}, time.Now()}
	element := p.list.PushFront(newEntry)
	p.table[key] = element
	p.size += newEntry.size
	p.checkCapacity()
}

func (p *LRUCache) checkCapacity() {
	// Partially duplicated from Delete
	for p.size > p.capacity {
		delElem := p.list.Back()
		delValue := delElem.Value.(*LRUHandle)
		p.list.Remove(delElem)
		delete(p.table, delValue.key)
		p.size -= delValue.size
	}
}
