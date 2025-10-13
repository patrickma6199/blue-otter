package main

import (
	"fmt"
	"strings"
	"context"
	"strconv"
	"encoding/json"

	tcell "github.com/gdamore/tcell/v2"
	common "github.com/patrickma6199/blue-otter/internal/blueottercommon"
	client "github.com/patrickma6199/blue-otter/internal/blueotterclient"
	tui "github.com/patrickma6199/blue-otter/internal/blueottertui"
	"github.com/urfave/cli/v2"
	"github.com/rivo/tview"
)

// clientCmd is the main function run to facilitate the functionality of the client command
func clientCmd(c *cli.Context) error {

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
}