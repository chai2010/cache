// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ingore

package main

import (
	"fmt"

	"github.com/chai2010/cache"
)

func main() {
	cache := cache.NewLRUCache(10)
	defer cache.Close()

	id0 := cache.NewId()
	id1 := cache.NewId()
	id2 := cache.NewId()
	fmt.Println("id0:", id0)
	fmt.Println("id1:", id1)
	fmt.Println("id2:", id2)

	// add new
	h1 := cache.Insert("123", "data:123", len("data:123"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q:%q)\n", key, value)
	})

	// fetch ok
	h2, ok := cache.Lookup("123")
	assert(ok)

	// remove
	cache.Erase("123")

	// fetch failed
	h3, ok := cache.Lookup("123")
	assert(h3 == nil)
	assert(!ok)

	// h1&h2 still valid!
	fmt.Printf("user1(%s)\n", h1.Value().(string))
	fmt.Printf("user2(%s)\n", h2.Value().(string))

	// release h1
	// because the h2 handle the value, so the deleter is not ivoked!
	h1.Release()

	// invoke the deleter
	fmt.Println("invoke deleter(123) begin")
	h2.Release()
	fmt.Println("invoke deleter(123) end")

	// add new
	h4 := cache.Insert("abc", "data:abc", len("data:abc"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q:%q)\n", key, value)
	})
	// release h4
	// because the cache handle the value, so the deleter is not ivoked!
	h4.Release()

	// cache length
	length := cache.Length()
	assert(length == 1)

	// cache size
	size := cache.Size()
	assert(size == 8, "size:", size)

	// add h5
	// this will cause the capacity(10) overflow, so the h4 deleter will be invoked
	fmt.Println("invoke deleter(h4) begin")
	h5 := cache.Insert("456", "data:456", len("data:456"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q:%q)\n", key, value)
	})
	fmt.Println("invoke deleter(h4) end")

	// must release all handles
	h5.Release()

	// stats
	fmt.Println("StatsJSON:", cache.StatsJSON())

	// done
	fmt.Println("Done")
}

func assert(v bool, a ...interface{}) {
	if !v {
		panic(fmt.Sprint(a...))
	}
}
