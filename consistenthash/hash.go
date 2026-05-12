package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 定义哈希函数类型。
// 参数 data 是待哈希的数据，返回值会被映射到一致性哈希环上。
type Hash func(data []byte) uint32

// Map 表示一致性哈希环。
// 每个真实节点会按照 replicas 数量生成多个虚拟节点，让 key 分布更均匀。
type Map struct {
	hash     Hash
	replicas int
	keys     []int
	hashMap  map[int]string
}

// New 创建一个一致性哈希环。
// 参数 replicas 表示每个真实节点对应的虚拟节点数量；如果 replicas <= 0，会使用默认值 50。
// 参数 fn 表示哈希函数；如果 fn 为 nil，会使用 crc32.ChecksumIEEE。
func New(replicas int, fn Hash) *Map {
	if replicas <= 0 {
		replicas = 50
	}
	if fn == nil {
		fn = crc32.ChecksumIEEE
	}

	return &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
}

// Add 添加一个或多个真实节点到哈希环。
// 参数 nodes 通常是节点地址，例如 "127.0.0.1:8081"。
func (m *Map) Add(nodes ...string) {
	for _, node := range nodes {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + "#" + node)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = node
		}
	}
	sort.Ints(m.keys)
}

// Remove 从哈希环中移除一个真实节点以及它对应的所有虚拟节点。
// 参数 node 必须和 Add 时传入的节点字符串一致；如果节点不存在，则直接忽略。
func (m *Map) Remove(node string) {
	if len(m.keys) == 0 {
		return
	}

	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + "#" + node)))
		if _, ok := m.hashMap[hash]; !ok {
			continue
		}

		delete(m.hashMap, hash)
		idx := sort.SearchInts(m.keys, hash)
		if idx < len(m.keys) && m.keys[idx] == hash {
			m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
		}
	}
}

// Get 根据 key 在哈希环上找到对应的真实节点。
// 参数 key 是缓存 key；如果哈希环为空，返回空字符串。
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	if idx == len(m.keys) {
		idx = 0
	}

	return m.hashMap[m.keys[idx]]
}
