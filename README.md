# wireguard-manager

Daemon for configuring peers for wireguard interfaces, as well as collecting metrics from wireguard

### Building

Clone this repository, and run `make` to build.
This will produce a `wireguard-manager` binary and put them in your `GOBIN`.

## Testing
To run the tests, run `make test`.
To run the integration tests as well, run `go test ./...`. Note that this requires wireguard to be running on the machine, and root privileges.

## Usage
All options can be either configured via command line flags, or via their respective environment variable, as denoted by `[ENVIRONMENT_VARIABLE]`.

```
$ wireguard-manager -h
Usage of wireguard-manager:
  -delay duration
    	max random delay for the synchronization [WG_DELAY] (default 45s)
  -interfaces string
    	wireguard interfaces to configure. Pass a comma delimited list to configure multiple interfaces, eg 'wg0,wg1,wg2' [WG_INTERFACES] (default "wg0")
  -interval duration
    	how often wireguard peers will be synchronized with the api [WG_INTERVAL] (default 1m0s)
  -ipv4 string
    	ipv4 net to use for peer ip addresses [WG_IPV4] (default "10.99.0.0/16")
  -ipv6 string
    	ipv4 net to use for peer ip addresses [WG_IPV6] (default "fc00:bbbb:bbbb:bb01::/64")
  -password string
    	api password [WG_PASSWORD]
  -url string
    	api url [WG_URL] (default "https://api.mullvad.net")
  -username string
    	api username [WG_USERNAME]
```
