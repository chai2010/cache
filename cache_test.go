// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

const (
	tCacheSize = 1000
)

var (
	t_deleted_keys_   []int
	t_deleted_values_ []int
	t_cache_          = tNewTCache(tCacheSize)
)

type TCache struct {
	Cache
}

func tNewTCache(capacity int64) *TCache {
	return &TCache{
		Cache: NewLRUCache(capacity),
	}
}

func (p *TCache) Lookup(key int) int {
	return -1
}

func (p *TCache) Insert(key, value, size int) {
	//
}

func (p *TCache) Erase(key int) {
}

func (p *TCache) onDeleter(key string, value interface{}) {
	//
}

func TestCache_NewId(t *testing.T) {
	a := t_cache_.NewId()
	b := t_cache_.NewId()
	tAssert(t, a != b)
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
