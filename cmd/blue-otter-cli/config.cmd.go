package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
	"regexp"
	"strconv"

	management "github.com/patrickma6199/blue-otter/internal/blueottermanagement"
	"github.com/urfave/cli/v2"
)

// addBootstrapCmd is the main function run to facilitate the functionality of the add-bootstrap command
func addBootstrapCmd(c *cli.Context) error {
	if c.String("address") == "" {
		return fmt.Errorf("no bootstrap address specified. use --address or -a flag")
	}

	address := c.String("address")

	// Account for windows paths in powershell
	sanitizedAddr := sanitizeAddress(address)

	if !isAddressValid(sanitizedAddr) {
		return fmt.Errorf("invalid bootstrap address format")
	}

	if err := management.AddBootstrapAddress(sanitizedAddr); err != nil {
		return fmt.Errorf("failed to add bootstrap address: %w", err)
	}

	fmt.Printf("Bootstrap address '%s' added successfully\n", sanitizedAddr)
	return nil
}

// removeBootstrapCmd is the main function run to facilitate the functionality of the remove-bootstrap command
func removeBootstrapCmd(c *cli.Context) error {
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

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter the number of the address you wish to remove: ")
	if scanner.Scan() {
	input := scanner.Text()

	value, err := strconv.Atoi(input);
	if err != nil {
		return fmt.Errorf("please enter the addresses number from the list")
	}
	if value < 1 || value > len(info.Addresses) {
		return fmt.Errorf("please enter a number within the range of the list")
	}

	if err := management.RemoveBootstrapAddress(info.Addresses[value-1]); err != nil {
		return fmt.Errorf("failed to remove bootstrap address: %w", err)
	}
	
	fmt.Printf("Bootstrap address '%s' removed successfully\n", info.Addresses[value-1])
	}

	return nil
}

// listBootstrapCmd is the main function run to facilitate the functionality of the list-bootstrap command
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

// cleanupConfig is the main function run to facilitate the functionality of the clean-up command
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

// --------------- Helper Functions ---------------

// isAddressValid checks if an address is a valid form for libp2p peers
func isAddressValid(address string) bool {
	var multiaddrRegex = regexp.MustCompile(`^/ip4/(\d{1,3}\.){3}\d{1,3}/tcp/\d+/p2p/[A-Za-z0-9]+$`)
	return multiaddrRegex.MatchString(address)
}

// sanitizeAddress is to protect against windows command-line address auto-resolution
func sanitizeAddress(address string) string {
	sanitizedAddr := address
	if strings.HasPrefix(address, "C:/") {
		if idx := strings.Index(address, "/ip4"); idx != -1 {
			sanitizedAddr = address[idx:]
		}
	}
	return sanitizedAddr
}
