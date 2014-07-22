// Copyright (c) 2014 Datacratic. All rights reserved.

package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCache(t *testing.T) {
	var a int32
	var b int32

	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hello":
			atomic.AddInt32(&a, 1)
			w.Write([]byte("world"))
		case "/jack":
			atomic.AddInt32(&b, 1)
			w.Write([]byte("jane"))
		default:
			w.WriteHeader(http.StatusOK)
		}
	})

	h := httptest.NewServer(endpoint)

	p := ProducerFunc(func(key string) (result interface{}, err error) {
		url := fmt.Sprintf("%s/%s", h.URL, key)

		r, err := http.Get(url)
		if err != nil {
			return
		}

		result, err = ioutil.ReadAll(r.Body)
		r.Body.Close()

		return
	})

	cache := Cache{}

	world, err := cache.Get("hello", p)
	if err != nil {
		t.Fatal(err)
	}

	if w := world.Request.([]byte); string(w) != "world" {
		t.Fatalf("expected to get back 'world' instead of '%s'", string(w))
	}

	if world.Miss != true {
		t.Fatalf("expected cache miss")
	}

	jack, err := cache.Get("jack", p)
	if err != nil {
		t.Fatal(err)
	}

	if j := jack.Request.([]byte); string(j) != "jane" {
		t.Fatalf("expected to get back 'jane' instead of '%s'", string(j))
	}

	if jack.Miss != true {
		t.Fatalf("expected cache miss")
	}

	second, err := cache.Get("hello", p)
	if err != nil {
		t.Fatal(err)
	}

	if w := second.Request.([]byte); string(w) != "world" {
		t.Fatalf("expected to get back 'world' instead of '%s'", string(w))
	}

	if second.Hit != true {
		t.Fatal("expected cache hit")
	}
}
