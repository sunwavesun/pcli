package key

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the private key",
	Run: func(cmd *cobra.Command, args []string) {
		delete()
	},
}

func delete() {
	viper.Set("private_key", "")
	if err := viper.WriteConfig(); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		return
	}

	fmt.Println("Private key deleted successfully.")
}
