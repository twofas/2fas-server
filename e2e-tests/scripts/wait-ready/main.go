package main

import (
	"flag"
	"log"
	"net"
	"strings"
	"time"
)

func main() {
	addrFlag := flag.String("addr", ":80;:8081;:8082", "list of addresses to check sep by ;")
	flag.Parse()

	addresses := strings.Split(*addrFlag, ";")
	if len(addresses) < 1 {
		log.Fatal("-addr value not provided")
	}
	for _, address := range addresses {
		running := waitForApp(address, 30*time.Second)
		if !running {
			log.Fatal("App not running on addr: ", address)
		}
	}
}

// waitForApp returns true if app is listening on provided address.
// If it cannot connect up to specified timeout, it returns false.
func waitForApp(address string, timeout time.Duration) bool {
	done := make(chan struct{})

	go func() {
		for {
			_, err := net.DialTimeout("tcp", address, time.Second)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			close(done)
			return
		}
	}()
	timeoutCh := time.After(timeout)
	select {
	case <-done:
		return true
	case <-timeoutCh:
		return false
	}
}
