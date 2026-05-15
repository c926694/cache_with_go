package cache

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultEtcdDialTimeout = 3 * time.Second
	defaultEtcdLeaseTTL    = 10
)

type EtcdRegistry struct {
	client *clientv3.Client
	lease  clientv3.LeaseID
	key    string
	cancel context.CancelFunc
}

func RegisterEtcd(ctx context.Context, endpoints []string, serviceName, addr string) (*EtcdRegistry, error) {
	endpoints = normalizeEtcdEndpoints(endpoints)
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints are empty")
	}
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, fmt.Errorf("cache advertise address is empty")
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: defaultEtcdDialTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %v", err)
	}

	grant, err := cli.Grant(ctx, defaultEtcdLeaseTTL)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("failed to grant etcd lease: %v", err)
	}

	key := etcdNodeKey(serviceName, addr)
	if _, err = cli.Put(ctx, key, addr, clientv3.WithLease(grant.ID)); err != nil {
		cli.Close()
		return nil, fmt.Errorf("failed to register cache node to etcd: %v", err)
	}

	keepAliveCtx, cancel := context.WithCancel(context.Background())
	ch, err := cli.KeepAlive(keepAliveCtx, grant.ID)
	if err != nil {
		cancel()
		cli.Close()
		return nil, fmt.Errorf("failed to keep etcd lease alive: %v", err)
	}

	registry := &EtcdRegistry{
		client: cli,
		lease:  grant.ID,
		key:    key,
		cancel: cancel,
	}
	go registry.consumeKeepAlive(ch)
	return registry, nil
}

func (r *EtcdRegistry) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	if r.cancel != nil {
		r.cancel()
	}

	var err error
	if r.client != nil {
		if r.key != "" {
			if _, deleteErr := r.client.Delete(ctx, r.key); deleteErr != nil {
				err = deleteErr
			}
		}
		if r.lease != 0 {
			if _, revokeErr := r.client.Revoke(ctx, r.lease); revokeErr != nil && err == nil {
				err = revokeErr
			}
		}
		if closeErr := r.client.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

func (r *EtcdRegistry) consumeKeepAlive(ch <-chan *clientv3.LeaseKeepAliveResponse) {
	for res := range ch {
		if res == nil {
			log.Printf("etcd keepalive stopped for %s", r.key)
			return
		}
	}
}

func NewEtcdClusterClient(ctx context.Context, replicas uint, endpoints []string, serviceName string) (*ClusterClient, error) {
	endpoints = normalizeEtcdEndpoints(endpoints)
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints are empty")
	}

	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: defaultEtcdDialTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %v", err)
	}

	clusterClient := newEmptyClusterClient(replicas)
	if err := updateClusterNodesFromEtcd(ctx, etcdCli, clusterClient, serviceName); err != nil {
		etcdCli.Close()
		return nil, err
	}

	watchCtx, cancel := context.WithCancel(ctx)
	clusterClient.closeFn = func() error {
		cancel()
		return etcdCli.Close()
	}
	go watchEtcdNodes(watchCtx, etcdCli, clusterClient, serviceName)
	return clusterClient, nil
}

func updateClusterNodesFromEtcd(ctx context.Context, etcdCli *clientv3.Client, clusterClient *ClusterClient, serviceName string) error {
	resp, err := etcdCli.Get(ctx, etcdPrefix(serviceName), clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to discover cache nodes from etcd: %v", err)
	}

	addrs := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}
	return clusterClient.UpdateNodes(addrs)
}

func watchEtcdNodes(ctx context.Context, etcdCli *clientv3.Client, clusterClient *ClusterClient, serviceName string) {
	watchCh := etcdCli.Watch(ctx, etcdPrefix(serviceName), clientv3.WithPrefix())
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-watchCh:
			if !ok {
				return
			}
			if resp.Err() != nil {
				log.Printf("failed to watch cache nodes from etcd: %v", resp.Err())
				continue
			}
			if err := updateClusterNodesFromEtcd(ctx, etcdCli, clusterClient, serviceName); err != nil {
				log.Printf("failed to update cache nodes from etcd: %v", err)
			}
		}
	}
}

func normalizeEtcdEndpoints(endpoints []string) []string {
	normalized := make([]string, 0, len(endpoints))
	seen := make(map[string]struct{})
	for _, endpoint := range endpoints {
		endpoint = strings.TrimSpace(endpoint)
		if endpoint == "" {
			continue
		}
		if _, ok := seen[endpoint]; ok {
			continue
		}
		seen[endpoint] = struct{}{}
		normalized = append(normalized, endpoint)
	}
	return normalized
}

func etcdPrefix(serviceName string) string {
	serviceName = strings.Trim(strings.TrimSpace(serviceName), "/")
	if serviceName == "" {
		serviceName = "cache"
	}
	return path.Join("/", serviceName, "nodes") + "/"
}

func etcdNodeKey(serviceName, addr string) string {
	return etcdPrefix(serviceName) + strings.ReplaceAll(addr, "/", "_")
}
