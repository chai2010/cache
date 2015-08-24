// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"fmt"
	"log"

	"github.com/chai2010/cache"
)

func ExampleCache() {
	cache := cache.New(100)
	defer cache.Close()

	h1 := cache.Insert("100", "101", 1, func(key string, value interface{}) {
		fmt.Printf("deleter(%q, %q)\n", key, value.(string))
	})
	v1 := h1.Value().(string)
	fmt.Printf("v1: %s\n", v1)
	h1.Release()

	h2, ok := cache.Lookup("100")
	if !ok {
		log.Fatal("lookup failed!")
	}
	defer h2.Release()

	// h2 still valid after Erase
	cache.Erase("100")
	v2 := h2.Value().(string)
	fmt.Printf("v2: %s\n", v2)

	// but new lookup will failed
	_, ok = cache.Lookup("100")
	if ok {
		log.Fatal("lookup succeed!")
	}

	fmt.Println("Done")
	// Output:
	// v1: 101
	// v2: 101
	// Done
	// deleter("100", "101")
}

func Example() {
	cache := cache.NewLRUCache(100)
	defer cache.Close()

	h1 := cache.Insert("100", "101", 1, func(key string, value interface{}) {
		fmt.Printf("deleter(%q, %q)\n", key, value.(string))
	})
	v1 := h1.Value().(string)
	fmt.Printf("v1: %s\n", v1)
	h1.Release()

	h2, ok := cache.Lookup("100")
	if !ok {
		log.Fatal("lookup failed!")
	}
	defer h2.Release()

	// h2 still valid after Erase
	cache.Erase("100")
	v2 := h2.Value().(string)
	fmt.Printf("v2: %s\n", v2)

	// but new lookup will failed
	_, ok = cache.Lookup("100")
	if ok {
		log.Fatal("lookup succeed!")
	}

	fmt.Println("Done")
	// Output:
	// v1: 101
	// v2: 101
	// Done
	// deleter("100", "101")
}

func ExampleHandle() {
	c := cache.New(1)
	defer c.Close()

	type Value struct {
		V string
	}

	todoList := make(chan cache.Handle, 1)
	h1 := c.Insert("100", &Value{V: "101"}, 1, func(key string, value interface{}) {
		fmt.Printf("deleter(%q, %q)\n", key, value.(*Value).V)
		value.(*Value).V = "nil"
	})
	fmt.Printf("h1: %s\n", h1.Value().(*Value).V)
	todoList <- h1.Retain()
	h1.Release()

	c.Erase("100")

	h2 := <-todoList
	fmt.Printf("h2: %s\n", h2.Value().(*Value).V)
	h2.Release()

	// Output:
	// h1: 101
	// h2: 101
	// deleter("100", "101")
}
