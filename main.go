package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const cantTorServer = 5

func main() {
	ips = make(map[string]int)

	var wg sync.WaitGroup
	wg.Add(cantTorServer)
	for i := 1; i <= cantTorServer; i++ {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := checkTorServer(i); err != nil {
				fmt.Println(err)
			}

			for {
				var ip string
				var err error

				tInitGetIp := time.Now()
				if ip, err = externalRawIP(i); err != nil {
					fmt.Println(err)
				} else {
					tFinishGetIp := time.Now()
					tGetIp := tFinishGetIp.Sub(tInitGetIp)
					fmt.Printf("[%v] [TorServer %d] Time Get Ip: %v\n", time.Now().Format("15:04:05"), i, tGetIp)

					mutexIp.Lock()
					if err := checkIp(ip, i) ; err != nil {
						fmt.Println(err)
					}
					if len(ips) > 1000 {
						break
					}
					mutexIp.Unlock()
				}

				if err := RestartTorServer(ip, i); err != nil {
					fmt.Println(err)
				}
			}
		}(i, &wg)
	}
	wg.Wait()
}

func externalRawIP(index int) (string, error) {
	if torTransport, err := TorTransport(index); err != nil {
		return "", nil
	} else {
		client := http.Client{
			Timeout:   1 * time.Minute,
			Transport: torTransport,
		}

		resp, err := client.Get("http://myexternalip.com/raw")
		if err != nil {
			return "", errors.New("error get: " + err.Error())
		}
		defer resp.Body.Close()

		if byteIP, err := ioutil.ReadAll(resp.Body); err != nil {
			return "", errors.New("error read: " + err.Error())
		} else {
			fmt.Println("My External IP: ", string(byteIP))
			return string(byteIP), nil
		}
	}
}
