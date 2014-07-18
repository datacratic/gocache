// Copyright (c) 2014 Datacratic. All rights reserved.

/*
Package lru provides a efficient LRU cache.

	// create a cache with a limited number (2) of slots.
	cache := lru.New(2)

	// fill
	cache.Set("key-a", "a")
	cache.Set("key-b", "b")

	// get
	a, ok := cache.Get("key-a")
	b, ok := cache.Get("key-b")

	// a new key will evict the least recently used i.e. "key-a"
	cache.Set("key-c", "c")

If limit is set to 0, the LRU cache will never evict an element on an insert
operation. It has to be done manually.
*/
package lru
