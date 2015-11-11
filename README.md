# LRU Cache

[![Build Status](https://travis-ci.org/chai2010/cache.svg)](https://travis-ci.org/chai2010/cache)
[![GoDoc](https://godoc.org/github.com/chai2010/cache?status.svg)](https://godoc.org/github.com/chai2010/cache)


# Install

1. `go get github.com/chai2010/cache`
2. `go run hello.go`


# Example

## Simple GC Object

This is a simple example:

```Go
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
```

Output:

```
key1: value1
key2: value2
key3: null
key4: <nil>
Done
```

## Non GC Object

Support non GC object, such as `os.File` or some cgo memory.

```Go
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
	v2, h2 := c.Lookup("123")
	assert(h2 != nil)

	// remove
	c.Erase("123")

	// fetch failed
	_, h3 := c.Lookup("123")
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
```

Output:

```
id0: 1
id1: 2
id2: 3
user1(data:123)
user2(data:123)
invoke deleter(123) begin
deleter("123":"data:123")
invoke deleter(123) end
invoke deleter(h4) begin
deleter("abc":"data:abc")
invoke deleter(h4) end
StatsJSON: {
        "Length": 1,
        "Size": 8,
        "Capacity": 10,
        "OldestAccess": "2015-08-21 18:00:24.0119469 +0800 CST"
}
Done
deleter("456":"data:456")
```

# BUGS

Report bugs to <chaishushan@gmail.com>.

Thanks!
