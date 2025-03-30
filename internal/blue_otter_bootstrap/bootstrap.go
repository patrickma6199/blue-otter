package blue_otter_bootstrap

import (
	"context"
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	autonat "github.com/libp2p/go-libp2p/p2p/host/autonat"
)

// StartBootstrapNode starts a libp2p node in bootstrap mode
func StartBootstrapNode(ctx context.Context, port string, quitCh <-chan struct{}) (host.Host, error) {
	// Initialize libp2p host with the specified port
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Set up AutoNAT for sensing if host is behind a NAT and helping with Hole Punching
	_, err = autonat.New(host)
	if err != nil {
		log.Printf("AutoNAT warning: %v\n", err)
	}

	// Display node information
	fmt.Println("Bootstrap node started with ID:", host.ID())
	fmt.Println("This node can be reached at:")
	for _, addr := range host.Addrs() {
		fmt.Printf(" - %s/p2p/%s\n", addr, host.ID())
	}

	// Set up DHT in server mode for better peer discovery
	kDht, err := dht.New(ctx, host, dht.Mode(dht.ModeServer))
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Start DHT bootstrap process
	if err := kDht.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Wait for quit signal
	go func() {
		<-quitCh
		fmt.Println("Shutting down bootstrap node...")
	}()

	return host, nil
}