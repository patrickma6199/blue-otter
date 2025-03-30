package management

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// BootstrapInfo represents the bootstrap node information
type BootstrapInfo struct {
	Addresses []string `json:"addresses"`
}

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

// LoadBootstrapAddresses loads the bootstrap addresses from the configuration file
func LoadBootstrapAddresses() (BootstrapInfo, error) {
	var info BootstrapInfo
	
	configPath, err := GetBootstrapFilePath()
	if err != nil {
		return info, err
	}

	// If file doesn't exist, return empty info
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return BootstrapInfo{Addresses: []string{}}, nil
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

// SaveBootstrapAddresses saves the bootstrap addresses to the configuration file
func SaveBootstrapAddresses(info BootstrapInfo) error {
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
	for _, addr := range info.Addresses {
		if addr == address {
			return errors.New("bootstrap address already exists")
		}
	}

	// Add the new address
	info.Addresses = append(info.Addresses, address)
	return SaveBootstrapAddresses(info)
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
	return SaveBootstrapAddresses(info)
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
