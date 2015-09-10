// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"fmt"
	"log"

	"github.com/chai2010/cache"
)

func ExampleLRUCache_simple() {
	cache := cache.NewLRUCache(100)
	defer cache.Close()

	cache.Set("key1", "value1", 1)
	value, ok := cache.Get("key1")
	if !ok {
		log.Fatal("not found key1")
	}
	if v := value.(string); v != "value1" {
		log.Fatal("not equal value1")
	}

	fmt.Println("Done")
	// Output:
	// Done
}

func ExampleLRUCache_getAndSet() {
	cache := cache.NewLRUCache(100)
	defer cache.Close()

	// set dont return handle
	cache.Set("key1", "value1", len("value1"))

	// set's deleter is optional
	cache.Set("key2", "value2", len("value2"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q, %q)\n", key, value.(string))
	})

	value, ok := cache.Get("key1")
	if !ok {
		log.Fatal("not found key1")
	}
	if v := value.(string); v != "value1" {
		log.Fatal("not equal value1")
	}

	fmt.Println("Done")
	// Output:
	// Done
	// deleter("key2", "value2")
}
