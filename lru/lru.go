// Copyright (c) 2014 Datacratic. All rights reserved.

package lru

import (
	"bytes"
	"container/list"
	"fmt"
)

// Cache implements an LRU cache.
type Cache struct {
	// Size defines the maximum number of element in the cache. There is no limits if 0.
	Size int

	hash map[interface{}]*list.Element
	list *list.List
}

// Key defines any comparable value.
type Key interface{}

type element struct {
	key   Key
	value interface{}
}

// New creates a new LRU cache. If size is 0, the cache has no limits and eviction must be done manually.
func New(size int) (result *Cache) {
	result = &Cache{
		Size: size,
	}

	return
}

// Len returns the number of elements in the cache.
func (cache *Cache) Len() (result int) {
	if cache.list != nil {
		result = cache.list.Len()
	}

	return
}

// Set inserts a new element or replaces an existing element in the cache.
func (cache *Cache) Set(key Key, value interface{}) {
	if cache.hash == nil {
		cache.hash = make(map[interface{}]*list.Element)
		cache.list = list.New()
	}

	if item, ok := cache.hash[key]; ok {
		cache.list.MoveToFront(item)
		item.Value.(*element).value = value
		return
	}

	entry := cache.list.PushFront(&element{
		key:   key,
		value: value,
	})

	cache.hash[key] = entry

	if cache.Size != 0 && cache.list.Len() > cache.Size {
		cache.Evict()
	}
}

// Get returns an element from the cache.
func (cache *Cache) Get(key Key) (value interface{}, ok bool) {
	if cache.hash == nil {
		return
	}

	item, ok := cache.hash[key]
	if ok {
		cache.list.MoveToFront(item)
		value = item.Value.(*element).value
	}

	return
}

// Remove clears an element from the cache.
func (cache *Cache) Remove(key Key) (ok bool) {
	if cache.hash == nil {
		ok = false
		return
	}

	item, ok := cache.hash[key]
	if ok {
		cache.list.Remove(item)
		delete(cache.hash, item.Value.(*element).key)
	}

	return
}

// Evict removes the least recently used element from the cache.
func (cache *Cache) Evict() {
	if cache.hash == nil {
		return
	}

	item := cache.list.Back()
	if item != nil {
		cache.list.Remove(item)
		delete(cache.hash, item.Value.(*element).key)
	}
}

func (cache *Cache) String() string {
	buf := bytes.Buffer{}
	for i := cache.list.Front(); i != nil; i = i.Next() {
		item := i.Value.(*element)
		buf.WriteString(fmt.Sprintf("%s:%s\n", item.key, item.value))
	}

	return buf.String()
}
