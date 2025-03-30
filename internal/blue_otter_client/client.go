package blue_otter_client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	routing "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	autonat "github.com/libp2p/go-libp2p/p2p/host/autonat"
	multiaddr "github.com/multiformats/go-multiaddr"
	common "github.com/patrickma6199/blue-otter/internal/blue_otter_common"
	management "github.com/patrickma6199/blue-otter/internal/blue_otter_management"
)

// SetupConnectionNotifications configures the host to log connection events
func SetupConnectionNotifications(host host.Host) {
	host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			remoteAddr := conn.RemoteMultiaddr()
			fmt.Printf("[Networking] Connected to peer: %s via %s\n", remotePeer.String(), remoteAddr)
		},
		DisconnectedF: func(n network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			remoteAddr := conn.RemoteMultiaddr()
			fmt.Printf("[Networking] Disconnected from peer: %s via %s\n", remotePeer.String(), remoteAddr)
		},
	})
}

func StartServer(ctx context.Context, username string, roomName string, port string, quitCh <-chan struct{}) (host.Host, *pubsub.Subscription, *pubsub.Topic) {
	host := networkConfiguration(ctx, port)

	// Set up connection notifications
	SetupConnectionNotifications(host)

	sub, topic := pubSubConfiguration(ctx, host, roomName)

	// 4. Read messages in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := sub.Next(ctx)
				if err != nil {
					return
				}

				// Try to parse as ChatMessage first
				var chatMsg common.ChatMessage
				err = json.Unmarshal(msg.Data, &chatMsg)
				if err != nil {
					// If we fail to parse as ChatMessage, try SystemNotification
					var sysMsg common.SystemNotification
					err = json.Unmarshal(msg.Data, &sysMsg)
					if err == nil {
						// Successfully parsed as system notification
						fmt.Printf("[%s | notification] %s\n", roomName, sysMsg.Message)
					} else {
						// If all parsing fails, fallback to raw
						fmt.Printf("[%s] Message from %s (unparsed): %s\n", roomName, msg.ReceivedFrom, string(msg.Data))
					}
					continue
				}

				// Now we can show: "Message from Alice: Hello"
				fmt.Printf("[%s] Message from %s: %s\n", roomName, chatMsg.Sender, chatMsg.Text)
			}
		}
	}()

	return host, sub, topic
}

func networkConfiguration(ctx context.Context, port string) host.Host {
	// ---------------------- Network Connection Configuration ----------------------

	// Initialize libp2p host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[Networking] Host created. We are %s\n", host.ID())

	// Set up AutoNAT for sensing if host is behind a NAT and helping with Hole Punching
	_, err = autonat.New(host)
	if err != nil {
		log.Printf("[Networking] AutoNAT warning: %v\n", err)
	}

	fmt.Println("[Networking] My Peer ID:", host.ID())
	for _, addr := range host.Addrs() {
		fmt.Printf(" - %s/p2p/%s\n", addr, host.ID())
	}

	// Set up the Kademlia DHT for peer discovery
	kDht, err := dht.New(ctx, host, dht.Mode(dht.ModeAuto))
	if err != nil {
		log.Fatal(err)
	}
	if err := kDht.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	// Load bootstrap addresses
	bootstrapAddrs, err := management.LoadBootstrapAddressesForConnections()
	if err != nil {
		log.Printf("[Networking] Warning: Failed to load bootstrap addresses: %v\n", err)
		bootstrapAddrs = []string{} // Use empty list if loading fails
	}

	// If no bootstrap addresses found, use default placeholder
	if len(bootstrapAddrs) == 0 {
		bootstrapAddrs = []string{}
		log.Println("[Networking] No bootstrap peers found. Please add some using the management commands.")
	} else {
		log.Printf("[Networking] Loaded %d bootstrap peers\n", len(bootstrapAddrs))
	}

	// Connect to bootstrap peers
	for _, ba := range bootstrapAddrs {
		maddr, err := multiaddr.NewMultiaddr(ba)
		if err != nil {
			log.Printf("[Networking] Invalid bootstrap address: %s, error: %v\n", ba, err)
			continue
		}
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Printf("[Networking] Failed to get peer info from address: %s, error: %v\n", ba, err)
			continue
		}
		if err := host.Connect(ctx, *info); err == nil {
			fmt.Println("[Networking] Connected to bootstrap:", info.String())
		} else {
			log.Printf("[Networking] Failed to connect to bootstrap peer %s: %v\n", info.ID, err)
		}
	}

	disc := routing.NewRoutingDiscovery(kDht)
	disc.Advertise(ctx, "--blue-otter-namespace--")

	return host
}

func pubSubConfiguration(ctx context.Context, host host.Host, roomName string) (*pubsub.Subscription, *pubsub.Topic) {
	// ---------------------- PubSub Configuration ----------------------

	// 1. Initialize PubSub
	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Join a topic (e.g. the same roomName, or "chat-topic")
	topic, err := ps.Join(roomName)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Subscribe to the topic
	sub, err := topic.Subscribe()
	if err != nil {
		log.Fatal(err)
	}

	return sub, topic
}
