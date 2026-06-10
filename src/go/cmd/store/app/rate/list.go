package rate

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Apps with a specific rating (default unrated)",
	Args:  cobra.ExactArgs(0),
	RunE:  runListCmd,
}

func init() {
	rateCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("rating", "r", "unrated", "filter by rating: unrated, experimental, stable, edge, deprecated, broken, denied")
}

func runListCmd(cmd *cobra.Command, args []string) error {
	inputRating, _ := cmd.Flags().GetString("rating")

	if !slices.Contains([]string{"unrated", "experimental", "stable", "edge", "deprecated", "broken", "denied"}, inputRating) {
		return fmt.Errorf("invalid rating: '%s'", inputRating)
	}

	res, err := backend.GetAppsByRating(cmd.Context(), inputRating)
	if err != nil {
		return err
	}

	if res == nil || res.AppRevisions == nil || res.AppRevisions.Items == nil || len(res.AppRevisions.Items) == 0 {
		return fmt.Errorf("no matching App Revisions found")
	}

	appRevisions := []structs.AppRevision{}
	for _, appRevision := range res.AppRevisions.Items {
		createdAt, _ := time.Parse(time.RFC3339, appRevision.CreatedAt)
		appRevisions = append(appRevisions, structs.AppRevision{
			Revision: appRevision.Revision,
			Version:  appRevision.Version,
			App: structs.App{
				Name:         appRevision.App.AppName,
				Architecture: string(appRevision.App.Architecture),
			},
			Description:   appRevision.Description,
			UploadMessage: appRevision.UploadMessage,
			Rating:        appRevision.AppStatus.Name,
			CreatedAt:     createdAt,
			CreatedBy: structs.User{
				Id: tools.MaybeStringToString(appRevision.CreatedBy, ""),
			},
		})
	}

	return format.PrintFormattedOutput(cmd, appRevisions, customSnapListFormatter)
}

func customSnapListFormatter(appRevisions []structs.AppRevision) (string, error) {
	table := format.NewTable(map[string]string{
		"app":           "App",
		"version":       "Version",
		"description":   "Description",
		"uploadMessage": "Developer Note",
		"uploaded":      "Uploaded",
		"uploader":      "Uploader",
		"architecture":  "Architecture",
		"rating":        "Rating", // Todo: Controversial, but I think it feels better to have the rating in the table
	})

	for _, appRevision := range appRevisions {
		table.AddRow(map[string]string{
			"app":           appRevision.App.Name,
			"version":       fmt.Sprintf("%s (%d)", appRevision.Version, appRevision.Revision),
			"uploadMessage": appRevision.UploadMessage,
			"description":   tools.MaybeStringToString(appRevision.Description, "n/a"),
			"uploaded":      appRevision.CreatedAt.Format("2006-01-02 15:04"),
			"uploader":      appRevision.CreatedBy.Id,
			"architecture":  appRevision.App.Architecture,
			"rating":        appRevision.Rating,
		})
	}
	table = table.Sort("app:asc,revision:nasc")
	return table.StringSelect([]string{"app", "architecture", "version", "uploaded", "uploadMessage"}), nil
}
