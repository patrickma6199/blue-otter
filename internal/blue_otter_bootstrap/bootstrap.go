package blue_otter_bootstrap

import (
	"context"
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	autonat "github.com/libp2p/go-libp2p/p2p/host/autonat"
	management "github.com/patrickma6199/blue-otter/internal/blue_otter_management"
)

// SetupConnectionNotifications configures the host to log connection events
func SetupConnectionNotifications(host host.Host) {
	host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			remoteAddr := conn.RemoteMultiaddr()
			fmt.Printf("[Networking] New connection from peer: %s via %s\n", remotePeer.String(), remoteAddr)
		},
		DisconnectedF: func(n network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			remoteAddr := conn.RemoteMultiaddr()
			fmt.Printf("[Networking] Disconnected from peer: %s via %s\n", remotePeer.String(), remoteAddr)
		},
	})
}

// StartBootstrapNode starts a libp2p node in bootstrap mode
func StartBootstrapNode(ctx context.Context, port string, quitCh <-chan struct{}) (host.Host, error) {
	// First, try to get the saved private key
	savedPrivKey, err := management.GetBootstrapPrivateKey()
	if err != nil {
		log.Printf("[Networking] Warning: Failed to load private key: %v. Will create new identity.", err)
	}

	var options []libp2p.Option

	// Add basic options
	options = append(options,
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
		libp2p.EnableHolePunching(),
	)

	// Add identity option if we have a saved key
	if savedPrivKey != nil {
		log.Println("[Networking] Using saved identity for bootstrap node")
		options = append(options, libp2p.Identity(savedPrivKey))
	} else {
		log.Println("[Networking] Creating new identity for bootstrap node")
	}

	// Initialize libp2p host with the specified options
	host, err := libp2p.New(options...)
	if err != nil {
		return nil, fmt.Errorf("[Networking] Failed to create libp2p host: %w", err)
	}

	// Set up connection notifications
	SetupConnectionNotifications(host)

	// Set up AutoNAT for sensing if host is behind a NAT and helping with Hole Punching
	_, err = autonat.New(host)
	if err != nil {
		log.Printf("[Networking] AutoNAT warning: %v\n", err)
	}

	// Display node information
	fmt.Println("[Networking] Bootstrap Node Started")
	fmt.Println("[Networking] Peer ID:", host.ID())
	fmt.Println("Listening on:")
	for _, addr := range host.Addrs() {
		fmt.Printf(" - %s/p2p/%s\n", addr, host.ID())
	}

	// Set up DHT in server mode for better peer discovery
	kDht, err := dht.New(ctx, host, dht.Mode(dht.ModeServer))
	if err != nil {
		return nil, fmt.Errorf("[Networking] Failed to create DHT: %w", err)
	}

	// Start DHT bootstrap process
	if err := kDht.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("[Networking] Failed to bootstrap DHT: %w", err)
	}

	disc := routing.NewRoutingDiscovery(kDht)
	disc.Advertise(ctx, "--blue-otter-namespace--")

	// Save bootstrap info to file
	if err := management.SaveBootstrapInfo(host); err != nil {
		log.Printf("[Config] Warning: Failed to save bootstrap info: %v", err)
	}

	// Wait for quit signal
	go func() {
		<-quitCh
		fmt.Println("Shutting down bootstrap node...")
	}()

	return host, nil
}
