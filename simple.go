// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ingore

package main

import (
	"fmt"
    "log"

	"github.com/chai2010/cache"
)

func main() {
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
}
