package rate

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List with a specific rating in the store (default unrated)",
	Args:  cobra.ExactArgs(0),
	RunE:  runListCmd,
}

func init() {
	rateCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("rating", "r", "unrated", "filter by rating: unrated, experimental, stable, edge, deprecated, broken, denied")
}

func runListCmd(cmd *cobra.Command, args []string) error {
	inputRating, _ := cmd.Flags().GetString("rating")
	snapRating, err := gql.SnapRatingByName(cmd.Context(), inputRating)
	if err != nil {
		return err
	}

	snaps, err := gql.SnapsByRating(cmd.Context(), snapRating.Id)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, snaps, customSnapListFormatter)
}

func customSnapListFormatter(rating *gql.SnapRating) (string, error) {
	table := format.NewTable(map[string]string{
		"revision":      "Revision",
		"version":       "Version",
		"alias":         "Alias",
		"description":   "Description",
		"developerNote": "Developer Note",
		"uploaded":      "Uploaded",
		"architecture":  "Architecture",
		"rating":        "Rating", // Todo: Controversial, but I think it feels better to have the rating in the table
	})

	for _, revision := range rating.SnapRevisions {
		arch := revision.SnapDeclaration.SnapDeviceArchitecture
		table.AddRow(map[string]string{
			"revision":      fmt.Sprintf("%d", revision.Revision),
			"version":       revision.SnapVersion,
			"alias":         revision.SnapDeclaration.SnapName,
			"developerNote": revision.UploadMessage,
			"description":   revision.Description,
			"uploaded":      string(revision.CreatedAt), // Todo: parse and reformat time?
			"architecture":  arch,
			"rating":        rating.Name,
		})
	}
	table = table.Sort("revision:nasc")
	return table.String(), nil
}
