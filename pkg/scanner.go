package pkg

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// PortScanner : struct for port scanning
type PortScanner struct {
	host    string
	timeout time.Duration
	threads int
}

// NewPortScanner : returns new PortScanner struct
func NewPortScanner(host string, timeout time.Duration, threads int) *PortScanner {
	return &PortScanner{host, timeout, threads}
}

// IsOpen : checks if a port is open
func (h PortScanner) IsOpen(port int) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", h.hostPort(port))
	if err != nil {
		return false
	}
	conn, err := net.DialTimeout("tcp", tcpAddr.String(), h.timeout)
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

func (h PortScanner) hostPort(port int) string {
	return fmt.Sprintf("%s:%d", h.host, port)
}

// Run : returns a list of open ports from a given list
func (h PortScanner) Run(ports []int) []int {
	rv := []int{}
	l := sync.Mutex{}
	sem := make(chan bool, h.threads)
	for _, port := range ports {
		sem <- true
		go func(port int) {
			if h.IsOpen(port) {
				l.Lock()
				rv = append(rv, port)
				l.Unlock()
			}
			<-sem
		}(port)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return rv
}
