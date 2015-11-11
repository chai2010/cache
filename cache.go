// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"io"
)

// Cache is a thread-safe cache.
//
// See https://github.com/google/leveldb/blob/master/include/leveldb/cache.h
type Cache interface {
	// Return a new numeric id.  May be used by multiple clients who are
	// sharing the same cache to partition the key space.  Typically the
	// client will allocate a new id at startup and prepend the id to
	// its cache keys.
	NewId() uint64

	// Insert a mapping from key->value into the cache and assign it
	// the specified size against the total cache capacity.
	//
	// Return a handle that corresponds to the mapping.  The caller
	// must call handle.Close() when the returned mapping is no
	// longer needed.
	//
	// When the inserted entry is no longer needed, the key and
	// value will be passed to "deleter".
	Insert(key string, value interface{}, size int, deleter func(key string, value interface{})) (handle io.Closer)

	// If the cache has no mapping for "key", returns nil, nil.
	//
	// Else return a handle that corresponds to the mapping.  The caller
	// must call handle.Close() when the returned mapping is no
	// longer needed.
	Lookup(key string) (value interface{}, handle io.Closer)

	// If the cache contains entry for key, erase it.  Note that the
	// underlying entry will be kept around until all existing handles
	// to it have been released.
	Erase(key string)

	// Destroys all existing entries by calling the "deleter"
	// function that was passed to the constructor.
	// REQUIRES: all handles must have been released.
	Close() error
}

// New creates a new empty cache with the given capacity.
func New(capacity int64) Cache {
	return NewLRUCache(capacity)
}
