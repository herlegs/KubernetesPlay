package serverutil

import (
	"math/rand"
	"net"
	"sync"
	"time"
)

var (
	machineIP string
	getIPOnce sync.Once
)

func GetIPAddr() string {
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
