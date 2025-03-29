package blue_otter_connect

import (
	"context"
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peer "github.com/libp2p/go-libp2p/core/peer"
	autonat "github.com/libp2p/go-libp2p/p2p/host/autonat"
	discovery "github.com/libp2p/go-libp2p/p2p/discovery/discovery"
	routing "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	multiaddr "github.com/multiformats/go-multiaddr"
)

func StartServer(ctx context.Context, roomName string) {
	// Initialize libp2p host
	host, err := libp2p.New(
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer host.Close()
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

	disc := routing.RoutingDiscovery(kDht)
    discovery.Advertise(ctx, disc, roomName)
}