package iputil

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"sort"
)

// GetIPv4 returns an ipv4 from the given subnet, with the given least significant bits
func GetIPv4(ipv4Net net.IP, ipLeastSig int) (ipv4 *net.IPNet, err error) {
	ip, err := ipv4ToInt(ipv4Net)
	if err != nil {
		return nil, err
	}

	ip += uint32(ipLeastSig)
	return &net.IPNet{
		IP:   int2ipv4(ip),
		Mask: net.CIDRMask(32, 32),
	}, nil
}

func ipv4ToInt(ip net.IP) (uint32, error) {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0, fmt.Errorf("couldn't convert ip to ipv4")
	}

	return binary.BigEndian.Uint32(ipv4), nil
}

func int2ipv4(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

// GetIPv6 returns an ipv6 from the given subnet, with the given least significant bits
func GetIPv6(ipv6Net net.IP, ipLeastSig int) (ipv6 *net.IPNet, err error) {
	ip, err := ipv6ToInt(ipv6Net)
	if err != nil {
		return nil, err
	}

	ip = ip.Add(ip, big.NewInt(int64(ipLeastSig)))
	return &net.IPNet{
		IP:   int2ipv6(ip),
		Mask: net.CIDRMask(128, 128),
	}, nil
}

func ipv6ToInt(ip net.IP) (*big.Int, error) {
	ipv6 := ip.To16()
	if ipv6 == nil {
		return nil, fmt.Errorf("couldn't convert ip to ipv6")
	}

	ipv6Int := big.NewInt(0)
	ipv6Int.SetBytes(ip)
	return ipv6Int, nil
}

func int2ipv6(nn *big.Int) net.IP {
	ip := net.IP(nn.Bytes())
	return ip
}

// sortIPNet sorts a slice of IPNet
func sortIPNet(ips []net.IPNet) func(i int, j int) bool {
	return func(i int, j int) bool {
		return ips[i].String() < ips[j].String()
	}
}

// EqualIPNet checks whether two slices of IPNet are equal
func EqualIPNet(a []net.IPNet, b []net.IPNet) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	sort.Slice(a, sortIPNet(a))
	sort.Slice(b, sortIPNet(b))

	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}

	return true
}
