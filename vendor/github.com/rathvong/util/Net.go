package util

import (
	"net"
	"log"
)

func ParseRequestIP(address string) (ipNet net.IPNet, err error) {
	ipString, port, err := net.SplitHostPort(address)
	ip := net.ParseIP(ipString)
	mask := ip.DefaultMask()
	ipNet = net.IPNet{IP: ip, Mask: mask}
	log.Printf("Port: %v Ip: %v err: %v", port, ipString, err)
	return
}

func ConvertStringToIpNet(address string) (net.IPNet, error) {

	ip, _, err := net.ParseCIDR(address)

	if err != nil {
		ip := net.ParseIP(address)
		mask := ip.DefaultMask()

		return net.IPNet{IP: ip, Mask: mask}, nil

	}
	mask := ip.DefaultMask()

	return net.IPNet{IP: ip, Mask: mask}, err

}

