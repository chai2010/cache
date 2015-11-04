// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package limit

import (
	"os"
	"testing"
)

func TestOpener(t *testing.T) {
	limitOpener := NewOpener(
		func(name string) (f interface{}, err error) {
			file, err := os.Open(name)
			return file, nil
		},
		func(f interface{}) error {
			return f.(*os.File).Close()
		},
		2,
		nil,
	)
	defer limitOpener.Close()

	f1, h1, err := limitOpener.Open("limit.go")
	tAssertNil(t, err)
	defer h1.Release()

	f2, h2, err := limitOpener.Open("limit.go")
	tAssertNil(t, err)
	defer h2.Release()

	tAssert(t, f2 == f1)

	f3, h3, err := limitOpener.Open("example_test.go")
	tAssertNil(t, err)
	defer h3.Release()

	f4, h4, err := limitOpener.Open("testing_test.go")
	tAssert(t, err == ErrLimit)
	tAssert(t, f4 == nil)
	tAssert(t, h4 == nil)

	_ = f1.(*os.File)
	_ = f2.(*os.File)
	_ = f3.(*os.File)
	_ = f4 // f4 is nil
}
