// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/chai2010/cache"
)

func main() {
	cache := cache.NewLRUCache(10)
	defer cache.Close()

	// Get failed
	value, ok := cache.Get("123")
	assert(!ok)
	assert(value == nil)

	// add new
	cache.Set("123", "data:123", len("data:123"))

	// Get ok
	value, ok = cache.Get("123")
	assert(ok)
	assert(value.(string) == "data:123")

	// done
	fmt.Println("Done")
}

func assert(v bool, a ...interface{}) {
	if !v {
		panic(fmt.Sprint(a...))
	}
}
