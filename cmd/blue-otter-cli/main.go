package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	tcell "github.com/gdamore/tcell/v2"
	bootstrap "github.com/patrickma6199/blue-otter/internal/blue_otter_bootstrap"
	client "github.com/patrickma6199/blue-otter/internal/blue_otter_client"
	common "github.com/patrickma6199/blue-otter/internal/blue_otter_common"
	management "github.com/patrickma6199/blue-otter/internal/blue_otter_management"
	tui "github.com/patrickma6199/blue-otter/internal/blue_otter_tui"
	"github.com/rivo/tview"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "blue-otter-cli",
		Usage:   "CLI Interface for Blue Otter Mesh Messaging",
		Version: "0.1.0",
		Authors: []*cli.Author{
			{
				Name:  "Patrick Ma",
				Email: "patrickma6199@gmail.com",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "client",
				Aliases: []string{"c"},
				Usage:   "Start the Blue Otter service",
				Action: func(c *cli.Context) error {

					fmt.Println(`
    ____  __    __  ______   ____ _______________  ____     _____ __      ____
   / __ )/ /   / / / / __/  / __ /_  __/_  __/ __/ / __ \   / ___// /    /  _/
  / __  / /   / / / / /_   / / / // /   / / / /_  / /_/ /  / /   / /     / /  
 / /_/ / /___/ /_/ / __/  / /_/ // /   / / / __/ / _, _/  / /_  / /___  / / 
/_____/_____/\____/___/   \____//_/   /_/ /___/ /_/ |_|  /____//_____//___/  
                                                                            
CLIENT NODE - v0.1.0                                                                           
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

					app := tview.NewApplication()

					layout, _, chatView, systemLogView, inputField := tui.CreateUI(c.String("username"), c.String("room"))

					// Start the server and get the host
					host, _, topic := client.StartServer(ctx, c.String("username"), c.String("room"), c.String("port"), quitCh, chatView, systemLogView)
					defer host.Close()

					// Announce our arrival
					joinMsg := common.SystemNotification{
						Type:    "join",
						Message: fmt.Sprintf("[%s] User %s has joined the room", c.String("room"), c.String("username")),
					}
					joinData, _ := json.Marshal(joinMsg)
					topic.Publish(ctx, joinData)

					chatView.Write([]byte(fmt.Sprintf("[%s] Blue Otter started! Type /quit to exit.\n", c.String("room"))))

					// Set up the input field to send messages
					inputField.SetDoneFunc(func(key tcell.Key) {
						text := inputField.GetText()
						defer inputField.SetText("")
						if strings.TrimSpace(text) == "" {
							return
						}
						
						switch text {
						case "/quit":
							// Send leave message before quitting
							leaveMsg := common.SystemNotification{
								Type:    "leave",
								Message: fmt.Sprintf("[%s] User %s has left the room", c.String("room"), c.String("username")),
							}
							leaveData, _ := json.Marshal(leaveMsg)
							topic.Publish(ctx, leaveData)

							systemLogView.Write([]byte("Shutting down Blue Otter...\n"))

							app.Stop()
							close(quitCh)
							cancel()
							return
						case "/help":
							systemLogView.Write([]byte("Available commands:\n"))
							systemLogView.Write([]byte("/quit - Exit the chat\n"))
							systemLogView.Write([]byte("/help - Show this help message\n"))
							systemLogView.Write([]byte("/list - List all connected peers\n"))
							systemLogView.Write([]byte("/clear - Clear the chat window\n"))
							systemLogView.Write([]byte("/clear-log - Clear the system log window\n"))
							systemLogView.Write([]byte("/clear-all - Clear both chat and system log windows\n"))
							return
						case "/list":
							// List all connected peers
							peers := host.Peerstore().Peers()
							systemLogView.Write([]byte("Connected peers:\n"))
							for _, peer := range peers {
								systemLogView.Write([]byte(fmt.Sprintf("- %s\n", peer.String())))
							}
						case "/clear":
							// Clear the chat window
							chatView.SetText("")
							systemLogView.Write([]byte("Chat window cleared.\n"))
						case "/clear-log":
							// Clear the system log window
							systemLogView.SetText("")
							chatView.Write([]byte("System log window cleared.\n"))
						case "/clear-all":
							// Clear both chat and system log windows
							chatView.SetText("")
							systemLogView.SetText("")
							systemLogView.Write([]byte("Both chat and system log windows cleared.\n"))
						default:
							// Handle other commands or messages
							if strings.HasPrefix(text, "/") {
								systemLogView.Write([]byte(fmt.Sprintf("Unknown command: %s\n", text)))
								return
							}

							msg := common.ChatMessage{Sender: c.String("username"), Text: text}
							data, err := json.Marshal(msg)
							if err != nil {
								systemLogView.Write([]byte(fmt.Sprintf("Error encoding message: %s\n", err)))
								return
							}

							topic.Publish(ctx, data)
						}
					})

					app.SetFocus(inputField)
					chatView.SetChangedFunc(func() {
						app.QueueUpdateDraw(func() {
							chatView.ScrollToEnd()
						})
					})
					systemLogView.SetChangedFunc(func() {
						app.QueueUpdateDraw(func() {
							systemLogView.ScrollToEnd()
						})
					})

					// Start the TUI application
					if err := app.SetRoot(layout, true).Run(); err != nil {
						panic(err)
					}

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
			{
				Name:    "bootstrap",
				Aliases: []string{"b"},
				Usage:   "Run as a bootstrap node for other Blue Otter instances",
				Action: func(c *cli.Context) error {
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
					host, err := bootstrap.StartBootstrapNode(ctx, c.String("port"), quitCh)
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
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "Port to run the bootstrap node on",
					},
				},
			},
			{
				Name:    "add-bootstrap",
				Aliases: []string{"ab"},
				Usage:   "Add a bootstrap node address to the configuration",
				Action: func(c *cli.Context) error {
					if c.String("address") == "" {
						return fmt.Errorf("no bootstrap address specified. use --address or -a flag")
					}

					address := c.String("address")

					// Account for windows paths in powershell
					normalizedAddr := address
					if strings.HasPrefix(address, "C:/") {
						if idx := strings.Index(address, "/ip4"); idx != -1 {
							normalizedAddr = address[idx:]
						}
					}

					if err := management.AddBootstrapAddress(normalizedAddr); err != nil {
						return fmt.Errorf("failed to add bootstrap address: %w", err)
					}

					fmt.Printf("Bootstrap address '%s' added successfully\n", normalizedAddr)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "address",
						Aliases: []string{"a"},
						Usage:   "Bootstrap node address to add (e.g. /ip4/127.0.0.1/tcp/42069/p2p/QmHashValue)",
					},
				},
			},
			{
				Name:    "remove-bootstrap",
				Aliases: []string{"rb"},
				Usage:   "Remove a bootstrap node address from the configuration",
				Action: func(c *cli.Context) error {
					if c.String("address") == "" {
						return fmt.Errorf("no bootstrap address specified. use --address or -a flag")
					}

					address := c.String("address")
					if err := management.RemoveBootstrapAddress(address); err != nil {
						return fmt.Errorf("failed to remove bootstrap address: %w", err)
					}

					fmt.Printf("Bootstrap address '%s' removed successfully\n", address)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "address",
						Aliases: []string{"a"},
						Usage:   "Bootstrap node address to remove",
					},
				},
			},
			{
				Name:    "list-bootstrap",
				Aliases: []string{"lb"},
				Usage:   "List all saved bootstrap node addresses",
				Action: func(c *cli.Context) error {
					info, err := management.LoadBootstrapAddresses()
					if err != nil {
						return fmt.Errorf("failed to load bootstrap addresses: %w", err)
					}

					if len(info.Addresses) == 0 {
						fmt.Println("No bootstrap addresses saved")
						return nil
					}

					fmt.Println("Saved bootstrap addresses:")
					for i, addr := range info.Addresses {
						fmt.Printf("%d. %s\n", i+1, addr)
					}

					return nil
				},
			},
			{
				Name:    "clean-up",
				Aliases: []string{"cu"},
				Usage:   "Clean up the Blue Otter configuration directory",
				Action: func(c *cli.Context) error {
					if !c.Bool("force") {
						fmt.Println("This will delete all Blue Otter configuration data. Are you sure? (y/n)")
						reader := bufio.NewReader(os.Stdin)
						response, err := reader.ReadString('\n')
						if err != nil {
							return fmt.Errorf("error reading response: %w", err)
						}

						response = strings.TrimSpace(strings.ToLower(response))
						if response != "y" && response != "yes" {
							fmt.Println("Operation cancelled")
							return nil
						}
					}

					if err := management.CleanupConfig(); err != nil {
						return fmt.Errorf("failed to clean up configuration: %w", err)
					}

					fmt.Println("Blue Otter configuration directory cleaned up successfully")
					return nil
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force cleanup without confirmation",
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
