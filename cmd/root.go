package cmd

import (
	"pcli/cmd/key"
	"pcli/cmd/option"
	"pcli/cmd/pool"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "panoptics",
	Short: "Interact with the Panoptics protocol",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// private key management
	rootCmd.AddCommand(key.KeyCmd)

	// pool management
	rootCmd.AddCommand(pool.PoolCmd)

	// TODO: mint options
	rootCmd.AddCommand(option.OptionCmd)

	// TODO: burn options

	// initialize configs
	cobra.OnInitialize(InitConfig)
}
