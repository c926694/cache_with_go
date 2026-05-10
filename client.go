package cache

import (
	"cache/pb"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	grpcCli pb.CacheClient
}

func NewClient(addr string) (*Client, error) { 
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %v", err)
	}
	grpcCli := pb.NewCacheClient(conn)
	return &Client{conn: conn, grpcCli: grpcCli}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Set(ctx context.Context, key string, value []byte) error {
	req := &pb.SetRequest{
		Key:   key,
		Value: value,
	}
	res,err:=c.grpcCli.Set(ctx,req)
	if err != nil {
		return fmt.Errorf("failed to set value to cache: %v", err)
	}
	log.Printf("grpc set req:%v res:%v",req,res)
	return nil
}
func (c *Client) SetWithExpiration(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	req := &pb.SetWithExpirationRequest{
		Key:   key,
		Value: value,
		Expiration: int64(expiration),
	}
	res,err:=c.grpcCli.SetWithExpiration(ctx,req)
	if err != nil {
		return fmt.Errorf("failed to set value with expiration to cache: %v", err)
	}
	log.Printf("grpc set with expiration req:%v res:%v",req,res)
	return nil
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	req := &pb.GetRequest{
		Key: key,
	}
	res,err:=c.grpcCli.Get(ctx,req)
	if err != nil {
		return nil, fmt.Errorf("failed to get value from cache: %v", err)
	}
	log.Printf("grpc get req:%v res:%v",req,res)
	return res.Value,nil
}

func (c *Client) Delete(ctx context.Context, key string) error { 
	req := &pb.DeleteRequest{
		Key: key,
	}
	res,err:=c.grpcCli.Delete(ctx,req)
	if err != nil {
		return fmt.Errorf("failed to delete value from cache: %v", err)
	}
	log.Printf("grpc delete req:%v res:%v",req,res)
	return nil
}