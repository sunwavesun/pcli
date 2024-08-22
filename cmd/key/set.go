package key

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the private key to be used for executing transactions",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Please provide a private key.")
			return
		}
		set(args[0])
	},
}

func set(privateKey string) {
	if !isValidKey(privateKey) {
		fmt.Println("Invalid private key format.")
		return
	}

	viper.Set("private_key", privateKey)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	configPath := filepath.Join(homeDir, ".pcli")
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		return
	}

	configFile := filepath.Join(configPath, "config.yaml")
	if err := viper.WriteConfigAs(configFile); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		return
	}

	fmt.Println("Private key saved successfully.")
}

func isValidKey(key string) bool {
	// check if the key is a 64-character hexadecimal string
	matched, err := regexp.MatchString("^[a-fA-F0-9]{64}$", key)
	if err != nil {
		return false
	}
	return matched
}
