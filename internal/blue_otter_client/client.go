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

	SetupConnectionNotifications(host, systemLogView)

	sub, topic := pubSubConfiguration(ctx, host, roomName)

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

				var chatMsg common.ChatMessage
				err = json.Unmarshal(msg.Data, &chatMsg)
				if err != nil {
					var sysMsg common.SystemNotification
					err = json.Unmarshal(msg.Data, &sysMsg)
					if err == nil {
						systemLogView.Write([]byte(fmt.Sprintf("[%s | notification] %s\n", roomName, sysMsg.Message)))
					} else {
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

	options = append(options,
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
		libp2p.EnableHolePunching(),
	)

	if savedPrivKey != nil {
		systemLogView.Write([]byte("[Networking] Using saved identity for node\n"))
		options = append(options, libp2p.Identity(savedPrivKey))
	} else {
		systemLogView.Write([]byte("[Networking] Creating new identity for node\n"))
	}

	host, err := libp2p.New(options...)
	if err != nil {
		log.Fatal(err)
	}
	systemLogView.Write([]byte(fmt.Sprintf("[Networking] Host created. We are %s\n", host.ID())))

	_, err = autonat.New(host)
	if err != nil {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] AutoNAT warning: %v\n", err)))
	}

	systemLogView.Write([]byte(fmt.Sprintf("[Networking] My Peer ID: %s\n", host.ID())))
	for _, addr := range host.Addrs() {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] Listening on: %s/p2p/%s\n", addr, host.ID())))
	}

	kDht, err := dht.New(ctx, host, dht.Mode(dht.ModeClient), dht.ProtocolPrefix("/ipfs/blue-otter"))
	if err != nil {
		log.Fatal(err)
	}
	if err := kDht.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	bootstrapAddrs, err := management.LoadBootstrapAddressesForConnections()
	if err != nil {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] Warning: Failed to load bootstrap addresses: %v\n", err)))
		bootstrapAddrs = []string{}
	}

	if len(bootstrapAddrs) == 0 {
		bootstrapAddrs = []string{}
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] No bootstrap peers found. Please add some using the management commands.\n")))
	} else {
		systemLogView.Write([]byte(fmt.Sprintf("[Networking] Loaded %d bootstrap peers\n", len(bootstrapAddrs))))
	}

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
			_, err := disc.Advertise(ctx, "--blue-otter-namespace--")
			if err != nil {
				if err.Error() != "failed to find any peer in table" {
					systemLogView.Write([]byte(fmt.Sprintf("[Discovery] Error advertising: %v\n", err)))
				}
			}

			peerChan, err := disc.FindPeers(ctx, "--blue-otter-namespace--")
			if err != nil {
				if err.Error() != "failed to find any peer in table" {
					systemLogView.Write([]byte(fmt.Sprintf("[Discovery] Error finding peers: %v\n", err)))
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
					systemLogView.Write([]byte(fmt.Sprintf("[Discovery] Connecting to peer from peer list: %s\n", p.ID)))
					if err := host.Connect(ctx, p); err != nil {
						systemLogView.Write([]byte(fmt.Sprintf("[Discovery] Failed to connect to peer from peer list: %v\nRetrying in 20 minutes...\n", err)))
						deadPeers[p.ID] = time.Now().Add(20 * time.Minute)
					} else {
						delete(deadPeers, p.ID)
					}
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()

	if err := management.SaveAddressInfo(host); err != nil {
		systemLogView.Write([]byte(fmt.Sprintf("[Config] Warning: Failed to save bootstrap info: %v\n", err)))
	}

	return host
}

func pubSubConfiguration(ctx context.Context, host host.Host, roomName string) (*pubsub.Subscription, *pubsub.Topic) {
	// ---------------------- PubSub Configuration ----------------------

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		log.Fatal(err)
	}

	topic, err := ps.Join(roomName)
	if err != nil {
		log.Fatal(err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		log.Fatal(err)
	}

	return sub, topic
}
