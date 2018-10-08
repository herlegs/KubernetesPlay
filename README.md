This project is to explore a kubernetes + envoy/istio integration. In the first part we will explore the basic infrastructure
of a kubernetes cluster.

####Micro-service in Kubernetes

#####Tools prerequisites:

1.  docker must be installed  
    https://docs.docker.com/
2. virtual box must be installed (for minikube)  
    https://www.virtualbox.org/wiki/Downloads
3. minikube must be installed  
    __brew cask install minikube__ or https://kubernetes.io/docs/tasks/tools/install-minikube/
4. kubectl must be installed  
    __brew install kubernetes-cli__ or https://kubernetes.io/docs/tasks/tools/install-kubectl/

#####Steps

#####1). Service deployment with replicas

this will create a vm running kubernetes master
>__minikube start__

// use this to view the vm ip address (ssh username: docker, password: tcuser)
>__kubectl cluster-info__ (or "__minikube ip__")

We will write several simple go server programs to simulate our micro-service.  
In ./siplehttpserver we created an go server will reply "Hello {/path}, I'm {IP address}".

(_You don't have to build docker image here, you can just reuse the image built in my hub stardust1991_)  
 _Or you can just run "cd ./docker && ./buildSimpleServer.sh"_  

>CGO_ENABLED=0 GOOS=linux go build -a -o main simplehttpserver  
>_If you encountered "golang.org/x/sys/unix not found", run "go get golang.org/x/sys/unix"_  


>// tag and push the image to docker hub  
>docker build -t stardust1991/hellomain -f SimpleDockerfile .  
>docker push stardust1991/hellomain  
>rm main  

From here we will need some kubernetes concepts  
Reading Prerequisites:
1. Understand Pods in Kubernetes  
    https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/  
2. Understand Controllers (ReplicaSet, Deployment in our case)  
    https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/  
    https://kubernetes.io/docs/concepts/workloads/controllers/deployment/  
3. Understand Services  
    https://kubernetes.io/docs/concepts/services-networking/service/  

In ./kube directory, we've created several kubernetes configuration files.  
We can use simple-deployment.yaml to create a deployment, this will bring up 3 hello server instances.  
>kubectl create -f simple-deployment --save-config

run this command to get information about Pods, Deployments...  
>"kubectl get all -o wide"

After deployment finished, let's add our server instances behind an abstraction called Service.  
>kubectl create -f simple-service --save-config

The service will be bound to Node port 30001 in the configuration.
So we can access our go server endpoint in several ways:  
1. use the ip address get from __kubectl cluster-info__, say 192.168.99.101,
    run "curl 192.168.99.101:30001/anything"
2. ssh into 192.168.99.101 (use docker:tcuser), get helloservice cluster-ip address from __kubectl get all -o wide__, say 10.107.114.97,
    run __curl 10.107.114.97:7777/anything__
3. select any of your pod (or helloservice), say "hellodeployment-7475db8bd7-2kb46", run __kubectl port-forward hellodeployment-7475db8bd7-2kb46 8888:8080__  
    (if it's service then run __kubectl port-forward service/helloservice 8888:7777__)
    and in another terminal run __curl 127.0.0.1:8888/anything__

We will observe the requests are load balanced from the IP address in server's response.

#####2). Service communication
Let's explore how the services communication happens in kubernetes . For this purpose we will host another go grpc server, which has an API to count bytes in string.
Implementation is under ./simplegrpcserver.  
Our previous simple http server will call the grpc server, just to count the byte of the url path

In ./docker folder, we build and push the grpc server image similarly.

we will deploy the counter service the same way (here we name it as backenddeployment)
>kubectl create -f backend-deployment --save-config (under folder ./kube)

and create service for the counter server:
>kubectl create -f backend-service --save-config

in the simple service program, the server will get the backendservice address from the environment variable:  
BACKENDSERVICE_SERVICE_HOST and BACKENDSERVICE_SERVICE_PORT_GRPC_PORT  
and kubernetes will do the routing to the counter server.  
when we query the simple server, the response will be:  
>➜  kube git:(master) ✗ curl 192.168.99.101:30001/myname  
>Hello myname, I'm 172.17.0.9. Counter server says: 6  





