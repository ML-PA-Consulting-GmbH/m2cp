package store

import (
	"errors"
	"github.com/spf13/cobra"
	"m2cpcli/backend"
	"m2cpcli/format"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Report version of the store",
	RunE:  runVersionCmd,
}

func init() {
	StoreCmd.AddCommand(versionCmd)
}

func runVersionCmd(cmd *cobra.Command, args []string) error {
	res, err := backend.BackendInfo(cmd.Context())
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("unexpected empty result")
	}

	return format.PrintFormattedOutput(cmd, res.Version.Version, nil)
}

// TODO: how about a UTC timestamp from the server? We had that before.
