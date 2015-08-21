// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func tAssert(t *testing.T, v bool, a ...interface{}) {
	if !v {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprint(a...); msg != "" {
			t.Fatalf("%s:%d: tAssert failed, %s!", file, line, msg)
		} else {
			t.Fatalf("%s:%d: tAssert failed!", file, line)
		}
	}
}

func tAssertf(t *testing.T, v bool, format string, a ...interface{}) {
	if !v {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprintf(format, a...); msg != "" {
			t.Fatalf("%s:%d: tAssert failed, %s!", file, line, msg)
		} else {
			t.Fatalf("%s:%d: tAssert failed!", file, line)
		}
	}
}

func tAssertEQ(t *testing.T, a, b interface{}, s ...interface{}) {
	if !reflect.DeepEqual(a, b) {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprint(s...); msg != "" {
			t.Fatalf("%s:%d: tAssertEQ failed, %v != %v, %s!", file, line, a, b, msg)
		} else {
			t.Fatalf("%s:%d: tAssertEQ failed, %v != %v!", file, line, a, b)
		}
	}
}

func tAssertNE(t *testing.T, a, b interface{}, s ...interface{}) {
	if reflect.DeepEqual(a, b) {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprint(s...); msg != "" {
			t.Fatalf("%s:%d: tAssertNE failed, %v == %v, %s!", file, line, a, b, msg)
		} else {
			t.Fatalf("%s:%d: tAssertNE failed, %v == %v!", file, line, a, b)
		}
	}
}

func tAssertLE(t *testing.T, a, b int, s ...interface{}) {
	if !(a <= b) {
		file, line := tCallerFileLine(1)
		if msg := fmt.Sprint(s...); msg != "" {
			t.Fatalf("%s:%d: tAssertNE failed, %v > %v, %s!", file, line, a, b, msg)
		} else {
			t.Fatalf("%s:%d: tAssertNE failed, %v > %v!", file, line, a, b)
		}
	}
}

func tCallerFileLine(skip int) (file string, line int) {
	_, file, line, ok := runtime.Caller(skip + 1)
	if ok {
		// Truncate file name at last file name separator.
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
	} else {
		file = "???"
		line = 1
	}
	return
}

func tAtoi(s string, defaultV int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultV
}
