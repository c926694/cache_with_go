package lru

type LRUCache interface {
	Set(key string, value Value)
	Get(key string) (Value, bool)
	Delete(key string) bool
	UsedBytes() int64
}