package cache

import (
	"cache/consistenthash"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type ClusterClient struct {
	mu       sync.RWMutex
	replicas uint
	hash     *consistenthash.Map
	clients  map[string]*Client
	closeFn  func() error
}

type Getter func(args ...any) (any, error)

func NewClusterClient(replicas uint, addrs ...string) (*ClusterClient, error) {
	uniqueAddrs, err := validNodeAddrs(addrs)
	if err != nil {
		return nil, err
	}
	if len(uniqueAddrs) == 0 {
		return nil, fmt.Errorf("cache node address is empty")
	}

	clusterClient := newEmptyClusterClient(replicas)
	if err := clusterClient.UpdateNodes(uniqueAddrs); err != nil {
		return nil, err
	}
	return clusterClient, nil
}

func newEmptyClusterClient(replicas uint) *ClusterClient {
	return &ClusterClient{
		replicas: replicas,
		hash:     consistenthash.New(int(replicas), nil),
		clients:  make(map[string]*Client),
	}
}

func validNodeAddrs(addrs []string) ([]string, error) {
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

func (c *ClusterClient) UpdateNodes(addrs []string) error {
	uniqueAddrs, err := validNodeAddrs(addrs)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	nextClients := make(map[string]*Client, len(uniqueAddrs))
	for _, addr := range uniqueAddrs {
		if cli, ok := c.clients[addr]; ok {
			nextClients[addr] = cli
			continue
		}

		cli, err := NewClient(addr)
		if err != nil {
			closeClients(nextClients, c.clients)
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = cli.Ping(ctx)
		cancel()
		if err != nil {
			cli.Close()
			closeClients(nextClients, c.clients)
			return fmt.Errorf("cache node %s is unavailable: %v", addr, err)
		}
		nextClients[addr] = cli
	}

	nextHash := consistenthash.New(int(c.replicas), nil)
	nextHash.Add(uniqueAddrs...)
	for addr, cli := range c.clients {
		if _, ok := nextClients[addr]; !ok {
			cli.Close()
		}
	}

	c.hash = nextHash
	c.clients = nextClients
	return nil
}

func closeClients(clients map[string]*Client, keep map[string]*Client) {
	for addr, cli := range clients {
		if keep[addr] != cli {
			cli.Close()
		}
	}
}

func (c *ClusterClient) Close() error {
	c.mu.Lock()

	var closeErr error
	for _, cli := range c.clients {
		if err := cli.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	c.clients = make(map[string]*Client)
	c.hash = consistenthash.New(int(c.replicas), nil)
	closeFn := c.closeFn
	c.closeFn = nil
	c.mu.Unlock()

	if closeFn != nil {
		if err := closeFn(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

func (c *ClusterClient) SetWithExpiration(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if expiration < 0 {
		return fmt.Errorf("expiration must be non-negative")
	}
	if value == nil {
		return fmt.Errorf("value is nil")
	}
	node := c.getNode(key)
	if node == nil {
		return fmt.Errorf("cache cluster has no available nodes")
	}
	return node.SetWithExpiration(ctx, key, value, expiration)
}

func (c *ClusterClient) Delete(ctx context.Context, key string) error {
	node := c.getNode(key)
	if node == nil {
		return fmt.Errorf("cache cluster has no available nodes")
	}
	return node.Delete(ctx, key)
}

func (c *ClusterClient) Get(ctx context.Context, key string, getter Getter, cacheExpiration time.Duration, dbArgs ...any) ([]byte, error) {
	node := c.getNode(key)
	if node == nil {
		return nil, fmt.Errorf("cache cluster has no available nodes")
	}
	return node.Get(ctx, key, getter, cacheExpiration, dbArgs...)
}

func (c *ClusterClient) Set(ctx context.Context, key string, value []byte) error {
	return c.SetWithExpiration(ctx, key, value, 0)
}

func (c *ClusterClient) getNode(key string) *Client {
	c.mu.RLock()
	defer c.mu.RUnlock()

	node := c.hash.Get(key)
	if node == "" {
		return nil
	}
	return c.clients[node]
}
