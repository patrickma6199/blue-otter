package blue_otter_connect

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
	routing "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	autonat "github.com/libp2p/go-libp2p/p2p/host/autonat"
	multiaddr "github.com/multiformats/go-multiaddr"
)

//TODO: move this and messaging logic to its own file
type ChatMessage struct {
    Sender string `json:"sender"`
    Text   string `json:"text"`
}

func StartServer(ctx context.Context, username string, roomName string, port string, quitCh <-chan struct{}) (host.Host, *pubsub.Subscription, *pubsub.Topic) {
	host := networkConfiguration(ctx, roomName, port)

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
					// subscription closed
					return
				}
				
				var chatMsg ChatMessage
				err = json.Unmarshal(msg.Data, &chatMsg)
				if err != nil {
					// If we fail to parse, fallback to raw
					fmt.Printf("Message from %s (unparsed): %s\n", msg.ReceivedFrom, string(msg.Data))
					continue
				}

				// Now we can show: "Message from Alice: Hello"
				fmt.Printf("Message from %s: %s\n", chatMsg.Sender, chatMsg.Text)
			}
		}
	}()

	return host, sub, topic
}

func networkConfiguration(ctx context.Context, roomName string, port string) host.Host {
	// ---------------------- Network Connection Configuration ----------------------

	// Initialize libp2p host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Host created. We are %s\n", host.ID())

	// Set up AutoNAT for sensing if host is behind a NAT and helping with Hole Punching
	// Hole punching is important here because you want to be able to advertise your address
	// to other peers in the public network (from what I can understand)
	_, err = autonat.New(host)
	if err != nil {
		log.Printf("AutoNAT warning: %v\n", err)
	}

	fmt.Println("My Peer ID:", host.ID())
	for _, addr := range host.Addrs() {
		fmt.Printf(" - %s/p2p/%s\n", addr, host.ID())
	}

	//Set up the Kademlia DHT for peer discovery
	//dht.Mode(dht.ModeAuto) is used to automatically determine the best mod for the DHT:
	// - ModeServer: the node will act as a DHT server (store more routing data)
	// - ModeClient: the node will act as a DHT client (store less routing data)
	// - ModeAuto: the node will act as both a DHT server and client
	kDht, err := dht.New(ctx, host, dht.Mode(dht.ModeAuto))
	if err != nil {
		log.Fatal(err)
	}
	// Tells your node to start participating in the DHT so it can route queries and respond to lookups.
	if err := kDht.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	// TODO: implement way to persist bootstrap peer addresses
	// bootstrap peers are known peers that allow you to join the mesh network (decentralized)
	bootstrapAddrs := []string{"<multiaddr of bootstrap peer>"}
	for _, ba := range bootstrapAddrs {
		maddr, err := multiaddr.NewMultiaddr(ba)
		if err != nil {
			continue
		}
		info, _ := peer.AddrInfoFromP2pAddr(maddr)
		if err := host.Connect(ctx, *info); err == nil {
			fmt.Println("Connected to bootstrap:", info.String())
		}
	}

	disc := routing.NewRoutingDiscovery(kDht)
	disc.Advertise(ctx, roomName)

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
