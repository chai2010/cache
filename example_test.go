// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"fmt"
	"log"

	"github.com/chai2010/cache"
)

func Example() {
	c := cache.NewLRUCache(100)
	defer c.Close()

	h1 := c.Insert("100", "101", 1, func(key string, value interface{}) {
		fmt.Printf("deleter(%q, %q)\n", key, value.(string))
	})
	v1 := h1.(cache.Handle).Value().(string)
	fmt.Printf("v1: %s\n", v1)
	h1.Close()

	h2, ok := c.Lookup("100")
	if !ok {
		log.Fatal("lookup failed!")
	}
	defer h2.Close()

	// h2 still valid after Erase
	c.Erase("100")
	v2 := h2.(cache.Handle).Value().(string)
	fmt.Printf("v2: %s\n", v2)

	// but new lookup will failed
	_, ok = c.Lookup("100")
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
