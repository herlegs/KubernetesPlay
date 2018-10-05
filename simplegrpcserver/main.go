package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/herlegs/KubernetesPlay/simplegrpcserver/pb"

	"golang.org/x/net/context"

	"github.com/herlegs/KubernetesPlay/simplegrpcserver/server"

	"google.golang.org/grpc/reflection"

	"github.com/herlegs/KubernetesPlay/simplegrpcserver/constant"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

//go:generate protoc -I/usr/local/include -I./pb -I$PROTO_PATH/include -I. -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I$GOPATH/src --go_out=plugins=grpc:./pb --grpc-gateway_out=logtostderr=true:./pb --swagger_out=logtostderr=true:./pb ./pb/counter.proto

func main() {
	go startGRPCServer()
	startHTTPServer()
}

func startGRPCServer() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", constant.GRPCPort))
	if err != nil {
		fmt.Printf("grpc server failed to listen: %v\n", err)
		return
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCounterServer(grpcServer, &server.CounterServer{})
	reflection.Register(grpcServer)
	fmt.Printf("starting grpc server...\n")
	if err := grpcServer.Serve(listen); err != nil {
		fmt.Printf("failed to start grpc server: %v\n", err)
	}
}

func startHTTPServer() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterCounterHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%v", constant.GRPCPort), opts)
	if err != nil {
		fmt.Printf("error registering http server: %v\n", err)
		return
	}
	fmt.Printf("starting http server...\n")
	if err := http.ListenAndServe(fmt.Sprintf(":%v", constant.HTTPPort), mux); err != nil {
		fmt.Printf("failed to start http server: %v\n", err)
	}
}
