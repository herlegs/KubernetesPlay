package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	machineIP string
	getIPOnce sync.Once
)

func main() {
	fmt.Printf("starting hello server at [%v]...\n", getIPAddr())
	http.HandleFunc("/", hello)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	message := fmt.Sprintf("Hello %v, I'm %v", path, getIPAddr())
	w.Write([]byte(message))
}

func getIPAddr() string {
	getIPOnce.Do(func() {
		ifaces, err := net.Interfaces()
		if err != nil {
			machineIP = randomString()
			return
		}
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				machineIP = randomString()
				return
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if !ip.IsLoopback() && ip.To4() != nil {
					machineIP = ip.String()
					return
				}
			}
		}
	})
	return machineIP
}

func randomString() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 5)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
