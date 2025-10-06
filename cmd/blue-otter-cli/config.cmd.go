package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"

	management "github.com/patrickma6199/blue-otter/internal/blueottermanagement"
	"github.com/urfave/cli/v2"
)

func addBootstrapCmd(c *cli.Context) error {
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
}

func removeBootstrapCmd(c *cli.Context) error {
	if c.String("address") == "" {
		return fmt.Errorf("no bootstrap address specified. use --address or -a flag")
	}

	address := c.String("address")
	if err := management.RemoveBootstrapAddress(address); err != nil {
		return fmt.Errorf("failed to remove bootstrap address: %w", err)
	}

	fmt.Printf("Bootstrap address '%s' removed successfully\n", address)
	return nil
}

func listBootstrapCmd(c *cli.Context) error {
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
}

func cleanupConfig(c *cli.Context) error {
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
}