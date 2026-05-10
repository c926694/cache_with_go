package lru

import (
	"container/list"
	"sync"
	"time"
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
	expires  map[string]time.Time
	mu        sync.Mutex
}

func NewLRU(maxBytes int64) *LRUCacheImpl {
	return &LRUCacheImpl{
		maxBytes: maxBytes,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
		expires: make(map[string]time.Time),
	}
}

func (c *LRUCacheImpl) Get(key string) (Value, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ele, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	expireTime:=c.expires[key]
	//判断有没有超过过期时间
	currentTime:=time.Now()
	if expireTime.IsZero() || currentTime.Before(expireTime) {
		//没超时或者没有设置过期时间,更新最近使用时间并返回缓存
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return nil, false
}

func (c *LRUCacheImpl) Set(key string, value Value) {
	c.SetWithExpiration(key,value,0)
}

func (c *LRUCacheImpl) checkMemory() {
	for c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.removeOldestLocked()
	}
}

func (c *LRUCacheImpl) SetWithExpiration(key string, value Value, expiration time.Duration) {
	// Implementation for setting key-value with expiration
	if expiration<0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	expTime:=time.Now().Add(expiration)
	if ele, ok := c.cache[key]; ok {
		//key存在,更新value和expiration时间
		kv := ele.Value.(*entry)
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		c.ll.MoveToFront(ele)
		c.expires[key] = expTime
	} else {
		//key不存在,新增key-value和expiration时间
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.usedBytes += int64(len(key)) + int64(value.Len())
		c.expires[key] = expTime
	}
	//检查内存是否超出限制,如果超出则删除最近最少使用的key
	c.checkMemory()
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
