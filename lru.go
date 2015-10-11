// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	_ Cache  = (*LRUCache)(nil)
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

// LRUHandle handle to an entry stored in the LRUCache.
type LRUHandle struct {
	c             *LRUCache
	key           string
	value         interface{}
	size          int64
	deleter       func(key string, value interface{})
	time_created  time.Time
	time_accessed atomic.Value // time.Time
	refs          uint32
}

func (h *LRUHandle) Key() string {
	return h.key
}

func (h *LRUHandle) Value() interface{} {
	return h.value
}

func (h *LRUHandle) Size() int {
	return int(h.size)
}
func (h *LRUHandle) TimeCreated() time.Time {
	return h.time_created
}

func (h *LRUHandle) TimeAccessed() time.Time {
	return h.time_accessed.Load().(time.Time)
}

func (h *LRUHandle) Retain() Handle {
	h.c.mu.Lock()
	defer h.c.mu.Unlock()
	h.c.addref(h)
	return h
}

func (h *LRUHandle) Release() {
	h.c.mu.Lock()
	defer h.c.mu.Unlock()
	h.c.unref(h)
}

// match for io.Closer, same as h.Release.
func (h *LRUHandle) Close() error {
	h.c.mu.Lock()
	defer h.c.mu.Unlock()
	h.c.unref(h)
	return nil
}

// NewLRUCache creates a new empty cache with the given capacity.
func NewLRUCache(capacity int64) *LRUCache {
	assert(capacity > 0)
	return &LRUCache{
		list:     list.New(),
		table:    make(map[string]*list.Element),
		capacity: capacity,
	}
}

func (p *LRUCache) Get(key string) (value interface{}, ok bool) {
	h, ok := p.Lookup(key)
	if !ok {
		return nil, false
	}
	value = h.Value()
	h.Release()
	return
}

func (p *LRUCache) Value(key string, defaultValue ...interface{}) interface{} {
	h, ok := p.Lookup(key)
	if !ok {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		} else {
			return nil
		}
	}
	v := h.Value()
	h.Release()
	return v
}

func (p *LRUCache) Set(key string, value interface{}, size int, deleter ...func(key string, value interface{})) {
	if len(deleter) > 0 {
		h := p.Insert(key, value, size, deleter[0])
		h.Release()
	} else {
		h := p.Insert(key, value, size, nil)
		h.Release()
	}
}

// Return a new numeric id.  May be used by multiple clients who are
// sharing the same cache to partition the key space.  Typically the
// client will allocate a new id at startup and prepend the id to
// its cache keys.
func (p *LRUCache) NewId() uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.last_id++
	return p.last_id
}

// Insert a mapping from key->value into the cache and assign it
// the specified size against the total cache capacity.
//
// Return a handle that corresponds to the mapping.  The caller
// must call handle.Release() when the returned mapping is no
// longer needed.
//
// When the inserted entry is no longer needed, the key and
// value will be passed to "deleter".
func (p *LRUCache) Insert(key string, value interface{}, size int, deleter func(key string, value interface{})) Handle {
	p.mu.Lock()
	defer p.mu.Unlock()

	assert(key != "" && size > 0)
	if element := p.table[key]; element != nil {
		p.list.Remove(element)
		delete(p.table, key)

		h := element.Value.(*LRUHandle)
		p.unref(h)
	}

	h := &LRUHandle{
		c:            p,
		key:          key,
		value:        value,
		size:         int64(size),
		deleter:      deleter,
		time_created: time.Now(),
		refs:         2, // One from LRUCache, one for the returned handle
	}
	h.time_accessed.Store(time.Now())

	element := p.list.PushFront(h)
	p.table[key] = element
	p.size += h.size
	p.checkCapacity()
	return h
}

