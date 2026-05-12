package cache

import (
	"cache/consistenthash"
	"context"
	"fmt"
	"strings"
	"time"
)

type ClusterClient struct {
	hash    *consistenthash.Map
	clients map[string]*Client
}

func NewClusterClient(replicas uint, addrs ...string) (*ClusterClient, error) {
	uniqueAddrs, err := validNodeAddrs(addrs)
	if err != nil {
		return nil, err
	}

	hash := consistenthash.New(int(replicas), nil)
	clusterClient := &ClusterClient{
		hash:    hash,
		clients: make(map[string]*Client),
	}
	for _, addr := range uniqueAddrs {
		cli, err := NewClient(addr)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = cli.Ping(ctx)
		cancel()
		if err != nil {
			cli.Close()
			return nil, fmt.Errorf("cache node %s is unavailable: %v", addr, err)
		}
		clusterClient.clients[addr] = cli
	}
	hash.Add(uniqueAddrs...)
	return clusterClient, nil
}

func validNodeAddrs(addrs []string) ([]string, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("cache node address is empty")
	}

	seen := make(map[string]struct{})
	uniqueAddrs := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			return nil, fmt.Errorf("cache node address is empty")
		}
		if _, ok := seen[addr]; ok {
			continue
		}

		seen[addr] = struct{}{}
		uniqueAddrs = append(uniqueAddrs, addr)
	}

	return uniqueAddrs, nil
}

func (c *ClusterClient) SetWithExpiration(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if expiration < 0 {
		return fmt.Errorf("expiration must be non-negative")
	}
	if value == nil {
		return fmt.Errorf("value is nil")
	}
	//获取Key所在结点
	node := c.getNode(key)
	return node.SetWithExpiration(ctx, key, value, expiration)
}

func (c *ClusterClient) Delete(ctx context.Context, key string) error {
	node := c.getNode(key)
	return node.Delete(ctx, key)
}

func (c *ClusterClient) Get(ctx context.Context, key string) ([]byte, error) {
	node := c.getNode(key)
	return node.Get(ctx, key)
}

func (c *ClusterClient) Set(ctx context.Context, key string, value []byte) error {
	return c.SetWithExpiration(ctx, key, value, 0)
}

func (c *ClusterClient) getNode(key string) *Client {
	node := c.hash.Get(key)
	return c.clients[node]
}
