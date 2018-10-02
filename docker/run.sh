#!/bin/bash

CGO_ENABLED=0 GOOS=linux go build -a -o main ../simplehttpserver

docker build -t hellomain -f Dockerfile .

rm main

docker run -it -p 80:8080 hellomain