// If the cache has no mapping for "key", returns nil, false.
//
// Else return a handle that corresponds to the mapping.  The caller
// must call handle.Release() when the returned mapping is no
// longer needed.
func (p *LRUCache) Lookup(key string) (Handle, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil {
		return nil, false
	}

	p.list.MoveToFront(element)
	h := element.Value.(*LRUHandle)
	h.time_accessed.Store(time.Now())
	p.addref(h)
	return h, true
}

// If the cache has no mapping for "key", returns nil, false.
//
// Else return a handle that corresponds to the mapping and erase it.
// The caller must call handle.Release() when the returned mapping is no
// longer needed.
func (p *LRUCache) Take(key string) (Handle, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil {
		return nil, false
	}

	p.list.Remove(element)
	delete(p.table, key)

	h := element.Value.(*LRUHandle)
	return h, true
}

// If the cache contains entry for key, erase it.  Note that the
// underlying entry will be kept around until all existing handles
// to it have been released.
func (p *LRUCache) Erase(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil {
		return
	}

	p.list.Remove(element)
	delete(p.table, key)

	h := element.Value.(*LRUHandle)
	p.unref(h)
	return
}

// SetCapacity will set the capacity of the cache. If the capacity is
// smaller, and the current cache size exceed that capacity, the cache
// will be shrank.
func (p *LRUCache) SetCapacity(capacity int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	assert(capacity > 0)
	p.capacity = capacity
	p.checkCapacity()
}

// Stats returns a few stats on the cache.
func (p *LRUCache) Stats() (length, size, capacity int64, oldest time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if lastElem := p.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*LRUHandle).time_accessed.Load().(time.Time)
	}
	return int64(p.list.Len()), p.size, p.capacity, oldest
}

// StatsJSON returns stats as a JSON object in a string.
func (p *LRUCache) StatsJSON() string {
	if p == nil {
		return "{}"
	}
	l, s, c, o := p.Stats()
	return fmt.Sprintf(`{
	"Length": %v,
	"Size": %v,
	"Capacity": %v,
	"OldestAccess": "%v"
}`, l, s, c, o)
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

// Newest returns the insertion time of the newest element in the cache,
// or a IsZero() time if cache is empty.
func (p *LRUCache) Newest() (newest time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if frontElem := p.list.Front(); frontElem != nil {
		newest = frontElem.Value.(*LRUHandle).time_accessed.Load().(time.Time)
	}
	return
}

// Oldest returns the insertion time of the oldest element in the cache,
// or a IsZero() time if cache is empty.
func (p *LRUCache) Oldest() (oldest time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if lastElem := p.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*LRUHandle).time_accessed.Load().(time.Time)
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

func (p *LRUCache) addref(h *LRUHandle) {
	h.refs++
}

func (p *LRUCache) unref(h *LRUHandle) {
	assert(h.refs > 0)
	h.refs--
	if h.refs <= 0 {
		p.size -= h.size
		if h.deleter != nil {
			h.deleter(h.key, h.value)
		}
	}
}

func (p *LRUCache) checkCapacity() {
	// Partially duplicated from Delete
	// must keep the front element valid!!!
	for p.size > p.capacity && len(p.table) > 1 {
		delElem := p.list.Back()
		h := delElem.Value.(*LRUHandle)
		p.list.Remove(delElem)
		delete(p.table, h.key)
		p.unref(h)
	}
}

// Destroys all existing entries by calling the "deleter"
// function that was passed to the constructor.
func (p *LRUCache) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, element := range p.table {
		h := element.Value.(*LRUHandle)
		p.unref(h)
	}

	p.list = list.New()
	p.table = make(map[string]*list.Element)
	p.size = 0
	return
}

// Destroys all existing entries by calling the "deleter"
// function that was passed to the constructor.
// REQUIRES: all handles must have been released.
func (p *LRUCache) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, element := range p.table {
		h := element.Value.(*LRUHandle)
		assert(h.refs == 1, "h.refs = ", h.refs)
		p.unref(h)
	}

	p.list = nil
	p.table = nil
	p.size = 0
	return nil
}
