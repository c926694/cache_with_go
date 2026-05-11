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
	addr string
}

func NewServer(capacity int,addr string) *Server {
	return &Server{
		cache: NewCache(int64(capacity)),
		addr: addr,
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
	l,err:=net.Listen("tcp",s.addr)
	if err!=nil{
		log.Fatalf("listen addr %s  failed",s.addr)
	}
	server:=grpc.NewServer()
	pb.RegisterCacheServer(server,s)
	return server.Serve(l)
}
