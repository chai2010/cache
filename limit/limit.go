// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package limit provides limit file opener support.
package limit

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/chai2010/cache"
)

var (
	ErrLimit = errors.New("limit: limit error!")
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type Opener struct {
	logger Logger
	open   func(name string) (f interface{}, err error)
	close  func(f interface{}) error
	cache  *cache.LRUCache
	limit  chan int // buffered chan
	wg     sync.WaitGroup
	mu     sync.Mutex
}

func NewOpener(
	open func(name string) (f interface{}, err error),
	close func(f interface{}) error,
	capacity int,
	log Logger,
) *Opener {
	assert(open != nil)
	assert(close != nil)
	assert(capacity > 0)

	p := &Opener{
		open:   open,
		close:  close,
		cache:  cache.NewLRUCache(int64(capacity)),
		limit:  make(chan int, capacity),
		logger: log,
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

func (p *Opener) Open(name string) (f interface{}, h io.Closer, err error) {
	assert(name != "")

	// get from cache
	if f, h, ok := p.cache.Lookup(name); ok {
		return f, h, nil
	}

	// limit opened files
	select {
	case p.limit <- 1:
		p.wg.Add(1)
	default:
		return nil, nil, ErrLimit
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
			if p.logger != nil {
				p.logger.Printf("limit: Opener.close(key=%q) failed, err = %v\n", err)
			}
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
