package cache

import (
	"cache/pb"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedCacheServer
	cache *Cache
}

func NewServer(capacity int) *Server {
	return &Server{
		cache: NewCache(int64(capacity)),
	}
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	// TODO: implement the Set method
	value:=NewByteView(req.Value)
	s.cache.Set(req.Key,value)
	return &pb.SetResponse{}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	// TODO: implement the Get method
	value,ok:=s.cache.Get(req.Key)
	if !ok{
		return nil, fmt.Errorf("key not found")
	}
	return &pb.GetResponse{
		Value: value.ByteSlice(),
	}, nil
}
func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	// TODO: implement the Delete method
	ok:=s.cache.Delete(req.Key)
	if !ok{
		return nil, fmt.Errorf("key not found")
	}
	return &pb.DeleteResponse{}, nil
}

func (s *Server) SetWithExpiration(ctx context.Context, req *pb.SetWithExpirationRequest) (*pb.SetWithExpirationResponse, error) {
	// TODO: implement the SetWithExpiration method
	value:=NewByteView(req.Value)
	s.cache.SetWithExpiration(req.Key,value,time.Duration(req.Expiration))
	return &pb.SetWithExpirationResponse{}, nil
}

func (s *Server) Start() error {
	// TODO: implement the Start method
	l,err:=net.Listen("tcp","127.0.0.1:8080")
	if err!=nil{
		log.Fatal("listen 8080 failed")
	}
	server:=grpc.NewServer()
	pb.RegisterCacheServer(server,s)
	return server.Serve(l)
}

