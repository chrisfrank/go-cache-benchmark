package main

import (
	"go-cache-benchmark/fifo"
	"log"
)

func NewFIFO(size int) Cache {
	store, err := fifo.New[string, any](size)
	if err != nil {
		log.Fatal(err)
	}
	return &FIFOCache{store}
}

type FIFOCache struct {
	store fifo.Cache[string, any]
}

func (f *FIFOCache) Name() string   { return "fifo" }
func (f *FIFOCache) Set(key string) { f.store.Add(key, "") }
func (f *FIFOCache) Close()         {}
func (f *FIFOCache) Get(key string) bool {
	_, ok := f.store.Get(key)
	return ok
}
