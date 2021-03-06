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
	c := cache.NewLRUCache(10)
	defer c.Close()

	id0 := c.NewId()
	id1 := c.NewId()
	id2 := c.NewId()
	fmt.Println("id0:", id0)
	fmt.Println("id1:", id1)
	fmt.Println("id2:", id2)

	// add new
	v1 := "data:123"
	h1 := c.Insert("123", "data:123", len("data:123"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q:%q)\n", key, value)
	})

	// fetch ok
	v2, h2, ok := c.Lookup("123")
	assert(ok)
	assert(h2 != nil)

	// remove
	c.Erase("123")

	// fetch failed
	_, h3, ok := c.Lookup("123")
	assert(!ok)
	assert(h3 == nil)

	// h1&h2 still valid!
	fmt.Printf("user1(%s)\n", v1)
	fmt.Printf("user2(%s)\n", v2.(string))

	// release h1
	// because the h2 handle the value, so the deleter is not ivoked!
	h1.Close()

	// invoke the deleter
	fmt.Println("invoke deleter(123) begin")
	h2.Close()
	fmt.Println("invoke deleter(123) end")

	// add new
	h4 := c.Insert("abc", "data:abc", len("data:abc"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q:%q)\n", key, value)
	})
	// release h4
	// because the cache handle the value, so the deleter is not ivoked!
	h4.Close()

	// cache length
	length := c.Length()
	assert(length == 1)

	// cache size
	size := c.Size()
	assert(size == 8, "size:", size)

	// add h5
	// this will cause the capacity(10) overflow, so the h4 deleter will be invoked
	fmt.Println("invoke deleter(h4) begin")
	h5 := c.Insert("456", "data:456", len("data:456"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q:%q)\n", key, value)
	})
	fmt.Println("invoke deleter(h4) end")

	// must release all handles
	h5.Close()

	// stats
	fmt.Println("StatsJSON:", c.StatsJSON())

	// done
	fmt.Println("Done")
}

func assert(v bool, a ...interface{}) {
	if !v {
		panic(fmt.Sprint(a...))
	}
}
