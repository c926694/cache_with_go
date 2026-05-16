# Distributed Cache

基于 Go 实现的分布式缓存系统，支持 gRPC 远程访问、LRU 容量淘汰、TTL 过期、一致性哈希分片、singleflight 防击穿，以及基于 etcd 的服务注册与发现。

## Features

- gRPC 接口：支持 `Get`、`Set`、`Delete`、`SetWithExpiration`。
- LRU 淘汰：按容量限制缓存占用，空间不足时淘汰最近最少使用的数据。
- TTL 过期：支持为缓存 key 设置过期时间。
- 一致性哈希：客户端根据 key 路由到固定缓存节点，减少扩缩容时的数据迁移。
- singleflight：合并同一个 key 的并发缓存 miss 请求，降低重复回源压力。
- etcd 服务发现：缓存节点通过 lease/keepalive 注册，客户端 watch 节点变化并动态更新哈希环。
- Docker 部署：支持镜像构建和参数化启动。

## Architecture

```text
Business Service
      |
      | NewEtcdClusterClient
      v
    etcd  <---------------- cache-server register/keepalive
      |
      | discover / watch
      v
ClusterClient
      |
      | consistent hash by key
      v
cache-server-1 / cache-server-2 / cache-server-N
```

etcd 只负责服务注册与发现，不保存缓存数据。缓存数据仍然存储在各个 cache-server 节点内存中。

## Project Layout

```text
.
├── cmd/                  # cache-server 启动入口
├── consistenthash/       # 一致性哈希实现
├── lru/                  # LRU 缓存实现
├── pb/                   # protobuf 与 gRPC 生成代码
├── test/                 # 手动测试和压测示例
├── cache.go              # 缓存封装
├── client.go             # 单节点客户端
├── cluster_client.go     # 多节点客户端与一致性哈希路由
├── etcd.go               # etcd 注册发现
├── server.go             # gRPC 服务端
└── Dockerfile
```

## Quick Start

### Start cache-server without etcd

```bash
go run ./cmd -addr 127.0.0.1:3000
```

静态客户端连接：

```go
cli, err := cache.NewClusterClient(60, "127.0.0.1:3000")
if err != nil {
	panic(err)
}
defer cli.Close()
```

### Start with etcd

先启动 etcd，确保客户端地址为 `127.0.0.1:2379`。

启动 cache-server 并注册到 etcd：

```bash
go run ./cmd \
  -addr 127.0.0.1:3000 \
  -advertise-addr 127.0.0.1:3000 \
  -etcd-endpoints 127.0.0.1:2379 \
  -service-name cache
```

检查节点注册：

```bash
etcdctl --endpoints=127.0.0.1:2379 get /cache/nodes/ --prefix
```

客户端通过 etcd 发现节点：

```go
cli, err := cache.NewEtcdClusterClient(
	context.Background(),
	60,
	[]string{"127.0.0.1:2379"},
	"cache",
)
if err != nil {
	panic(err)
}
defer cli.Close()
```

## Client Example

```go
ctx := context.Background()

getter := func(args ...any) (any, error) {
	id := args[0].(int)
	return fmt.Sprintf("user:%d", id), nil
}

if err := cli.Set(ctx, "k1", []byte("v1")); err != nil {
	panic(err)
}

value, err := cli.Get(ctx, "user:1", getter, time.Minute, 1)
if err != nil {
	panic(err)
}

fmt.Println(string(value))
```

`Get` 在缓存 miss 时会调用 `getter` 回源，并把返回结果写回缓存。若只想读取已存在 key，需要确保 key 已经预热，否则 miss 会返回错误或触发回源逻辑。

## Docker

Build image:

```bash
docker build -t cache-server .
```

Run a single node:

```bash
docker run --rm -p 3000:3000 cache-server \
  -addr 0.0.0.0:3000
```

Run with etcd on host:

```bash
docker run -d --name cache-3000 -p 3000:3000 cache-server \
  -addr 0.0.0.0:3000 \
  -advertise-addr 127.0.0.1:3000 \
  -etcd-endpoints host.docker.internal:2379 \
  -service-name cache
```

Linux Docker 如果无法解析 `host.docker.internal`，可以改用宿主机实际 IP，或者把 etcd 和 cache-server 放到同一个 Docker network 中。

启动第二个节点：

```bash
docker run -d --name cache-3001 -p 3001:3000 cache-server \
  -addr 0.0.0.0:3000 \
  -advertise-addr 127.0.0.1:3001 \
  -etcd-endpoints host.docker.internal:2379 \
  -service-name cache
```

## Command Options

```text
-addr             cache-server 监听地址，默认 0.0.0.0:3000
-capacity         缓存容量，单位 byte，默认 67108864
-advertise-addr   注册到 etcd 的可访问地址
-etcd-endpoints   etcd 地址，多个地址用逗号分隔
-service-name     etcd 服务名前缀，默认 cache
```

注意：`-addr` 是服务端监听地址，Docker 中通常写 `0.0.0.0:3000`；`-advertise-addr` 是客户端实际连接地址，不建议注册 `0.0.0.0:3000`。

## Benchmark

压测建议分场景进行：

- 纯 Set：测试写入吞吐。
- 纯 Get 命中：先预热 key，再并发读取。
- Get miss + getter：测试回源与 singleflight。
- 混合读写：例如 90% Get + 10% Set。
- 多节点：验证一致性哈希路由和吞吐变化。
- 节点下线：停止某个 cache-server，观察 etcd 节点删除和客户端路由更新。

本地压测示例结果：

```text
本地压测:1000 并发压测 100 万次请求，QPS 约 5.3w
```

实际结果会受机器配置、节点数量、日志输出、key 分布和缓存容量影响。

## Development

Run tests:

```bash
go test ./...
```

Format code:

```bash
gofmt -w .
```

## Notes

- 当前缓存数据存储在内存中，节点重启后缓存会丢失。
- etcd 用于节点注册发现，不承担缓存数据持久化。
- 如果需要持久化，可以优先实现快照恢复，用于降低缓存节点重启后的冷启动压力。
