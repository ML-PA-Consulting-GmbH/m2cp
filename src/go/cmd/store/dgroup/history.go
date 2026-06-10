package dgroup

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools/console"
	"time"

	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history <deployment-group-name>",
	Short: "Retrieve the change history of a specified Deployment Group",
	Long: `Fetches the change history of a given Deployment Group by querying the log books.
The command requires a Deployment Group name and retrieves logs within a specific time range.
By default, the time range is set to the past one week. `,
	Example: `m2cp store dgroup history my-dgroup`,
	Args:    cobra.ExactArgs(1),
	RunE:    runHistoryCmd,
}

func init() {
	historyCmd.Flags().StringP("newer-than", "n", "", "only print logs newer than given duration (format: 1m, 1h)")
	historyCmd.Flags().StringP("after", "", "", "show only logs newer than this UTC date (format: 2006-01-02T15:04)")
	historyCmd.Flags().StringP("before", "", "", "show only logs older than this UTC date (format: 2006-01-02T15:04")
	FleetCmd.AddCommand(historyCmd)
	DGroupCmd.AddCommand(historyCmd)
}

type historyCmdOutput struct {
	after    string
	before   string
	Response *backend.GetDeploymentGroupLogBooksResponse
}

func runHistoryCmd(cmd *cobra.Command, args []string) error {
	deploymentGroupName := args[0]
	after, _ := cmd.Flags().GetString("after")
	before, _ := cmd.Flags().GetString("before")
	newerThan, _ := cmd.Flags().GetString("newer-than")

	if newerThan != "" && after != "" {
		return fmt.Errorf("newer-than and after cannot be set at the same time")
	}
	if newerThan != "" && before != "" {
		return fmt.Errorf("newer-than and before cannot be set at the same time")
	}

	if newerThan != "" {
		duration, err := time.ParseDuration(newerThan)
		if err != nil {
			return err
		}

		before = time.Now().Format(time.RFC3339)
		after = time.Now().Add(-duration).Format(time.RFC3339)

	} else {
		if after != "" {
			_, err := time.Parse(time.RFC3339, after)
			if err != nil {
				return err
			}
		} else {
			after = time.Now().AddDate(0, 0, -7).Format(time.RFC3339)
		}
		if before != "" {
			_, err := time.Parse(time.RFC3339, before)
			if err != nil {
				return err
			}
		} else {
			before = time.Now().Format(time.RFC3339)
		}
	}

	deploymentGroupByIdResponse, err := backend.GetDeploymentGroupIdByName(cmd.Context(), deploymentGroupName)
	if err != nil {
		return err
	}

	if len(deploymentGroupByIdResponse.DeploymentGroups.Items) == 0 {
		return fmt.Errorf("Deployment Group not found")
	}

	deploymentGroupId := deploymentGroupByIdResponse.DeploymentGroups.Items[0].Id

	response, err := backend.GetDeploymentGroupLogBooks(cmd.Context(), deploymentGroupId, after, before)
	if err != nil {
		return err
	}

	output := historyCmdOutput{
		after:    after,
		before:   before,
		Response: response,
	}

	return format.PrintFormattedOutput(cmd, output, historyOutputFormatter)
}

func historyOutputFormatter(res historyCmdOutput) (string, error) {
	outputStr := ""

	parsedAfter, _ := time.Parse(time.RFC3339, res.after)
	parsedBefore, _ := time.Parse(time.RFC3339, res.before)

	outputStr += fmt.Sprintf("Logs from %s to %s\n", parsedAfter.Format("2006-01-02 15:04:05"), parsedBefore.Format("2006-01-02 15:04:05"))

	if len(res.Response.DeploymentGroup.LogBooks) > 0 {
		tree := format.NewTree(console.Colorize(console.Green, "Log Books"))

		for _, logBook := range res.Response.DeploymentGroup.LogBooks {
			parsedTime, _ := time.Parse(time.RFC3339, logBook.CreatedAt)
			node := tree.NewChild(console.Colorize(console.Green, parsedTime.Format("2006-01-02 15:04:05")))
			node.AddLeaf("Description: " + logBook.Description)
			node.AddLeaf("User: " + logBook.User.DisplayName + " (" + logBook.User.Email + ")")
		}

		outputStr += tree.String()
	}

	return outputStr, nil
}
