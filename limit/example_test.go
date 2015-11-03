// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package limit_test

import (
	"fmt"
	"os"

	"github.com/chai2010/cache/limit"
)

func Example() {
	limitOpener := limit.NewOpener(
		func(name string) (f interface{}, err error) {
			file, err := os.Open(name)
			return file, nil
		},
		func(f interface{}) error {
			return f.(*os.File).Close()
		},
		500,
		nil,
	)
	defer limitOpener.Close()

	f1, h1, err := limitOpener.Open("limit.go")
	assert(err == nil)

	f2, h2, err := limitOpener.Open("limit.go")
	assert(err == nil)

	f3, h3, err := limitOpener.Open("example_test.go")
	assert(err == nil)

	_, _, err = limitOpener.Open("unknown.go")
	assert(err != nil)

	_ = f1
	_ = f2
	_ = f3

	h1.Release()
	h2.Release()
	h3.Release()
}

func assert(v bool, a ...interface{}) {
	if !v {
		if msg := fmt.Sprint(a...); msg != "" {
			panic(fmt.Sprintf("assert failed, %s!", msg))
		} else {
			panic(fmt.Sprintf("assert failed!"))
		}
	}
}
