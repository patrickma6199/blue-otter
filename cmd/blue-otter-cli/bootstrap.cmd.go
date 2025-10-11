package main

import (
	"fmt"
	"strings"
	"context"
	"strconv"
	"bufio"
	"os"

	bootstrap "github.com/patrickma6199/blue-otter/internal/blueotterbootstrap"
	"github.com/urfave/cli/v2"
)

func bootstrapCmd(c *cli.Context) error {
	fmt.Println(`
    ____  __    __  ______   ____ _______________  ____     _____ __      ____
   / __ )/ /   / / / / __/  / __ /_  __/_  __/ __/ / __ \   / ___// /    /  _/
  / __  / /   / / / / /_   / / / // /   / / / /_  / /_/ /  / /   / /     / /  
 / /_/ / /___/ /_/ / __/  / /_/ // /   / / / __/ / _, _/  / /_  / /___  / / 
/_____/_____/\____/___/   \____//_/   /_/ /___/ /_/ |_|  /____//_____//___/  
																																		
BOOTSTRAP NODE - P2P Network Entry Point - v0.1.0                                                                           
	`)

	// Get port from command line or use default
	if c.String("port") == "" {
		fmt.Println("Port was not provided. Using default: 42069")
		c.Set("port", "42069")
	} else if _, err := strconv.Atoi(c.String("port")); err != nil {
		fmt.Println("Port must be a number. Using default: 42069")
		c.Set("port", "42069")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a quit channel for signaling termination
	quitCh := make(chan struct{})

	// Start the bootstrap node
	host, err := bootstrap.StartBootstrapNode(ctx, c.String("port"))
	if err != nil {
		return fmt.Errorf("failed to start bootstrap node: %w", err)
	}
	defer host.Close()

	fmt.Println("\nBootstrap node is running. Type /quit to exit.")
	fmt.Println("Other Blue Otter instances can now connect to this bootstrap node.")
	fmt.Println("Bootstrap info saved in ~/.blue-otter/bootstrap.json")

	// Read user input
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()

			switch text {
				case "/quit":
					fmt.Println("Shutting down bootstrap node...")
					close(quitCh)
					cancel()
					return
				case "/help":
					fmt.Println("Available commands:")
					fmt.Println("/quit - Exit the bootstrap node")
					fmt.Println("/help - Show this help message")
					fmt.Println("/list - List all connected peers")
					fmt.Println("/clear - Clear the console")
				case "/list":
					// List all connected peers
					peers := host.Peerstore().Peers()
					fmt.Println("Connected peers:")
					for _, peer := range peers {
						fmt.Printf("- %s\n", peer.String())
					}
				case "/clear":
					// Clear the console
					fmt.Print("\033[H\033[2J")
					fmt.Println("Console cleared.")
				default:
					if strings.HasPrefix(text, "/") {
						fmt.Println("Unknown command:", text)
					} else {
						fmt.Println("This is a bootstrap node. No messages can be sent.")
					}
			}
		}
	}()

	// Wait for quit signal
	<-quitCh

	return nil
}