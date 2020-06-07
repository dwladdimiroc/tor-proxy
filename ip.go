package main

import (
	"fmt"
	"sync"
)

var (
	mutexIp sync.Mutex
	ips     map[string]int
)

func checkIp(ipCurrent string, server int) error {
	if exist := existsIp(ipCurrent); !exist {
		if err := addIp(ipCurrent, server); err != nil {
			return err
		}
		fmt.Printf("ips lenght %d\n", len(ips))
	} else {
		fmt.Printf("the ip %v got for server %d exists by server %d\n", ipCurrent, server, ips[ipCurrent])
	}

	return nil
}

func existsIp(ipCurrent string) bool {
	for ip, _ := range ips {
		if ip == ipCurrent {
			return true
		}
	}
	return false
}

func addIp(ip string, server int) error {
	ips[ip] = server
	fmt.Printf("add ip %s for server %d\n", ip, server)
	if err := banIpAddress(ip); err != nil {
		return err
	} else {
		return nil
	}
}
