#!/bin/bash

CGO_ENABLED=0 GOOS=linux go build -a -o main ../simplehttpserver

docker build -t stardust1991/hellomain -f SimpleDockerfile .

rm main

docker push stardust1991/hellomain