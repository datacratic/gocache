// Copyright (c) 2014 Datacratic. All rights reserved.

package lru

import (
	"fmt"
	"testing"
)

func TestCache(t *testing.T) {
	cache := New(4)

	if cache.Len() != 0 {
		t.Fatalf("expected 0 instead of '%d'", cache.Len())
	}

	if ok := cache.Remove("a"); ok {
		t.Fatalf("not expected")
	}

	cache.Evict()

	if item, ok := cache.Get("a"); ok {
		t.Fatalf("not expecting '%d'", item)
	}

	cache.Set("a", "hello")
	cache.Set("b", "cruel")
	cache.Set("c", "black")
	cache.Set("d", "world")

	if cache.Len() != 4 {
		t.Fatalf("expected 4 instead of '%d'", cache.Len())
	}

	if b, ok := cache.Get("b"); !ok || b != "cruel" {
		t.Fatalf("expected to find 'cruel' instead of '%s'", b)
	}

	if c, ok := cache.Get("c"); !ok || c != "black" {
		t.Fatalf("expected to find 'black' instead of '%s'", c)
	}

	if e, ok := cache.Get("e"); ok {
		t.Fatalf("not expecting '%s'", e)
	}

	if ok := cache.Remove("a"); !ok {
		t.Fatalf("not expected")
	}

	cache.Set("c", "beats")
	cache.Set("f", "peace")
	cache.Set("g", "human")

	result := fmt.Sprintf("%s", cache)
	expect := "g:human\nf:peace\nc:beats\nb:cruel\n"
	if result != expect {
		t.Fatalf("not expecting:\n%s", result)
	}

	cache.Evict()
	cache.Evict()

	if f, ok := cache.Get("f"); !ok || f != "peace" {
		t.Fatalf("expected to find 'peace' instead of '%s'", f)
	}

	if g, ok := cache.Get("g"); !ok || g != "human" {
		t.Fatalf("expected to find 'human' instead of '%s'", g)
	}
}

func BenchmarkCacheSet(b *testing.B) {
	n := 4096
	m := n * 4

	cache := New(n)
	items := make([]string, m)

	for i := 0; i != m; i++ {
		items[i] = fmt.Sprintf("test-%d", i)
	}

	b.ResetTimer()

	k := 0
	for i := 0; i != b.N; i++ {
		k = (1664525*k + 1013904223) % m
		cache.Set(items[k], i)
	}
}
