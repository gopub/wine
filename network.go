package wine

import (
	"fmt"
	"net"
)

func SelectLocalPort(localIP string, minPort, maxPort int) int {
	for port := minPort; port <= maxPort; port++ {
		addr := fmt.Sprintf("%s:%d", localIP, port)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			l.Close()
			return port
		}
	}
	return 0
}

// GetOutboundIP returns preferred outbound ip of this machine
func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if err = conn.Close(); err != nil {
		return nil, err
	}
	return localAddr.IP, nil
}

// GetOutboundIPString returns preferred outbound ip of this machine
func GetOutboundIPString() string {
	ip, err := GetOutboundIP()
	if err != nil {
		logger.Errorf("GetOutboundIP: %v", err)
		return ""
	}
	return ip.String()
}

func GetMacAddresses() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	addrs := make([]string, len(ifaces))
	for i, ifa := range ifaces {
		addrs[i] = ifa.HardwareAddr.String()
	}

	return addrs, nil
}

func GetIPv4Info() (ifaceToIPList map[string][]string, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("get net interfaces: %w", err)
	}
	ifaceToIPList = make(map[string][]string)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, fmt.Errorf("get addrs %s: %w", i.Name, err)
		}

		for _, addr := range addrs {
			//var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				if ip4 := v.IP.To4(); ip4 != nil {
					ifaceToIPList[i.Name] = append(ifaceToIPList[i.Name], ip4.String())
				}
			case *net.IPAddr:
				//	if ip4 := v.IP.To4(); ip4 != nil {
				//		res[i.Name] = append(res[i.Name], ip4.String())
				//	}
			}
		}
	}
	return ifaceToIPList, nil
}

func GetIPv6Info() (ifaceToIPList map[string][]string, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("get net interfaces: %w", err)
	}
	ifaceToIPList = make(map[string][]string)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, fmt.Errorf("get addrs %s: %w", i.Name, err)
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if ip6 := v.IP.To16(); ip6 != nil {
					ifaceToIPList[i.Name] = append(ifaceToIPList[i.Name], ip6.String())
				}
			case *net.IPAddr:
				//	if ip4 := v.IP.To4(); ip4 != nil {
				//		res[i.Name] = append(res[i.Name], ip4.String())
				//	}
			}
		}
	}
	return ifaceToIPList, nil
}

func GetIFaceNames() []string {
	var a []string
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.Errorf("Cannot get network interfaces: %v", err)
		return a
	}
	for _, ifa := range ifaces {
		a = append(a, ifa.Name)
	}
	return a
}

func GetIPv4(ifaceName string) string {
	nameToIPList, err := GetIPv4Info()
	if err != nil {
		logger.Errorf("Cannot get ipv4 info: %v", err)
		return ""
	}

	if l, ok := nameToIPList[ifaceName]; ok {
		return l[0]
	}
	return ""
}

func GetIPv6(ifaceName string) string {
	nameToIPList, err := GetIPv6Info()
	if err != nil {
		logger.Errorf("Cannot get ipv6 info: %v", err)
		return ""
	}

	if l, ok := nameToIPList[ifaceName]; ok {
		return l[0]
	}
	return ""
}
