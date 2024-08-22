package pool

import (
	"github.com/spf13/cobra"
)

var PoolCmd = &cobra.Command{
	Use:   "pool",
	Short: "Manage liquidity pools",
}

func init() {
	PoolCmd.AddCommand(deployCmd)
	PoolCmd.AddCommand(getCmd)
}
