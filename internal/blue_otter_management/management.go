package blue_otter_management

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	common "github.com/patrickma6199/blue-otter/internal/blue_otter_common"
)

// GetConfigDir returns the path to the Blue Otter configuration directory
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".blue-otter")
	return configDir, nil
}

// GetBootstrapFilePath returns the path to the bootstrap.json file
func GetBootstrapFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "bootstrap.json"), nil
}

// EnsureConfigDir ensures the config directory exists
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return os.MkdirAll(configDir, 0755)
	}
	return nil
}

// SaveBootstrapInfo saves the bootstrap node's information to a file
func SaveBootstrapInfo(host host.Host) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	oldInfo, err := LoadBootstrapAddresses()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".blue-otter")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Get private key
	privateKeyData, err := crypto.MarshalPrivateKey(host.Peerstore().PrivKey(host.ID()))
	if err != nil {
		return fmt.Errorf("failed to get private key: %w", err)
	}

	// Encode private key as base64
	encodedPrivateKey := base64.StdEncoding.EncodeToString(privateKeyData)

	// Create array of full multiaddresses including peer ID
	var addresses []string
	for _, addr := range host.Addrs() {
		fullAddr := fmt.Sprintf("%s/p2p/%s", addr.String(), host.ID())
		addresses = append(addresses, fullAddr)
	}

	// Create bootstrap info
	info := common.BootstrapInfo{
		BootStrapNodeAddresses: addresses,
		Addresses:  oldInfo.Addresses,
		PrivateKey: encodedPrivateKey,
		PeerID:     host.ID().String(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bootstrap info: %w", err)
	}

	// Write to file
	filePath := filepath.Join(configDir, "bootstrap.json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write bootstrap info: %w", err)
	}

	fmt.Printf("Bootstrap node info saved to %s\n", filePath)
	fmt.Println("Share this file with other users to allow them to connect to this bootstrap node")
	return nil
}

// GetBootstrapPrivateKey retrieves the private key of the bootstrap node from config
func GetBootstrapPrivateKey() (crypto.PrivKey, error) {
	configPath, err := GetBootstrapFilePath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return error
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var info common.BootstrapInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	// If no private key in config, return nil
	if info.PrivateKey == "" {
		return nil, nil
	}

	// Decode private key from base64
	privateKeyData, err := base64.StdEncoding.DecodeString(info.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Unmarshal private key
	privateKey, err := crypto.UnmarshalPrivateKey(privateKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}

	return privateKey, nil
}

// LoadBootstrapAddressesForConnections loads bootstrap addresses from the config file
func LoadBootstrapAddressesForConnections() ([]string, error) {

	info, err := LoadBootstrapAddresses()
	if err != nil {
		return nil, err
	}

	return info.Addresses, nil
}

// LoadBootstrapAddresses loads the bootstrap addresses from the configuration file
func LoadBootstrapAddresses() (common.BootstrapInfo, error) {
	var info common.BootstrapInfo

	configPath, err := GetBootstrapFilePath()
	if err != nil {
		return info, err
	}

	// If file doesn't exist, return empty info
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return common.BootstrapInfo{Addresses: []string{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return info, err
	}

	if err := json.Unmarshal(data, &info); err != nil {
		return info, err
	}

	return info, nil
}

// SaveBootstrapAddress saves the bootstrap addresses to the configuration file
func SaveBootstrapAddress(info common.BootstrapInfo) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetBootstrapFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// AddBootstrapAddress adds a new bootstrap address to the configuration
func AddBootstrapAddress(address string) error {
	info, err := LoadBootstrapAddresses()
	if err != nil {
		return err
	}

	// Check if the address already exists
	for _, addr :=range info.Addresses {
		if addr == address {
			return errors.New("bootstrap address already exists")
		}
	}

	// Add the new address
	info.Addresses = append(info.Addresses, address)
	return SaveBootstrapAddress(info)
}

// RemoveBootstrapAddress removes a bootstrap address from the configuration
func RemoveBootstrapAddress(address string) error {
	info, err := LoadBootstrapAddresses()
	if err != nil {
		return err
	}

	// Find and remove the address
	found := false
	var newAddresses []string
	for _, addr := range info.Addresses {
		if addr != address {
			newAddresses = append(newAddresses, addr)
		} else {
			found = true
		}
	}

	if !found {
		return errors.New("bootstrap address not found")
	}

	info.Addresses = newAddresses
	return SaveBootstrapAddress(info)
}

// CleanupConfig removes the Blue Otter configuration directory
func CleanupConfig() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Check if directory exists before attempting to remove
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return nil
	}

	return os.RemoveAll(configDir)
}
