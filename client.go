package cache

import (
	"cache/pb"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn      *grpc.ClientConn
	grpcCli   pb.CacheClient
	healthCli grpc_health_v1.HealthClient
}




func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %v", err)
	}
	grpcCli := pb.NewCacheClient(conn)
	healthCli := grpc_health_v1.NewHealthClient(conn)
	return &Client{conn: conn, grpcCli: grpcCli, healthCli: healthCli}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Set(ctx context.Context, key string, value []byte) error {
	return c.SetWithExpiration(ctx, key, value, 0)
}

func (c *Client) Ping(ctx context.Context) error {
	res, err := c.healthCli.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return fmt.Errorf("failed to ping cache server: %v", err)
	}
	if res.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("cache server status is %s", res.Status.String())
	}
	return nil
}

func (c *Client) SetWithExpiration(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	req := &pb.SetWithExpirationRequest{
		Key:        key,
		Value:      value,
		Expiration: int64(expiration),
	}
	res, err := c.grpcCli.SetWithExpiration(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to set value with expiration to cache: %v", err)
	}
	log.Printf("grpc set with expiration req:%v res:%v", req, res)
	return nil
}

const getterCacheExpiration = 1 * time.Minute

func (c *Client) Get(ctx context.Context, key string, getter Getter,args ...any) ([]byte, error) {
	req := &pb.GetRequest{
		Key: key,
	}
	res, err := c.grpcCli.Get(ctx, req)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return nil, fmt.Errorf("failed to get value from cache: %v", err)
		}
		//key不存在，则调用getter获取值
		data,err:=getter(args...)
		if err != nil {
			return nil, fmt.Errorf("failed to get value from getter: %v", err)
		}
		//序列化值
		value, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value: %v", err)
		}
		//存入缓存
		if err=c.SetWithExpiration(ctx,key,value,getterCacheExpiration); err!=nil{
			return nil, fmt.Errorf("failed to set value with expiration to cache: %v", err)
		}
		return value, nil
	}
	log.Printf("grpc get req:%v res:%v", req, res)
	return res.Value, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	req := &pb.DeleteRequest{
		Key: key,
	}
	res, err := c.grpcCli.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete value from cache: %v", err)
	}
	log.Printf("grpc delete req:%v res:%v", req, res)
	return nil
}
