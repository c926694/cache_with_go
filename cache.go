package cache

import (
	"cache/lru"
	"time"
)

type Cache struct {
	lru *lru.LRUCacheImpl
}

func NewCache(maxBytes int64) *Cache {
	return &Cache{
		lru: lru.NewLRU(maxBytes),
	}
}

func (c *Cache) Get(key string) (ByteView, bool) {
	v, ok := c.lru.Get(key)
	if !ok {
		return ByteView{}, false
	}
	return v.(ByteView), true
}

func (c *Cache) Set(key string, value ByteView) {
	c.lru.Set(key, value)
}

func (c *Cache) SetWithExpiration(key string, value ByteView, expiration time.Duration) {
	c.lru.SetWithExpiration(key, value, expiration)
}

func (c *Cache) Delete(key string) bool {
	return c.lru.Delete(key)
}

func (c *Cache) UsedBytes() int64 {
	return c.lru.UsedBytes()
}
