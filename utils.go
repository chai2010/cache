// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"fmt"
)

func assert(v bool, a ...interface{}) {
	if !v {
		panic(fmt.Sprint(a...))
	}
}
