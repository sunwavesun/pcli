package key

import (
	"github.com/spf13/cobra"
)

var KeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage private key",
	Long:  `Set, view, and delete the private key used for executing transactions.`,
}

func init() {
	KeyCmd.AddCommand(setCmd)
	KeyCmd.AddCommand(viewCmd)
	KeyCmd.AddCommand(deleteCmd)
}
