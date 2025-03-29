package blue_otter_connect

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-pubsub"
)

func StartServer(ctx context.Context, port string, peerList []string) {
	// Initialize libp2p host
	host, err := libp2p.New()
	if err != nil {
		fmt.Println("Error creating lib p2p host:", err)
		return
	}
	fmt.Printf("Host created. We are %s\n", host.ID())
	for _, la := range host.Addrs() {
		fmt.Printf(" - %v/p2p/%s\n", la, host.ID())
	}

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		log.Fatalf("Failed to create pubsub: %v", err)
	}

	// Your code to start the server goes here...
}