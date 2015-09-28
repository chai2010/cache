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
	cache := cache.NewLRUCache(100)
	defer cache.Close()

	cache.Set("key1", "value1", 1)
	v := cache.Value("key1").(string)

	fmt.Println("key1:", v)
	// Output:
	// key1: value1
}

func ExampleLRUCache_stack() {
	cache := cache.NewLRUCache(100)
	defer cache.Close()

	cache.PushBack("key1", "value1", 1, nil)
	cache.PushBack("key2", "value2", 1, nil)
	cache.PushBack("key3", "value3", 1, nil)

	fmt.Println("front key:", cache.FrontKey())
	fmt.Println("front value:", cache.FrontValue().(string))
	fmt.Println("back key:", cache.BackKey())
	fmt.Println("back value:", cache.BackValue().(string))

	cache.RemoveFront()
	cache.RemoveBack()

	fmt.Println("front:", cache.FrontValue().(string))
	fmt.Println("back:", cache.BackValue().(string))

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
						h.Release()
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

func ExampleLRUHandle() {
	c := cache.NewLRUCache(1)
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

	value2 := cache.Value("key2", "null").(string)
	fmt.Println("key2:", value2)

	value3 := cache.Value("key3", "null").(string)
	fmt.Println("key3:", value3)

	fmt.Println("Done")
	// Output:
	// key2: value2
	// key3: null
	// Done
	// deleter("key2", "value2")
}
