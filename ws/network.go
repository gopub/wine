package ws

import (
	"net"
	"time"
)

type NetworkMonitor struct {
	C       <-chan struct{}
	c       chan<- struct{}
	stopped bool
}

func NewNetworkMonitor() *NetworkMonitor {
	c := make(chan struct{}, 1)
	m := &NetworkMonitor{
		C: c,
		c: c,
	}
	go m.start()
	return m
}

func (m *NetworkMonitor) start() {
	var ip net.IP
	for !m.stopped {
		newIP, _ := getOutboundIP()
		if ip != nil {
			if !ip.Equal(newIP) {
				m.c <- struct{}{}
			}
		}
		ip = newIP
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *NetworkMonitor) Stop() {
	m.stopped = true
}

func getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	addr := conn.LocalAddr().(*net.UDPAddr)
	if err = conn.Close(); err != nil {
		return nil, err
	}
	return addr.IP, nil
}
