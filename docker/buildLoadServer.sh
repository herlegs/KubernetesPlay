#!/bin/bash

CGO_ENABLED=0 GOOS=linux go build -a -o main ./loadtest

docker build -t stardust1991/loadtest -f ./docker/SimpleDockerfile .

rm main

docker push stardust1991/loadtest