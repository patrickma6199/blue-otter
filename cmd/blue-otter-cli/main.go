package main

import (
	"fmt"
	"os"

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
				Usage:   "Start the Blue Otter client service",
				Action:  clientCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "username",
						Aliases: []string{"u"},
						Usage:   "Username to display in chat",
					},
					&cli.StringFlag{
						Name:    "room",
						Aliases: []string{"r"},
						Usage:   "Chat rooms are distinguished by room name",
					},
					&cli.StringFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "Port to run the Blue Otter service on on your device",
					},
				},
			},
			{
				Name:    "bootstrap",
				Aliases: []string{"b"},
				Usage:   "Run as a bootstrap node for other Blue Otter instances",
				Action: bootstrapCmd,
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
				Action: addBootstrapCmd,
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
				Action: removeBootstrapCmd,
			},
			{
				Name:    "list-bootstrap",
				Aliases: []string{"lb"},
				Usage:   "List all saved bootstrap node addresses",
				Action: listBootstrapCmd,
			},
			{
				Name:    "clean-up",
				Aliases: []string{"cu"},
				Usage:   "Clean up the Blue Otter configuration directory",
				Action: cleanupConfig,
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
