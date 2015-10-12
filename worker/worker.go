// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package worker provides ansyc worker support.
package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/chai2010/cache"
)

type Worker struct {
	tasks  *cache.LRUCache
	stoped chan bool
	wg     sync.WaitGroup
}

type _WorkerItem struct {
	task func()
	done bool
	mu   sync.Mutex
}

func newWorkerItem(task func()) *_WorkerItem {
	return &_WorkerItem{
		task: task,
	}
}

func (p *_WorkerItem) IsDone() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.done
}

func (p *_WorkerItem) DoTask() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.done && p.task != nil {
		p.done = true
		p.task()
	}
}

func NewWorker(taskCacheSize int) *Worker {
	assert(taskCacheSize > 0)

	p := &Worker{
		tasks: cache.NewLRUCache(int64(taskCacheSize)),
	}
	p.wg.Add(1)
	return p
}

func (p *Worker) AddTask(key interface{}, task func()) {
	assert(key != nil)
	assert(task != nil)

	skey := fmt.Sprintf("%q", key)
	if h, ok := p.tasks.Lookup(skey); ok {
		h.Release()
		return
	}
	p.tasks.PushFront(skey, newWorkerItem(task), 1, nil)
}

func (p *Worker) Start() {
	assert(p.stoped == nil)

	p.wg.Add(1)
	p.stoped = make(chan bool)

	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-p.stoped:
				return
			default:
				if p.tasks.Length() == 0 {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if h := p.tasks.Front(); h != nil {
					if h.TimeAccessed().Add(30 * time.Second).After(time.Now()) {
						if !h.Value().(*_WorkerItem).IsDone() {
							h.Value().(*_WorkerItem).DoTask()
						}
						p.tasks.MoveToBack(h.Key())
					} else {
						p.tasks.Erase(h.Key())
					}
					h.Release()
				}
			}
		}
	}()
}

func (p *Worker) Stop() {
	assert(p.stoped != nil)

	p.wg.Done()
	close(p.stoped)
	p.wg.Wait()
	p.stoped = nil
}

func assert(v bool, a ...interface{}) {
	if !v {
		if msg := fmt.Sprint(a...); msg != "" {
			panic(fmt.Sprintf("assert failed, %s!", msg))
		} else {
			panic(fmt.Sprintf("assert failed!"))
		}
	}
}
