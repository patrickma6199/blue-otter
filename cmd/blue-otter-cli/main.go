package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/patrickma6199/blue-otter/internal/blue_otter_connect"
	"github.com/urfave/cli/v2"
)

var (
	peerList []string
	muPeerList sync.Mutex
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: blue-otter [port] [optional peer addresses...]")
		return
	}

	port := os.Args[1]

	if len(os.Args) > 2 {
		peerList = os.Args[2:]
	} else {
		peerList = []string{}
	}

	app := &cli.App{
		Name:  "blue-otter",
		Usage: "A CLI tool for Blue Otter",
		Commands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"s"},
				Usage:   "Start the Blue Otter service",
				Action: func(c *cli.Context) error {
					fmt.Printf("Starting Blue Otter on port %s with peers %v\n", port, peerList)
					ctx := context.Background()
					blue_otter_connect.StartServer(ctx, port, peerList)
					
					


				},
			},
			{
				Name:    "stop",
				Aliases: []string{"st"},
				Usage:   "Stop the Blue Otter service",
				Action: func(c *cli.Context) error {
					fmt.Println("Stopping Blue Otter")
					return nil
				},
			},
		},
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "Port to run the Blue Otter service on",
			Value:   port,
		},
	}
}