package wireguard

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/infosum/statsd"
	"github.com/mullvad/wireguard-manager/api"

	"github.com/mdlayher/wireguardctrl"
	"github.com/mdlayher/wireguardctrl/wgtypes"
	"github.com/mullvad/wireguard-manager/iputil"
)

// Wireguard is a utility for managing wireguard configuration
type Wireguard struct {
	client     *wireguardctrl.Client
	interfaces []string
	ipv4Net    net.IP
	ipv6Net    net.IP
	metrics    *statsd.Client
}

// New ensures that the interfaces given are valid, and returns a new Wireguard instance
func New(interfaces []string, ipv4Net net.IP, ipv6Net net.IP, metrics *statsd.Client) (*Wireguard, error) {
	client, err := wireguardctrl.New()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		_, err := client.Device(i)
		if err != nil {
			return nil, fmt.Errorf("error getting wireguard interface %s: %s", i, err.Error())
		}
	}

	return &Wireguard{
		client:     client,
		interfaces: interfaces,
		ipv4Net:    ipv4Net,
		ipv6Net:    ipv6Net,
		metrics:    metrics,
	}, nil
}

// UpdatePeers updates the configuration of the wireguard interfaces to match the given list of peers
func (w *Wireguard) UpdatePeers(peers api.WireguardPeerList) {
	peerMap := w.mapPeers(peers)

	var connectedPeers int
	for _, d := range w.interfaces {
		device, err := w.client.Device(d)
		// Log an error, but move on, so that one broken wireguard interface doesn't prevent us from configuring the rest
		if err != nil {
			log.Printf("error connecting to wireguard interface %s: %s", d, err.Error())
			continue
		}

		connectedPeers += countConnectedPeers(device.Peers)

		existingPeerMap := mapExistingPeers(device.Peers)
		cfgPeers := []wgtypes.PeerConfig{}

		// Loop through peers from the API
		// Add ones not currently existing in the wireguard config
		// Update ones that exist in the wireguard config but has changed
		for key, allowedIPs := range peerMap {
			existingAllowedIPs, ok := existingPeerMap[key]
			if !ok || !iputil.EqualIPNet(allowedIPs, existingAllowedIPs) {
				cfgPeers = append(cfgPeers, wgtypes.PeerConfig{
					PublicKey:         key,
					ReplaceAllowedIPs: true,
					AllowedIPs:        allowedIPs,
				})
			}
		}

		// Loop through the current peers in the wireguard config and remove ones that doesn't exist in the API
		for key := range existingPeerMap {
			if _, ok := peerMap[key]; !ok {
				cfgPeers = append(cfgPeers, wgtypes.PeerConfig{
					PublicKey: key,
					Remove:    true,
				})
			}
		}

		// No changes needed
		if len(cfgPeers) == 0 {
			continue
		}

		err = w.client.ConfigureDevice(d, wgtypes.Config{
			Peers: cfgPeers,
		})
		if err != nil {
			log.Printf("error configuring to wireguard interface %s: %s", d, err.Error())
			continue
		}
	}

	// Send metrics
	w.metrics.Gauge("connected_peers", connectedPeers)
}

// Take the wireguard peers and convert them into a map for easier comparison
func (w *Wireguard) mapPeers(peers api.WireguardPeerList) (peerMap map[wgtypes.Key][]net.IPNet) {
	peerMap = make(map[wgtypes.Key][]net.IPNet)

	// Ignore peers with errors, in-case we get bad data from the API
	for _, peer := range peers {
		key, err := wgtypes.ParseKey(peer.Pubkey)
		if err != nil {
			continue
		}

		ipv4, err := iputil.GetIPv4(w.ipv4Net, peer.IPLeastsig)
		if err != nil {
			continue
		}

		ipv6, err := iputil.GetIPv6(w.ipv6Net, peer.IPLeastsig)
		if err != nil {
			continue
		}

		peerMap[key] = []net.IPNet{
			*ipv4,
			*ipv6,
		}
	}

	return
}

// Take the existing wireguard peers and convert them into a map for easier comparison
func mapExistingPeers(peers []wgtypes.Peer) (peerMap map[wgtypes.Key][]net.IPNet) {
	peerMap = make(map[wgtypes.Key][]net.IPNet)

	for _, peer := range peers {
		peerMap[peer.PublicKey] = peer.AllowedIPs
	}

	return
}

// Wireguard sends a handshake roughly every 2 minutes
// So we consider all peers with a handshake within that interval to be connected
const handshakeInterval = time.Minute * 2

// Count the connected wireguard peers
func countConnectedPeers(peers []wgtypes.Peer) (connectedPeers int) {
	for _, peer := range peers {
		if time.Since(peer.LastHandshakeTime) <= handshakeInterval {
			connectedPeers++
		}
	}

	return
}

// Close closes the underlying wireguad client
func (w *Wireguard) Close() {
	w.client.Close()
}
