package metrics

import (
	"fmt"
	"net"
)

var (
	allocated = make(map[int]bool)
)

func getRandomPort(host, network string, start, end int) (int, error) {
	for si := start; si <= end; si++ {
		if allocated[si] {
			continue
		}
		addr := fmt.Sprintf("%s:%d", host, si)
		if portIsAvailable(network, addr) {
			allocated[si] = true
			return si, nil
		}
	}
	return 0, fmt.Errorf("no available port")
}

func portIsAvailable(network, addr string) bool {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
