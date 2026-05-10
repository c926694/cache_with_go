package cache

import (
	"cache/pb"
	"context"
)

type Server struct {
	pb.UnimplementedKamaCacheServer
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	// TODO: implement the Set method
	return nil, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	// TODO: implement the Get method
	return nil, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	// TODO: implement the Delete method
	return nil, nil
}
