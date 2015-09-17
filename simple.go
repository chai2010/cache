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
	cache := cache.NewLRUCache(100)
	defer cache.Close()

	cache.Set("key1", "value1", 1)
	value1 := cache.Value("key1").(string)
	fmt.Println("key1:", value1)

	cache.Set("key2", "value2", 1)
	value2 := cache.Value("key2", "null").(string)
	fmt.Println("key2:", value2)

	value3 := cache.Value("key3", "null").(string)
	fmt.Println("key3:", value3)

	value4 := cache.Value("key4") // value4 is nil
	fmt.Println("key4:", value4)

	fmt.Println("Done")
}
