package portforward_test

import (
	"encoding/base64"
	"net"
	"strings"
	"testing"

	"github.com/coreos/go-iptables/iptables"
	"github.com/google/go-cmp/cmp"
	"github.com/mullvad/wireguard-manager/api"
	"github.com/mullvad/wireguard-manager/portforward"
)

// Integration tests for portforwarding, not ran in short mode
// Requires an iptables nat chain named PORTFORWARDING in both iptables and ip6tables

var apiFixture = api.WireguardPeerList{
	api.WireguardPeer{
		IPLeastsig: 1,
		Ports:      []int{1234, 4321},
		Pubkey:     base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))),
	},
}

var rulesFixture = []string{
	"-A PORTFORWARDING -d 127.0.0.1/32 -p tcp -m multiport --dports 1234,4321 -j DNAT --to-destination 10.99.0.1",
	"-A PORTFORWARDING -d ::1/128 -p tcp -m multiport --dports 1234,4321 -j DNAT --to-destination fc00:bbbb:bbbb:bb01::1",
}

var ipv4Net = net.ParseIP("10.99.0.0")
var ipv6Net = net.ParseIP("fc00:bbbb:bbbb:bb01::")

const (
	chain = "PORTFORWARDING"
	table = "nat"
)

func TestPortforward(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	portforward, err := portforward.New([]string{"127.0.0.1", "::1"}, chain, ipv4Net, ipv6Net)
	if err != nil {
		t.Fatal(err)
	}

	ipts := setupIptables(t)

	t.Run("add rules", func(t *testing.T) {
		portforward.UpdatePortforwarding(apiFixture)

		rules := getRules(t, ipts)
		if diff := cmp.Diff(rulesFixture, rules); diff != "" {
			t.Fatalf("unexpected rules (-want +got):\n%s", diff)
		}
	})

	t.Run("remove rules", func(t *testing.T) {
		portforward.UpdatePortforwarding(api.WireguardPeerList{})

		rules := getRules(t, ipts)
		if diff := cmp.Diff([]string{}, rules); diff != "" {
			t.Fatalf("unexpected rules (-want +got):\n%s", diff)
		}
	})
}

func getRules(t *testing.T, ipts []*iptables.IPTables) []string {
	t.Helper()

	rules := []string{}
	for _, ipt := range ipts {
		listRules, err := ipt.List(table, chain)
		if err != nil {
			t.Fatal(err)
		}

		if len(listRules) > 0 {
			listRules = listRules[1:]
		}

		for _, rule := range listRules {
			rules = append(rules, rule)
		}
	}

	return rules
}

func setupIptables(t *testing.T) []*iptables.IPTables {
	t.Helper()

	ip4t, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		t.Fatal(err)
	}

	ip6t, err := iptables.NewWithProtocol(iptables.ProtocolIPv6)
	if err != nil {
		t.Fatal(err)
	}

	return []*iptables.IPTables{ip4t, ip6t}
}

func TestInvalidInterface(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	interfaceName := "nonexistant"
	_, err := portforward.New([]string{}, interfaceName, ipv4Net, ipv6Net)
	if err == nil {
		t.Fatal("no error")
	}
}

func TestInvalidIPs(t *testing.T) {
	interfaceName := "nonexistant"
	_, err := portforward.New([]string{"abcd"}, interfaceName, ipv4Net, ipv6Net)
	if err == nil {
		t.Fatal("no error")
	}
}
