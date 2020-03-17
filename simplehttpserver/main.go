package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/herlegs/KubernetesPlay/serverutil"

	"golang.org/x/net/context"

	"github.com/herlegs/KubernetesPlay/simplegrpcserver/pb"
	"google.golang.org/grpc"
)

var version = time.Now().In(time.FixedZone("GMT", 8*60*60)).Format("2006/01/02 15:04")

var podCount = 1

func main() {
	initKubeClient()
	fmt.Printf("starting hello server [%v] at [%v]...\n", version, serverutil.GetIPAddr())
	http.HandleFunc("/", hello)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

//initKubeClient
func initKubeClient() {
	fmt.Printf("trying to init kube client...\n")
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error getting kubernetes cluster config: %v\n", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error init kubernetes clientset: %v\n", err)
	}
	go func() {
		//watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", "test", fields.Everything())
		//
		//_, controller := cache.NewInformer(
		//	watchlist,
		//	&v1.Pod{},
		//	time.Second*0,
		//	cache.ResourceEventHandlerFuncs{
		//		AddFunc: func(obj interface{}) {
		//			fmt.Printf("add: %s \n", obj)
		//		},
		//		DeleteFunc: func(obj interface{}) {
		//			fmt.Printf("delete: %s \n", obj)
		//		},
		//		UpdateFunc: func(oldObj, newObj interface{}) {
		//			fmt.Printf("old: %s, new: %s \n", oldObj, newObj)
		//		},
		//	},
		//)
		//stop := make(chan struct{})
		//go controller.Run(stop)

		for {
			ticker := time.NewTicker(time.Second * 5)
			defer ticker.Stop()

			for range ticker.C {
				l, err := clientset.CoreV1().Pods("test").List(metav1.ListOptions{
					LabelSelector: "app=helloapp",
				})
				if err != nil {
					fmt.Printf("error calling api: %v\n", err)
				}

				fmt.Printf("try listing result...\n")
				if l != nil {
					for _, n := range l.Items {
						fmt.Printf("items: %v\n", n.Name)
					}
				}
			}
		}

	}()
}

func hello(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	message := fmt.Sprintf("version [%v] Hello %v, I'm %v.\n%v", version, path, serverutil.GetIPAddr(), callCounterServerViaK8S(path))
	w.Write([]byte(message))
}

func callCounterServerViaK8S(input string) string {
	counterServerAddr := fmt.Sprintf("%v:%v", os.Getenv("BACKENDSERVICE_SERVICE_HOST"), os.Getenv("BACKENDSERVICE_SERVICE_PORT_GRPC_PORT"))
	failMsgTmpl := "Failed to [%v] counter server[%v]: %v"
	conn, err := grpc.Dial(counterServerAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Sprintf(failMsgTmpl, "Dail", counterServerAddr, err)
	}
	client := pb.NewCounterClient(conn)
	resp, err := client.Count(context.Background(), &pb.CountRequest{Message: input})
	if err != nil {
		return fmt.Sprintf(failMsgTmpl, "Call", counterServerAddr, err)
	}
	return fmt.Sprintf("Counter server[%v] says: %v", resp.Address, resp.Message)
}
