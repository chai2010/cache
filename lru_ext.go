// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"container/list"
	"time"
)

func (p *LRUCache) HasKey(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, ok := p.table[key]
	return ok
}

func (p *LRUCache) FrontKey() (key string) {
	if h := p.Front(); h != nil {
		key = h.Key()
		h.Close()
		return key
	}
	return ""
}

func (p *LRUCache) BackKey() (key string) {
	if h := p.Back(); h != nil {
		key = h.Key()
		h.Close()
		return key
	}
	return ""
}

func (p *LRUCache) FrontValue(defaultValue ...interface{}) (value interface{}) {
	if h := p.Front(); h != nil {
		value = h.Value()
		h.Close()
		return
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return nil
	}
}

func (p *LRUCache) BackValue(defaultValue ...interface{}) (value interface{}) {
	if h := p.Back(); h != nil {
		value = h.Value()
		h.Close()
		return
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return nil
	}
}

func (p *LRUCache) RemoveFront() {
	if h := p.PopFront(); h != nil {
		h.Close()
	}
}

func (p *LRUCache) RemoveBack() {
	if h := p.PopBack(); h != nil {
		h.Close()
	}
}

func (p *LRUCache) Front() (h *LRUHandle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var element *list.Element
	if element = p.list.Front(); element == nil {
		return
	}

	h = element.Value.(*LRUHandle)
	p.addref(h)
	return
}

func (p *LRUCache) Back() (h *LRUHandle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var element *list.Element
	if element = p.list.Back(); element == nil {
		return
	}

	h = element.Value.(*LRUHandle)
	p.addref(h)
	return
}

func (p *LRUCache) PushFront(key string, value interface{}, size int, deleter func(key string, value interface{})) {
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
		refs:         1, // Only one from LRUCache, no returned handle
	}
	h.time_accessed.Store(time.Now())

	element := p.list.PushFront(h)
	p.table[key] = element
	p.size += h.size
	p.checkCapacity()
	return
}

func (p *LRUCache) PushBack(key string, value interface{}, size int, deleter func(key string, value interface{})) {
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
		refs:         1, // Only one from LRUCache, no returned handle
	}
	h.time_accessed.Store(time.Now())

	element := p.list.PushBack(h)
	p.table[key] = element
	p.size += h.size
	p.checkCapacity()
	return
}

func (p *LRUCache) PopBack() (h *LRUHandle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var element *list.Element
	if element = p.list.Back(); element == nil {
		return
	}

	h = element.Value.(*LRUHandle)
	delete(p.table, h.Key())
	p.list.Remove(element)
	return
}

func (p *LRUCache) PopFront() (h *LRUHandle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var element *list.Element
	if element = p.list.Front(); element == nil {
		return
	}

	h = element.Value.(*LRUHandle)
	delete(p.table, h.Key())
	p.list.Remove(element)
	return
}

func (p *LRUCache) MoveToFront(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil {
		return
	}

	p.list.MoveToFront(element)
	return
}

func (p *LRUCache) MoveToBack(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil {
		return
	}

	p.list.MoveToBack(element)
	return
}
