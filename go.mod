module github.com/mullvad/wireguard-manager

require (
	github.com/DMarby/jitter v0.0.0-20190312004500-d77fd504dcfa
	github.com/coreos/go-iptables v0.4.1
	github.com/google/go-cmp v0.3.0
	github.com/infosum/statsd v2.1.2+incompatible
	github.com/jamiealquiza/envy v1.1.0
	github.com/mdlayher/genetlink v0.0.0-20190617154021-985b2115c31a // indirect
	github.com/mdlayher/netlink v0.0.0-20190617153422-f82a9b10b2bc // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4 // indirect
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7 // indirect
	golang.org/x/sys v0.0.0-20190626221950-04f50cda93cb // indirect
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20190614145803-89b2114fdddf
)

replace golang.zx2c4.com/wireguard/wgctrl => github.com/mullvad/wgctrl-go v0.0.0-20190628223546-1ecbd52bc61b
