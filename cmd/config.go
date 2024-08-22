package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func InitConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory %v", err)
	}

	configFile := filepath.Join(homeDir, ".pcli", "config.yaml")
	viper.SetConfigFile(configFile)

	// default values for Harmony
	viper.SetDefault("RPC_URL", "https://api.s0.t.hmny.io")
	viper.SetDefault("UniswapV3Factory", "0x12d21f5d0Ab768c312E19653Bf3f89917866B8e8")
	viper.SetDefault("SemiFungiblePositionManager", "0x8189647419E5d72CBd29565586D048A248C07fEb")
	viper.SetDefault("PanopticFactory", "0x4F208393a3705289cd4068f4e76F1058Fc15e590")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}
