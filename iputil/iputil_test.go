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
		ExpectedResult   string
	}{
		{1, "10.99.0.1/32"},
		{10, "10.99.0.10/32"},
		{100, "10.99.0.100/32"},
		{1000, "10.99.3.232/32"},
		{10000, "10.99.39.16/32"},
		{100000, "10.100.134.160/32"},
		{4120707071, "255.255.255.255/32"},
	}

	for _, test := range tests {
		ip, err := iputil.GetIPv4(ipv4Net, test.LeastSignificant)
		if err != nil {
			t.Errorf("%d: %+v", test.LeastSignificant, err)
		}

		if ip.String() != test.ExpectedResult {
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
		ExpectedResult   string
	}{
		{1, "fc00:bbbb:bbbb:bb01::1/128"},
		{10, "fc00:bbbb:bbbb:bb01::a/128"},
		{100, "fc00:bbbb:bbbb:bb01::64/128"},
		{1000, "fc00:bbbb:bbbb:bb01::3e8/128"},
		{10000, "fc00:bbbb:bbbb:bb01::2710/128"},
		{100000, "fc00:bbbb:bbbb:bb01::1:86a0/128"},
	}

	for _, test := range tests {
		ip, err := iputil.GetIPv6(ipv6Net, test.LeastSignificant)
		if err != nil {
			t.Errorf("%d: %+v", test.LeastSignificant, err)
		}

		if ip.String() != test.ExpectedResult {
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
