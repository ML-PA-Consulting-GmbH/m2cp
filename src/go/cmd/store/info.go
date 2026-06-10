package store

import (
	"errors"
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Report version and current state of the store",
	RunE:  runInfoCmd,
}

func init() {
	StoreCmd.AddCommand(infoCmd)
}

func runInfoCmd(cmd *cobra.Command, args []string) error {
	res, err := backend.BackendInfo(cmd.Context())
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("unexpected empty result")
	}

	return format.PrintFormattedOutput(cmd, res, nil)
}
