package helpers

import "net"

func CheckIPInSubnet(subnetStr string, ipStr string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return false, err
	}

	ip := net.ParseIP(ipStr)

	return ipNet.Contains(ip), nil
}
