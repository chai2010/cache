// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package cache provides LevelDB style LRU cache.

The cached element is shared and thread safe.

This is s simple example:

    package main

    import (
        "fmt"
        "os"

        "github.com/chai2010/cache"
    )

    func main() {
        cache := cache.New(500)
        defer cache.Close()

        var wg sync.WaitGroup
        for i := 0; i < 1000; i++ {
            wg.Add(1)
            id := i % 100
            go func() {
                defer wg.Done()

                key := fmt.Sprintf("work-%d", id)
                h, ok := cache.Lookup(key)
                if !ok {
                    f, _ := os.Open(key)
                    h = cache.Insert(key, f, 1, func(key string, value interface{}) {
                        f := value.(*os.File)
                        f.Close()
                    })
                }
                defer h.Release()
                ReadFile(f)
            } ()
        }
        wg.Wait()
    }
*/
package cache // import "github.com/chai2010/cache"
