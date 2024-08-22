package key

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View private key",
	Run: func(cmd *cobra.Command, args []string) {
		view()
	},
}

func view() {
	privateKey := viper.GetString("private_key")
	if privateKey == "" {
		fmt.Println("No private key found.")
		return
	}

	fmt.Printf("%s\n", privateKey)
}
