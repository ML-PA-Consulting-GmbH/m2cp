package model

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/store/system"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage model assertions",
}

func init() {
	system.SystemCmd.AddCommand(modelCmd)
}
