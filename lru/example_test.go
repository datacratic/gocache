// Copyright (c) 2014 Datacratic. All rights reserved.

package lru

import (
	"fmt"
)

func ExampleCache() {
	// create a cache with a limited number (2) of slots.
	cache := New(2)

	// fill
	cache.Set("key-a", "a")
	cache.Set("key-b", "b")

	// get
	if a, ok := cache.Get("key-a"); ok {
		fmt.Printf("%s\n", a)
	}

	if b, ok := cache.Get("key-b"); ok {
		fmt.Printf("%s\n", b)
	}

	// a new key will evict the least recently used i.e. "key-a"
	cache.Set("key-c", "c")

	// print it
	fmt.Printf("%s", cache)

	// Output:
	// a
	// b
	// key-c:c
	// key-b:b
}
