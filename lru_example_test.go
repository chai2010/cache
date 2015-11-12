// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/chai2010/cache"
)

func ExampleLRUCache_simple() {
	c := cache.NewLRUCache(100)
	defer c.Close()

	c.Set("key1", "value1", 1)
	v := c.Value("key1").(string)

	fmt.Println("key1:", v)
	// Output:
	// key1: value1
}

func ExampleLRUCache_stack() {
	c := cache.NewLRUCache(100)
	defer c.Close()

	c.PushBack("key1", "value1", 1, nil)
	c.PushBack("key2", "value2", 1, nil)
	c.PushBack("key3", "value3", 1, nil)

	fmt.Println("front key:", c.FrontKey())
	fmt.Println("front value:", c.FrontValue().(string))
	fmt.Println("back key:", c.BackKey())
	fmt.Println("back value:", c.BackValue().(string))

	c.RemoveFront()
	c.RemoveBack()

	fmt.Println("front:", c.FrontValue().(string))
	fmt.Println("back:", c.BackValue().(string))

	// Output:
	// front key: key1
	// front value: value1
	// back key: key3
	// back value: value3
	// front: value2
	// back: value2
}

func ExampleLRUCache_realtimeTodoWorker() {
	result := make(map[int]int)

	todo := cache.NewLRUCache(8) // max todo size
	defer todo.Close()

	var wg sync.WaitGroup
	var closed = make(chan bool)

	// background worker
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-closed:
					return
				default:
					if h := todo.PopFront(); h != nil {
						h.Value().(func())()
						h.Close()
					}
				}
			}
		}()
	}

	// gen todo list
	for i := 0; i < 10; i++ {
		id := i
		if _, ok := result[id]; !ok {
			todo.PushFront(
				fmt.Sprintf("id:%d", id),
				func() { time.Sleep(time.Second / 10); result[id] = id * 1000 },
				1,
				nil,
			)
		}
	}
	time.Sleep(time.Second)

	// print result
	// todo work is a realtime worker, maybe lost work
	for k, v := range result {
		fmt.Printf("result[%d]:%d\n", k, v)
		time.Sleep(time.Second / 10)
	}

	// wait
	close(closed)
	wg.Wait()

	// ingore output
}

func ExampleLRUCache_handle() {
	c := cache.NewLRUCache(100)
	defer c.Close()

	h1 := c.Insert("100", "101", 1, func(key string, value interface{}) {
		fmt.Printf("deleter(%q, %q)\n", key, value.(string))
	})
	v1 := h1.(*cache.LRUHandle).Value().(string)
	fmt.Printf("v1: %s\n", v1)
	h1.Close()

	v2, h2, ok := c.Lookup("100")
	if !ok {
		log.Fatal("lookup failed!")
	}
	defer h2.Close()

	// h2 still valid after Erase
	c.Erase("100")
	fmt.Printf("v2: %s\n", v2.(string))

	// but new lookup will failed
	if _, _, ok := c.Lookup("100"); ok {
		log.Fatal("lookup succeed!")
	}

	fmt.Println("Done")
	// Output:
	// v1: 101
	// v2: 101
	// Done
	// deleter("100", "101")
}

func ExampleLRUHandle() {
	c := cache.NewLRUCache(1)
	defer c.Close()

	type Value struct {
		V string
	}

	todoList := make(chan *cache.LRUHandle, 1)
	h1 := c.Insert("100", &Value{V: "101"}, 1, func(key string, value interface{}) {
		fmt.Printf("deleter(%q, %q)\n", key, value.(*Value).V)
		value.(*Value).V = "nil"
	})
	fmt.Printf("h1: %s\n", h1.(*cache.LRUHandle).Value().(*Value).V)
	todoList <- h1.(*cache.LRUHandle).Retain()
	h1.Close()

	c.Erase("100")

	h2 := <-todoList
	fmt.Printf("h2: %s\n", h2.Value().(*Value).V)
	h2.Close()

	// Output:
	// h1: 101
	// h2: 101
	// deleter("100", "101")
}

func ExampleLRUCache_getAndSet() {
	c := cache.NewLRUCache(100)
	defer c.Close()

	// set dont return handle
	c.Set("key1", "value1", len("value1"))

	// set's deleter is optional
	c.Set("key2", "value2", len("value2"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q, %q)\n", key, value.(string))
	})

	value, ok := c.Get("key1")
	if !ok {
		log.Fatal("not found key1")
	}
	if v := value.(string); v != "value1" {
		log.Fatal("not equal value1")
	}

	value2 := c.Value("key2", "null").(string)
	fmt.Println("key2:", value2)

	value3 := c.Value("key3", "null").(string)
	fmt.Println("key3:", value3)

	fmt.Println("Done")
	// Output:
	// key2: value2
	// key3: null
	// Done
	// deleter("key2", "value2")
}
