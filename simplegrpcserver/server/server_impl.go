package server

import (
	"github.com/herlegs/KubernetesPlay/simplegrpcserver/pb"
	"golang.org/x/net/context"
)

type CounterServer struct {
}

func (s *CounterServer) Count(ctx context.Context, in *pb.CountRequest) (*pb.CountResponse, error) {
	return &pb.CountResponse{Length: int64(len(in.Message))}, nil
}
