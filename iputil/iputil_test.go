package iputil_test

import (
	"net"
	"testing"

	"github.com/mullvad/wireguard-manager/iputil"
)

var ipv4Net = net.ParseIP("10.99.0.0")
var ipv6Net = net.ParseIP("fc00:bbbb:bbbb:bb01::")

func TestIpv4(t *testing.T) {
	tests := []struct {
		LeastSignificant int
		ExpectedResult   net.IP
	}{
		{1, net.ParseIP("10.99.0.1")},
		{10, net.ParseIP("10.99.0.10")},
		{100, net.ParseIP("10.99.0.100")},
		{1000, net.ParseIP("10.99.3.232")},
		{10000, net.ParseIP("10.99.39.16")},
		{100000, net.ParseIP("10.100.134.160")},
		{4120707071, net.ParseIP("255.255.255.255")},
	}

	for _, test := range tests {
		ip, err := iputil.GetIPv4(ipv4Net, test.LeastSignificant)
		if err != nil {
			t.Errorf("%d: %+v", test.LeastSignificant, err)
		}

		if !ip.Equal(test.ExpectedResult) {
			t.Errorf("%d: got %s, expected %s", test.LeastSignificant, ip, test.ExpectedResult)
		}
	}
}

func TestInvalidIpv4(t *testing.T) {
	_, err := iputil.GetIPv4(net.ParseIP(""), 1)
	if err == nil {
		t.Fatalf("expected error")
	}

	if err.Error() != "couldn't convert ip to ipv4" {
		t.Errorf("got unexpected error %s", err)
	}
}

func TestIpv6(t *testing.T) {
	tests := []struct {
		LeastSignificant int
		ExpectedResult   net.IP
	}{
		{1, net.ParseIP("fc00:bbbb:bbbb:bb01::1")},
		{10, net.ParseIP("fc00:bbbb:bbbb:bb01::a")},
		{100, net.ParseIP("fc00:bbbb:bbbb:bb01::64")},
		{1000, net.ParseIP("fc00:bbbb:bbbb:bb01::3e8")},
		{10000, net.ParseIP("fc00:bbbb:bbbb:bb01::2710")},
		{100000, net.ParseIP("fc00:bbbb:bbbb:bb01::1:86a0")},
	}

	for _, test := range tests {
		ip, err := iputil.GetIPv6(ipv6Net, test.LeastSignificant)
		if err != nil {
			t.Errorf("%d: %+v", test.LeastSignificant, err)
		}

		if !ip.Equal(test.ExpectedResult) {
			t.Errorf("%d: got %s, expected %s", test.LeastSignificant, ip, test.ExpectedResult)
		}
	}
}

func TestInvalidIpv6(t *testing.T) {
	_, err := iputil.GetIPv6(net.ParseIP(""), 1)
	if err == nil {
		t.Fatalf("expected error")
	}

	if err.Error() != "couldn't convert ip to ipv6" {
		t.Errorf("got unexpected error %s", err)
	}
}
