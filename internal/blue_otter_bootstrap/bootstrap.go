package blue_otter_bootstrap

// bootstrap.go contains all functions related to bootstrap server functionality

import (
	"context"
	"fmt"
	"log"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	autonat "github.com/libp2p/go-libp2p/p2p/host/autonat"
	management "github.com/patrickma6199/blue-otter/internal/blue_otter_management"
)

// SetupConnectionNotifications (non-tui version) configures the host to log connection events
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

func StartBootstrapNode(ctx context.Context, port string, quitCh <-chan struct{}) (host.Host, error) {
	savedPrivKey, err := management.GetPrivateKey()
	if err != nil {
		log.Printf("[Networking] Warning: Failed to load private key: %v. Will create new identity.", err)
	}

	var options []libp2p.Option

	options = append(options,
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
		libp2p.EnableHolePunching(),
	)

	if savedPrivKey != nil {
		log.Println("[Networking] Using saved identity for node")
		options = append(options, libp2p.Identity(savedPrivKey))
	} else {
		log.Println("[Networking] Creating new identity for node")
	}

	host, err := libp2p.New(options...)
	if err != nil {
		return nil, fmt.Errorf("[Networking] Failed to create libp2p host: %w", err)
	}

	SetupConnectionNotifications(host)

	_, err = autonat.New(host)
	if err != nil {
		log.Printf("[Networking] AutoNAT warning: %v\n", err)
	}

	fmt.Println("[Networking] Bootstrap Node Started")
	fmt.Println("[Networking] Peer ID:", host.ID())
	fmt.Println("Listening on:")
	for _, addr := range host.Addrs() {
		fmt.Printf(" - %s/p2p/%s\n", addr, host.ID())
	}

	kDht, err := dht.New(ctx, host, dht.Mode(dht.ModeServer), dht.ProtocolPrefix("/ipfs/blue-otter"))
	if err != nil {
		return nil, fmt.Errorf("[Networking] Failed to create DHT: %w", err)
	}

	if err := kDht.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("[Networking] Failed to bootstrap DHT: %w", err)
	}

	disc := routing.NewRoutingDiscovery(kDht)

	go func() {
		deadPeers := make(map[peer.ID]time.Time)
		for {
			_, err := disc.Advertise(ctx, "--blue-otter-namespace--")
			if err != nil {
				if err.Error() != "failed to find any peer in table" {
					fmt.Println("[Discovery] Error advertising:", err)
				}
			}

			peerChan, err := disc.FindPeers(ctx, "--blue-otter-namespace--")
			if err != nil {
				if err.Error() != "failed to find any peer in table" {
					fmt.Println("[Discovery] Error finding peers:", err)
				}
				continue
			}

			for p := range peerChan {
				if p.ID == host.ID() || len(p.Addrs) == 0 {
					continue
				}

				if nextRetry, found := deadPeers[p.ID]; found && time.Now().Before(nextRetry) {
					continue
				}

				if host.Network().Connectedness(p.ID) != network.Connected {
					fmt.Println("[Discovery] Connecting to peer:", p.ID)
					if err := host.Connect(ctx, p); err != nil {
						fmt.Println("[Discovery] Failed to connect to peer:", err)
						deadPeers[p.ID] = time.Now().Add(1 * time.Minute)
					} else {
						delete(deadPeers, p.ID)
					}
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()

	if err := management.SaveAddressInfo(host); err != nil {
		log.Printf("[Config] Warning: Failed to save bootstrap info: %v", err)
	}

	return host, nil
}
