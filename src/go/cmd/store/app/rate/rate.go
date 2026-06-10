package rate

import (
	"m2cpcli/cmd/store/app"

	"github.com/spf13/cobra"
)

var rateCmd = &cobra.Command{
	Use:     "rate",
	Aliases: []string{"r"},
	Short:   "Manage app ratings in the app store",
}

func init() {
	app.AppCmd.AddCommand(rateCmd)
}
