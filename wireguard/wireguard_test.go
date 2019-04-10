package wireguard_test

import (
	"encoding/base64"
	"net"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/infosum/statsd"
	"github.com/mdlayher/wireguardctrl"
	"github.com/mdlayher/wireguardctrl/wgtypes"
	"github.com/mullvad/wireguard-manager/api"
	"github.com/mullvad/wireguard-manager/wireguard"
)

// Integration tests for wireguard, not ran in short mode
// Requires a wireguard interface named wg0 to be running on the system

const testInterface = "wg0"

var ipv4Net = net.ParseIP("10.99.0.0")
var ipv6Net = net.ParseIP("fc00:bbbb:bbbb:bb01::")

var apiFixture = api.WireguardPeerList{
	api.WireguardPeer{
		IPLeastsig: 1,
		Ports:      []int{1234, 4321},
		Pubkey:     base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))),
	},
}

var peerFixture = []wgtypes.Peer{{
	PublicKey: wgKey(),
	AllowedIPs: []net.IPNet{
		net.IPNet{
			IP:   net.ParseIP("10.99.0.1"),
			Mask: net.CIDRMask(32, 32),
		},
		net.IPNet{
			IP:   net.ParseIP("fc00:bbbb:bbbb:bb01::1"),
			Mask: net.CIDRMask(128, 128),
		},
	},
	ProtocolVersion: 1,
}}

func wgKey() wgtypes.Key {
	key, _ := wgtypes.NewKey([]byte(strings.Repeat("a", 32)))
	return key
}

func TestWireguard(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	metrics, err := statsd.New()
	if err != nil {
		t.Fatal(err)
	}

	client, err := wireguardctrl.New()
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()
	defer resetDevice(t, client)

	wg, err := wireguard.New([]string{testInterface}, ipv4Net, ipv6Net, metrics)
	if err != nil {
		t.Fatal(err)
	}
	defer wg.Close()

	t.Run("add peers", func(t *testing.T) {
		wg.UpdatePeers(apiFixture)

		device, err := client.Device(testInterface)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(peerFixture, device.Peers); diff != "" {
			t.Fatalf("unexpected peers (-want +got):\n%s", diff)
		}
	})

	t.Run("update peer ip", func(t *testing.T) {
		apiFixture[0].IPLeastsig = 2
		wg.UpdatePeers(apiFixture)

		device, err := client.Device(testInterface)
		if err != nil {
			t.Fatal(err)
		}

		peerFixture[0].AllowedIPs[0].IP = net.ParseIP("10.99.0.2")
		peerFixture[0].AllowedIPs[1].IP = net.ParseIP("fc00:bbbb:bbbb:bb01::2")

		if diff := cmp.Diff(peerFixture, device.Peers); diff != "" {
			t.Fatalf("unexpected peers (-want +got):\n%s", diff)
		}
	})

	t.Run("remove peers", func(t *testing.T) {
		wg.UpdatePeers(api.WireguardPeerList{})

		device, err := client.Device(testInterface)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff([]wgtypes.Peer(nil), device.Peers); diff != "" {
			t.Fatalf("unexpected peers (-want +got):\n%s", diff)
		}
	})
}

func resetDevice(t *testing.T, c *wireguardctrl.Client) {
	t.Helper()

	cfg := wgtypes.Config{
		ReplacePeers: true,
	}

	if err := c.ConfigureDevice(testInterface, cfg); err != nil {
		t.Fatalf("failed to reset%v", err)
	}
}

func TestInvalidInterface(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	interfaceName := "nonexistant"

	_, err := wireguard.New([]string{interfaceName}, net.ParseIP("127.0.0.1"), net.ParseIP("::1"), nil)
	if err == nil {
		t.Fatal("no error")
	}
}
