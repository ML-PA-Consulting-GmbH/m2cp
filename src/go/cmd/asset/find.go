package asset

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"

	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find <hw-serial|asset-id|asset-serial|asset-name>",
	Short: "find an asset - give a hw-serial, asset-id, asset-serial or asset-name and query for a matching asset",
	Args:  cobra.ExactArgs(1),
	RunE:  runFindCmd,
}

func init() {
	AssetCmd.AddCommand(findCmd)
}

func runFindCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid number of arguments. Expected 1, got %d", len(args))
	}
	q := args[0]

	assets, err := backend.FindAsset(cmd.Context(), q)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, assets, findFormatter)
}

func findFormatter(assets []structs.Asset) (string, error) {
	if len(assets) == 0 {
		return "not matching assets found", nil
	}

	out := ""
	for _, asset := range assets {
		out += FormatAsset(&asset, nil)
	}
	return out, nil
}
