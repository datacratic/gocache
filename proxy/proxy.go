// Copyright (c) 2014 Datacratic. All rights reserved.

package proxy

import (
	"fmt"
	"hash/fnv"
	"io"
	"sync"
	"time"

	"github.com/datacratic/gocache/lru"
)

// DefaultPartitionCount defines the default number of partitions used by the proxy cache.
var DefaultPartitionCount = 64

// DefaultCacheSize defines the maximum number of cached requests per partition.
var DefaultCacheSize = 4096

// Cache implements a cache that acts as a proxy for requests.
type Cache struct {
	// PartitionCount contains the number of partitions used by the cache.
	// Will use DefaultPartitionCount if 0.
	// This also roughly correspond to the maximum of possible concurrent requests.
	PartitionCount int
	// CacheSize contains the maximum number of cached requests per partition.
	// Will use DefaultCacheSize if 0.
	// Note that the maximum amount of cached requests for the whole cache is PartitionCount*CacheSize.
	CacheSize int
	// Expiration contains the duration after which the request will be invalidated.
	// An expiration of 0 indicates that there the request never expires.
	Expiration time.Duration

	shards []shard
	once   sync.Once
}

func (cache *Cache) start() {
	count := cache.PartitionCount
	if count == 0 {
		count = DefaultPartitionCount
	}

	cache.shards = make([]shard, count)

	for i := 0; i != count; i++ {
		cache.shards[i].start(cache.CacheSize, cache.Expiration)
	}
}

// Metrics is used to record events during a cache request.
type Metrics struct {
	// Request is set for every cache request.
	Request bool
	// Hit is set when the cached value is returned.
	Hit bool
	// Expired is set when a new request is required because the existing cached value exists but is expired.
	Expired bool
	// Miss is set when no cached value exists and a new request is required.
	Miss bool
	// Done is set when a cache request succeeds.
	Done bool
	// Shard records the distribution of requests over partitions.
	Shard string
	// Latency measures the time elapsed for the cache request.
	Latency time.Duration
}

// Result defines the result of a request returned by the cache.
type Result struct {
	Metrics
	Request interface{}
}

// Producer defines an interface that performs the request when the cache needs to perform a request.
type Producer interface {
	Request(key string) (result interface{}, err error)
}

// ProducerFunc defines an helper to support the Producer interface.
type ProducerFunc func(key string) (interface{}, error)

// Request invokes the function literal.
func (f ProducerFunc) Request(key string) (interface{}, error) {
	return f(key)
}

// Get queries the cache and performs the request if necessary.
func (cache *Cache) Get(key string, producer Producer) (result Result, err error) {
	cache.once.Do(cache.start)

	t0 := time.Now()
	result.Request = true

	// hash the key and select the shard
	h := fnv.New32()
	io.WriteString(h, key)
	n := int(h.Sum32()) % len(cache.shards)
	result.Shard = fmt.Sprintf("%d", n)

	// send the query
	reply := make(chan error)
	cache.shards[n].queue <- &query{
		result:   &result,
		key:      key,
		producer: producer,
		reply:    reply,
	}

	// wait for the answer
	err = <-reply
	if err != nil {
		result.Done = true
	}

	result.Latency = time.Since(t0)
	return
}

type query struct {
	result   *Result
	key      string
	producer Producer
	reply    chan error
}

type shard struct {
	queue chan *query
	cache *lru.Cache
}

type entry struct {
	data   interface{}
	expiry time.Time
}

func (shard *shard) start(size int, expiration time.Duration) {
	shard.queue = make(chan *query)

	if size == 0 {
		size = DefaultCacheSize
	}

	shard.cache = lru.New(size)

	go func() {
		for query := range shard.queue {
			element, ok := shard.cache.Get(query.key)
			if !ok {
				query.result.Miss = true
			} else {
				item := element.(entry)

				if expiration != 0 && time.Now().After(item.expiry) {
					query.result.Expired = true
					shard.cache.Remove(element)
				} else {
					query.result.Hit = true
					query.result.Request = item.data

					query.reply <- nil
				}
			}

			data, err := query.producer.Request(query.key)
			if err == nil {
				query.result.Request = data
				shard.cache.Set(query.key, entry{
					data:   data,
					expiry: time.Now().Add(expiration),
				})
			}

			query.reply <- err
		}
	}()
}
