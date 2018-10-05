#!/bin/bash

CGO_ENABLED=0 GOOS=linux go build -a -o main ../simplegrpcserver

docker build -t stardust1991/grpcserver -f GrpcDockerfile .

rm main

docker push stardust1991/grpcserver