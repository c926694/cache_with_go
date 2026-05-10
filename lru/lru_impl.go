package lru

import "container/list"

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
}

func NewLRU(maxBytes int64) *LRUCacheImpl {
	return &LRUCacheImpl{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

func (c *LRUCacheImpl) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		//存在,更新lru顺序
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	//不存在,返回false
	return nil, false
}

func (c *LRUCacheImpl) Set(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		//存在,更新lru顺序
		kv := ele.Value.(*entry)
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		c.ll.MoveToFront(ele)
	} else {
		//不存在,添加到lru队列
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}
	//检查是否超过最大缓存大小
	for c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *LRUCacheImpl) Delete(key string) bool {
	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		return true
	}
	return false
}

func (c *LRUCacheImpl) RemoveOldest() {
	ele := c.ll.Back()
	if ele == nil {
		return
	}

	c.ll.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.cache, kv.key)
	c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
}

func (c *LRUCacheImpl) UsedBytes() int64 {
	return c.usedBytes
}
