package blue_otter_client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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
	"github.com/rivo/tview"
)

// SetupConnectionNotifications configures the host to log connection events
func SetupConnectionNotifications(host host.Host, systemLogView *tview.TextView) {
	host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			remoteAddr := conn.RemoteMultiaddr()
			systemLogView.Write([]byte(fmt.Sprintf("[Networking] Connected to peer: %s via %s\n", remotePeer.String(), remoteAddr)))
		},
		DisconnectedF: func(n network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			remoteAddr := conn.RemoteMultiaddr()
			systemLogView.Write([]byte(fmt.Sprintf("[Networking] Disconnected from peer: %s via %s\n", remotePeer.String(), remoteAddr)))
		},
	})
}

func StartServer(ctx context.Context, username string, roomName string, port string, quitCh <-chan struct{}, chatView *tview.TextView, systemLogView *tview.TextView) (host.Host, *pubsub.Subscription, *pubsub.Topic) {
	host := networkConfiguration(ctx, port, systemLogView)

	// Set up connection notifications
	SetupConnectionNotifications(host, systemLogView)

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
						systemLogView.Write([]byte(fmt.Sprintf("[%s | notification] %s\n", roomName, sysMsg.Message)))
					} else {
						// If all parsing fails, fallback to raw
						chatView.Write([]byte(fmt.Sprintf("[%s] <%s> (unparsed): %s\n", roomName, msg.ReceivedFrom, string(msg.Data))))
					}
					continue
				}

				if(chatMsg.Sender != "" && chatMsg.Text != "") {
					chatView.Write([]byte(fmt.Sprintf("[%s] <%s>: %s\n", roomName, chatMsg.Sender, chatMsg.Text)))
				}
			}
		}
	}()

	return host, sub, topic
}

func networkConfiguration(ctx context.Context, port string, systemLogView *tview.TextView) host.Host {
	// ---------------------- Network Connection Configuration ----------------------

	savedPrivKey, err := management.GetPrivateKey()
	if err != nil {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] Warning: Failed to load private key: %v. Will create new identity.\n", err)))
	}

	var options []libp2p.Option

	// Add basic options
	options = append(options,
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
		libp2p.EnableHolePunching(),
	)

	// Add identity option if we have a saved key
	if savedPrivKey != nil {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] Using saved identity for bootstrap node\n")))
		options = append(options, libp2p.Identity(savedPrivKey))
	} else {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] Creating new identity for bootstrap node\n")))
	}

	// Initialize libp2p host with the specified options
	host, err := libp2p.New(options...)
	if err != nil {
		log.Fatal(err)
	}
	systemLogView.Write([]byte(fmt.Sprintf("[Networking] Host created. We are %s\n", host.ID())))

	// Set up AutoNAT for sensing if host is behind a NAT and helping with Hole Punching
	_, err = autonat.New(host)
	if err != nil {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] AutoNAT warning: %v\n", err)))
	}

	systemLogView.Write([]byte(fmt.Sprintf("[Networking] My Peer ID: %s\n", host.ID())))
	for _, addr := range host.Addrs() {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] Listening on: %s/p2p/%s\n", addr, host.ID())))
	}

	// Set up the Kademlia DHT for peer discovery
	kDht, err := dht.New(ctx, host, dht.Mode(dht.ModeClient), dht.ProtocolPrefix("/blue-otter"))
	if err != nil {
		log.Fatal(err)
	}
	if err := kDht.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	// Load bootstrap addresses
	bootstrapAddrs, err := management.LoadBootstrapAddressesForConnections()
	if err != nil {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] Warning: Failed to load bootstrap addresses: %v\n", err)))
		bootstrapAddrs = []string{} // Use empty list if loading fails
	}

	// If no bootstrap addresses found, use default placeholder
	if len(bootstrapAddrs) == 0 {
		bootstrapAddrs = []string{}
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] No bootstrap peers found. Please add some using the management commands.\n")))
	} else {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] Loaded %d bootstrap peers\n", len(bootstrapAddrs))))
	}

	// Connect to bootstrap peers
	for _, ba := range bootstrapAddrs {
		maddr, err := multiaddr.NewMultiaddr(ba)
		if err != nil {
			systemLogView.Write([]byte(fmt.Sprintf("[Networking] Invalid bootstrap address: %s, error: %v\n", ba, err)))
			continue
		}
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			systemLogView.Write([]byte(fmt.Sprintf("[Networking] Failed to get peer info from address: %s, error: %v\n", ba, err)))
			continue
		}
		if err := host.Connect(ctx, *info); err == nil {
			systemLogView.Write([]byte(fmt.Sprintf("[Networking] Connected to bootstrap: %s\n", info.String())))
		} else {
			systemLogView.Write([]byte(fmt.Sprintf("[Networking] Failed to connect to bootstrap peer %s: %v\n", info.ID, err)))
		}
	}

	disc := routing.NewRoutingDiscovery(kDht)
	
	go func() {
		deadPeers := make(map[peer.ID]time.Time)

		for {
			// 1) Advertise so others can discover us
			_, err := disc.Advertise(ctx, "--blue-otter-namespace--")
			if err != nil {
				if err.Error() != "failed to find any peer in table" {
					systemLogView.Write([]byte(fmt.Sprintf("[Discovery] Error advertising: %v\n", err)))
				}
			}

			// 2) Find all peers in that namespace
			peerChan, err := disc.FindPeers(ctx, "--blue-otter-namespace--")
			if err != nil {
				if err.Error() != "failed to find any peer in table" {
					systemLogView.Write([]byte(fmt.Sprintf("[Discovery] Error finding peers: %v\n", err)))
				}
				continue
			}

			// 3) Connect to each newly discovered peer
			for p := range peerChan {
				// Skip self or invalid addresses
				if p.ID == host.ID() || len(p.Addrs) == 0 {
					continue
				}
				
				// If we have a recorded “dead” status for this peer, skip unless cooldown has passed
				if nextRetry, found := deadPeers[p.ID]; found && time.Now().Before(nextRetry) {
					continue
				}

				if host.Network().Connectedness(p.ID) != network.Connected {
					systemLogView.Write([]byte(fmt.Sprintf("[Discovery] Connecting to peer from peer list: %s\n", p.ID)))
					if err := host.Connect(ctx, p); err != nil {
						systemLogView.Write([]byte(fmt.Sprintf("[Discovery] Failed to connect to peer from peer list: %v\nRetrying in 20 minutes...\n", err)))
						deadPeers[p.ID] = time.Now().Add(20 * time.Minute)
					} else {
						// Network notification system handles notifying user that peer is connected
						delete(deadPeers, p.ID)
					}
				}
			}

			// Sleep a bit before the next round
			time.Sleep(5 * time.Second)
		}
	}()

		// Save bootstrap info to file
	if err := management.SaveAddressInfo(host); err != nil {
		systemLogView.Write([]byte(fmt.Sprintf("[Config] Warning: Failed to save bootstrap info: %v\n", err)))
	}

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
