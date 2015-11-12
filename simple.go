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
	c := cache.NewLRUCache(100)
	defer c.Close()

	c.Set("key1", "value1", 1)
	value1 := c.Value("key1").(string)
	fmt.Println("key1:", value1)

	c.Set("key2", "value2", 1)
	value2 := c.Value("key2", "null").(string)
	fmt.Println("key2:", value2)

	value3 := c.Value("key3", "null").(string)
	fmt.Println("key3:", value3)

	value4 := c.Value("key4") // value4 is nil
	fmt.Println("key4:", value4)

	fmt.Println("Done")
}
