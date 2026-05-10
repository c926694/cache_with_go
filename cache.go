package cache

import "cache/lru"

type Cache struct {
	lru *lru.LRUCache
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

func (c *Cache) UsedBytes() int64 {
	return c.lru.UsedBytes()
}
