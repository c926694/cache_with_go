package lru

import "time"

type LRUCache interface {
	Set(key string, value Value)
	Get(key string) (Value, bool)
	Delete(key string) bool
	SetWithExpiration(key string, value Value, expiration time.Duration)
	UsedBytes() int64
}