package externalapi

import (
	"sync"
	"time"
)

const cacheTTL time.Duration = 1 * time.Minute

type CacheEntry struct {
	Data      []byte
	ExpiresAt time.Time
}

var cache sync.Map

func GetFromCache(key string) ([]byte, bool) {
	val, ok := cache.Load(key)
	if !ok {
		return nil, false
	}
	entry := val.(CacheEntry)
	if time.Now().After(entry.ExpiresAt) {
		cache.Delete(key)
		return nil, false
	}
	return entry.Data, true
}

func SetCache(key string, data []byte) {
	cache.Store(key, CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(cacheTTL),
	})
}

