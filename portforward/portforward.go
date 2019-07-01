package portforward

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/mullvad/wireguard-manager/iputil"

	"github.com/coreos/go-iptables/iptables"
	"github.com/mullvad/wireguard-manager/api"
)

// Portforward is a utility for managing portforwarding
type Portforward struct {
	iptables  *iptables.IPTables
	ip6tables *iptables.IPTables
	chain     string
	ips       []net.IP
	ipv4Net   net.IP
	ipv6Net   net.IP
}

// Iptables table to operate against
const table = "nat"

// New validates the addresses, ensures that the iptables portforwarding chain exists, and returns a new Portforward instance
func New(addresses []string, chain string, ipv4Net net.IP, ipv6Net net.IP) (*Portforward, error) {
	var ips []net.IP
	for _, address := range addresses {
		ip := net.ParseIP(address)
		if ip == nil {
			return nil, fmt.Errorf("%s is not a valid ip", address)
		}

		ips = append(ips, ip)
	}

	ipt, err := newIPTables(chain, iptables.ProtocolIPv4)
	if err != nil {
		return nil, err
	}

	ip6t, err := newIPTables(chain, iptables.ProtocolIPv6)
	if err != nil {
		return nil, err
	}

	return &Portforward{
		iptables:  ipt,
		ip6tables: ip6t,
		chain:     chain,
		ips:       ips,
		ipv4Net:   ipv4Net,
		ipv6Net:   ipv6Net,
	}, nil
}

func newIPTables(chain string, protocol iptables.Protocol) (*iptables.IPTables, error) {
	ipt, err := iptables.NewWithProtocol(protocol)
	if err != nil {
		return nil, err
	}

	exists, err := chainExists(chain, ipt)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("an iptables chain named %s does not exist", chain)
	}

	return ipt, nil
}

func chainExists(chain string, ipt *iptables.IPTables) (bool, error) {
	chains, err := ipt.ListChains("nat")

	if err != nil {
		return false, err
	}

	for _, c := range chains {
		if c == chain {
			return true, nil
		}
	}

	return false, nil
}

// UpdatePortforwarding updates the iptables rules for portforwarding to match the given list of peers
func (p *Portforward) UpdatePortforwarding(peers api.WireguardPeerList) {
	rules := make(map[string]iptables.Protocol)
	for _, peer := range peers {
		if len(peer.Ports) < 1 {
			continue
		}

		// Ignore ip's with errors, in-case we get bad data from the API
		for _, ip := range p.ips {
			if ip.To4() != nil {
				ipv4, err := iputil.GetIPv4(p.ipv4Net, peer.IPLeastsig)
				if err != nil {
					continue
				}

				tcpRule := fmt.Sprintf("-d %s -p tcp -m multiport --dports %s -j DNAT --to-destination %s", ip.String(), getPortsString(peer.Ports), ipv4.IP.String())
				udpRule := fmt.Sprintf("-d %s -p udp -m multiport --dports %s -j DNAT --to-destination %s", ip.String(), getPortsString(peer.Ports), ipv4.IP.String())
				rules[tcpRule] = iptables.ProtocolIPv4
				rules[udpRule] = iptables.ProtocolIPv4
			} else {
				ipv6, err := iputil.GetIPv6(p.ipv6Net, peer.IPLeastsig)
				if err != nil {
					continue
				}

				tcpRule := fmt.Sprintf("-d %s -p tcp -m multiport --dports %s -j DNAT --to-destination %s", ip.String(), getPortsString(peer.Ports), ipv6.IP.String())
				udpRule := fmt.Sprintf("-d %s -p udp -m multiport --dports %s -j DNAT --to-destination %s", ip.String(), getPortsString(peer.Ports), ipv6.IP.String())
				rules[tcpRule] = iptables.ProtocolIPv6
				rules[udpRule] = iptables.ProtocolIPv6
			}
		}
	}

	currentRules, err := p.getCurrentRules()
	if err != nil {
		log.Printf("error getting current iptables rules %s", err.Error())
		return
	}

	// Add new portforwarding rules
	for rule, protocol := range rules {
		if _, ok := currentRules[rule]; !ok {
			ipt := p.iptables
			if protocol == iptables.ProtocolIPv6 {
				ipt = p.ip6tables
			}

			err := ipt.Append(table, p.chain, strings.Split(rule, " ")...)
			if err != nil {
				log.Printf("error adding iptables rule")
				continue
			}
		}
	}

	// Remove old portforwarding rules
	for rule, protocol := range currentRules {
		if _, ok := rules[rule]; !ok {
			ipt := p.iptables
			if protocol == iptables.ProtocolIPv6 {
				ipt = p.ip6tables
			}

			err := ipt.Delete(table, p.chain, strings.Split(rule, " ")...)
			if err != nil {
				log.Printf("error deleting iptables rule")
				continue
			}
		}
	}
}

func getPortsString(ports []int) string {
	slice := make([]string, len(ports))
	for i, v := range ports {
		slice[i] = strconv.Itoa(v)
	}

	return strings.Join(slice, ",")
}

func (p *Portforward) getCurrentRules() (map[string]iptables.Protocol, error) {
	rules := make(map[string]iptables.Protocol)

	ipv4Rules, err := p.iptables.List("nat", p.chain)
	if err != nil {
		return nil, err
	}

	for _, rule := range p.filterRules(ipv4Rules) {
		rules[rule] = iptables.ProtocolIPv4
	}

	ipv6Rules, err := p.ip6tables.List("nat", p.chain)
	if err != nil {
		return nil, err
	}

	for _, rule := range p.filterRules(ipv6Rules) {
		rules[rule] = iptables.ProtocolIPv6
	}

	return rules, nil
}

func (p *Portforward) filterRules(rules []string) []string {
	// Remove the first entry as it's the rule for creating the chain
	if len(rules) > 0 {
		rules = rules[1:]
	}

	var filteredRules []string
	for _, rule := range rules {
		// Remove the chain name
		rule = strings.TrimPrefix(rule, fmt.Sprintf("-A %s ", p.chain))

		// Remove the ip masks
		rule = strings.Replace(rule, "/32", "", -1)
		rule = strings.Replace(rule, "/128", "", -1)

		filteredRules = append(filteredRules, rule)
	}

	return filteredRules
}
