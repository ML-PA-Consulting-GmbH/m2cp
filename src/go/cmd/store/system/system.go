package system

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/store"
)

var SystemCmd = &cobra.Command{
	Use:    "system",
	Short:  "Advanced administration commands",
	Hidden: true,
}

func init() {
	store.StoreCmd.AddCommand(SystemCmd)
}
