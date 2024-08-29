package option

import (
	"github.com/spf13/cobra"
)

var OptionCmd = &cobra.Command{
	Use:   "option",
	Short: "Manage option",
}

func init() {
	// OptionCmd.AddCommand(mintCmd)
}
