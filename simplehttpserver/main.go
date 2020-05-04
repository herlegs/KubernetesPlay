package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/herlegs/KubernetesPlay/simplehttpserver/monitor"

	"github.com/herlegs/KubernetesPlay/serverutil"

	"golang.org/x/net/context"

	"github.com/herlegs/KubernetesPlay/simplegrpcserver/pb"
	"google.golang.org/grpc"
)

var version = time.Now().In(time.FixedZone("GMT", 8*60*60)).Format("2006/01/02 15:04")

var podCount = 1

func main() {
	fmt.Printf("before init...\n")
	monitor.InitResouceMonitor("coban-prd8")
	fmt.Printf("starting hello server [%v] at [%v]...\n", version, serverutil.GetIPAddr())
	http.HandleFunc("/", hello)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
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
