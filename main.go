package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DMarby/jitter"
	"github.com/infosum/statsd"
	"github.com/jamiealquiza/envy"
	"github.com/mullvad/wireguard-manager/api"
	"github.com/mullvad/wireguard-manager/portforward"
	"github.com/mullvad/wireguard-manager/wireguard"
)

var (
	a          *api.API
	wg         *wireguard.Wireguard
	pf         *portforward.Portforward
	metrics    *statsd.Client
	appVersion string // Populated during build time
)

func main() {
	// Set up commandline flags
	interval := flag.Duration("interval", time.Minute, "how often wireguard peers will be synchronized with the api")
	delay := flag.Duration("delay", time.Second*45, "max random delay for the synchronization")
	url := flag.String("url", "https://api.mullvad.net", "api url")
	username := flag.String("username", "", "api username")
	password := flag.String("password", "", "api password")
	interfaces := flag.String("interfaces", "wg0", "wireguard interfaces to configure. Pass a comma delimited list to configure multiple interfaces, eg 'wg0,wg1,wg2'")
	portForwardingChain := flag.String("portforwarding-chain", "PORTFORWARDING", "iptables chain to use for portforwarding")
	portForwardingExitAddresses := flag.String("portforwarding-exit-addresses", "", "exit addresses to use for portforwarding. Pass a comma delimited list to configure multiple IPs, eg '127.0.0.1,127.0.0.2'")
	statsdAddress := flag.String("statsd-address", "127.0.0.1:8125", "statsd address to send metrics to")

	// Parse environment variables
	envy.Parse("WG")

	// Add flag to output the version
	version := flag.Bool("v", false, "prints current app version")

	// Parse commandline flags
	flag.Parse()

	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	log.Printf("starting wireguard-manager %s", appVersion)

	// Initialize metrics
	var err error
	metrics, err = statsd.New(statsd.TagsFormat(statsd.Datadog), statsd.Prefix("wireguard"), statsd.Address(*statsdAddress))
	if err != nil {
		log.Fatalf("Error initializing metrics %s", err)
	}
	defer metrics.Close()

	// Initialize the API
	a = &api.API{
		Username: *username,
		Password: *password,
		BaseURL:  *url,
		Client: &http.Client{
			Timeout: time.Second * 10, // By default http.Client doesn't have a timeout, so specify a reasonable one
		},
	}

	// Initialize Wireguard
	if *interfaces == "" {
		log.Fatalf("no wireguard interfaces configured")
	}

	interfacesList := strings.Split(*interfaces, ",")

	wg, err = wireguard.New(interfacesList, metrics)
	if err != nil {
		log.Fatalf("error initializing wireguard %s", err)
	}
	defer wg.Close()

	// Initialize portforward
	if *portForwardingExitAddresses == "" {
		log.Fatalf("no portforwarding exit addresses configured")
	}

	addressesList := strings.Split(*portForwardingExitAddresses, ",")

	pf, err = portforward.New(addressesList, *portForwardingChain)
	if err != nil {
		log.Fatalf("error initializing portforwarding %s", err)
	}

	// Set up context for shutting down
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	// Run an initial synchronization
	synchronize()

	// Create a ticker to run our logic for polling the api and updating wireguard peers
	ticker := jitter.NewTicker(*interval, *delay)
	go func() {
		for {
			select {
			case <-ticker.C:
				// We run this synchronously, the ticker will drop ticks if this takes too long
				// This way we don't need a mutex or similar to ensure it doesn't run concurrently either
				synchronize()
			case <-shutdownCtx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	// Wait for shutdown or error
	err = waitForInterrupt(shutdownCtx)
	log.Printf("shutting down: %s", err)
}

func synchronize() {
	peers, err := a.GetWireguardPeers()
	if err != nil {
		metrics.Increment("error_getting_peers")
		log.Printf("error getting peers %s", err.Error())
		return
	}

	wg.UpdatePeers(peers)
	pf.UpdatePortforwarding(peers)
}

func waitForInterrupt(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		return fmt.Errorf("received signal %s", sig)
	case <-ctx.Done():
		return errors.New("canceled")
	}
}
