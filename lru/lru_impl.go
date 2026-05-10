package lru

import (
	"container/list"
	"sync"
)

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

type LRUCacheImpl struct {
	maxBytes  int64
	usedBytes int64
	ll        *list.List
	cache     map[string]*list.Element
	mu        sync.Mutex
}

func NewLRU(maxBytes int64) *LRUCacheImpl {
	return &LRUCacheImpl{
		maxBytes: maxBytes,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
	}
}

func (c *LRUCacheImpl) Get(key string) (Value, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ele, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	c.ll.MoveToFront(ele)
	return ele.Value.(*entry).value, true
}

func (c *LRUCacheImpl) Set(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		c.ll.MoveToFront(ele)
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.removeOldestLocked()
	}
}

func (c *LRUCacheImpl) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	ele, ok := c.cache[key]
	if !ok {
		return false
	}

	c.removeElementLocked(ele)
	return true
}

func (c *LRUCacheImpl) UsedBytes() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.usedBytes
}

func (c *LRUCacheImpl) removeOldestLocked() {
	ele := c.ll.Back()
	if ele == nil {
		return
	}

	c.removeElementLocked(ele)
}

func (c *LRUCacheImpl) removeElementLocked(ele *list.Element) {
	c.ll.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.cache, kv.key)
	c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
}
