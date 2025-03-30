package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/patrickma6199/blue-otter/internal/blue_otter_connect"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "blue-otter",
		Usage:   "A CLI tool for Blue Otter",
		Version: "0.1.0",
		Authors: []*cli.Author{
			{
				Name:  "Patrick Ma",
				Email: "patrickma6199@gmail.com",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"s"},
				Usage:   "Start the Blue Otter service",
				Action: func(c *cli.Context) error {

					fmt.Println(`
    ____  __    __  ______   ____ _______________  ____     _____ __      ____
   / __ )/ /   / / / / __/  / __ /_  __/_  __/ __/ / __ \   / ___// /    /  _/
  / __  / /   / / / / /_   / / / // /   / / / /_  / /_/ /  / /   / /     / /  
 / /_/ / /___/ /_/ / __/  / /_/ // /   / / / __/ / _, _/  / /_  / /___  / / 
/_____/_____/\____/___/   \____//_/   /_/ /___/ /_/ |_|  /____//_____//___/  
                                                                            
P2P Communication Made Simple - v0.1.0                                                                           
					`)

					if c.String("room") == "" {
						fmt.Println("Room name was not provided. Using default: --blue-otter-public-default")
						c.Set("room", "--blue-otter-public-default")
					} else if !strings.HasPrefix(c.String("room"), "--blue-otter-") {
						newRoom := "--blue-otter-" + c.String("room")
						fmt.Printf("Room name modified to have required prefix: %s\n", newRoom)
						c.Set("room", newRoom)
					}

					if c.String("port") == "" {
						fmt.Println("Port was not provided. Using default: 42069")
						c.Set("port", "42069")
					} else if _, err := strconv.Atoi(c.String("port")); err != nil {
						fmt.Println("Port must be a number. Using default: 42069")
						c.Set("port", "42069")
					}

					if c.String("username") == "" {
						fmt.Println("No username provided. Using default: Guest")
						c.Set("username", "Guest")
					}

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					// Create a quit channel for signaling termination
					quitCh := make(chan struct{})

					// Start the server and get the host
					host, _, topic := blue_otter_connect.StartServer(ctx, c.String("username"), c.String("room"), c.String("port"), quitCh)
					defer host.Close()

					fmt.Println("Blue Otter started! Type /quit to exit.")

					// Start a goroutine to read user input
					go func() {
						scanner := bufio.NewScanner(os.Stdin)
						for scanner.Scan() {
							text := scanner.Text()
							if text == "/quit" {
								fmt.Println("Shutting down Blue Otter...")
								close(quitCh)
								cancel()
								return
							} else if text == "" {
								continue
							}

							msg := blue_otter_connect.ChatMessage{Sender: c.String("username"), Text: text}
							data, err := json.Marshal(msg)
							if err != nil {
								fmt.Println("Error encoding message:", err)
								continue
							}

							topic.Publish(ctx, data)
						}
					}()

					// Wait for quit signal
					<-quitCh

					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "username",
						Aliases: []string{"u"},
						Usage:   "Username to display in chat",
					},
					&cli.StringFlag{
						Name:    "room",
						Aliases: []string{"r"},
						Usage:   "Room name to join",
					},
					&cli.StringFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "Port to run the Blue Otter service on",
					},
				},
			},
		},
	}

	app.Version = "0.1.0"
	app.EnableBashCompletion = true

	// Run the CLI application and handle any error
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
