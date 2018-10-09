package server

import (
	"fmt"

	"github.com/herlegs/KubernetesPlay/serverutil"

	"github.com/herlegs/KubernetesPlay/simplegrpcserver/pb"
	"golang.org/x/net/context"
)

type CounterServer struct {
}

func (s *CounterServer) Count(ctx context.Context, in *pb.CountRequest) (*pb.CountResponse, error) {
	return &pb.CountResponse{
		Address: serverutil.GetIPAddr(),
		Message: fmt.Sprintf("Input has %v bytes.", len(in.Message)),
	}, nil
}
