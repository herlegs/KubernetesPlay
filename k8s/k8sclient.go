package k8s

import (
	"encoding/json"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//InitDummyKubeClient ...
func InitDummyKubeClient() {
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
			ticker := time.NewTicker(time.Second * 30)
			defer ticker.Stop()

			for range ticker.C {
				l, err := clientset.CoreV1().Pods("sprinkler8").List(metav1.ListOptions{
					LabelSelector: "app=helloapp",
				})
				if err != nil {
					fmt.Printf("error calling api: %v\n", err)
				}

				fmt.Printf("try listing result...\n")
				if l != nil {
					for i, n := range l.Items {
						if i <= 10 {
							fmt.Printf("items: %v\n", n.Name)
						}
					}
				}
			}
		}

	}()
}

// PodMetricsList ...
type PodMetricsList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink string `json:"selfLink"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			Namespace         string    `json:"namespace"`
			SelfLink          string    `json:"selfLink"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
		} `json:"metadata"`
		Timestamp  time.Time `json:"timestamp"`
		Window     string    `json:"window"`
		Containers []struct {
			Name  string `json:"name"`
			Usage struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"usage"`
		} `json:"containers"`
	} `json:"items"`
}

// InitResourceMonitor
func InitResourceMonitor()  {
	fmt.Printf("trying to init kube client...\n")
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error getting kubernetes cluster config: %v\n", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error init kubernetes clientset: %v\n", err)
	}

	metrics := &PodMetricsList{}
	go func(){
		for {
			ticker := time.NewTicker(time.Second * 60)
			defer ticker.Stop()

			for range ticker.C {
				data, err := clientset.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/namespaces/sprinkler8/pods").DoRaw()
				if err != nil {
					fmt.Printf("error getting pods resources: %v\n",err)
				} else {
					err = json.Unmarshal(data, metrics)
					if err != nil {
						fmt.Printf("error parsing pod metric list: %v\n",err)
					} else {
						for i := 0; i < 10 && i < len(metrics.Items); i++ {
							pod := metrics.Items[i]
							fmt.Printf("pod name: %v, containers:\n",pod.Metadata.Name)
							for _, c := range pod.Containers {
								fmt.Printf("name: %v, usage: %#v\n",c.Name, c.Usage)
							}
						}
					}
				}
			}
		}
	}()
}

// NewInClusterClient ...
func NewInClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
