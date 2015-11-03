// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package limit provides limit file opener support.
package limit

import (
	"fmt"
	"log"
	"sync"

	"github.com/chai2010/cache"
)

type Opener struct {
	open  func(name string) (f interface{}, err error)
	close func(f interface{}) error
	cache *cache.LRUCache
	limit chan int // buffered chan
	wg    sync.WaitGroup
	mu    sync.Mutex
}

func NewOpener(
	open func(name string) (f interface{}, err error),
	close func(f interface{}) error,
	capacity int,
) *Opener {
	assert(capacity > 0)

	p := &Opener{
		open:  open,
		close: close,
		cache: cache.NewLRUCache(int64(capacity)),
		limit: make(chan int, capacity),
	}
	p.wg.Add(1)
	return p
}

func (p *Opener) Close() error {
	p.cache.Close()
	p.wg.Done()
	p.wg.Wait()
	return nil
}

func (p *Opener) Open(name string) (f interface{}, h cache.Handle, err error) {
	// get from cache
	h, ok := p.cache.Lookup(name)
	if ok {
		f = h.Value()
		return
	}

	// limit opened files
	select {
	case p.limit <- 1:
		p.wg.Add(1)
	default:
		err = fmt.Errorf("limit: Opener.Open(name=%q): limit error!", name)
		return
	}

	defer func() {
		if err != nil {
			p.wg.Done()
			<-p.limit
		}
	}()

	// open file
	if f, err = p.open(name); err != nil {
		return
	}

	// put to cache
	h = p.cache.Insert(name, f, 1, func(key string, value interface{}) {
		if err := p.close(value); err != nil {
			log.Printf("limit: Opener.close(key=%q) failed, err = %v\n", err)
		}
		p.wg.Done()
		<-p.limit
		return
	})

	// OK
	return
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
