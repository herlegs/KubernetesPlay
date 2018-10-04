#!/usr/bin/env bash

kubectl create -f simple-pod

echo "try: curl 127.0.0.1:9090/test in another terminal"

kubectl port-forward ktest 9090:8080

