package main

import (
	"context"
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
					ctx := context.Background()
					blue_otter_connect.StartServer(ctx, c.String("room"), c.String("port"))

					return nil
				},
				Flags: []cli.Flag{
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
