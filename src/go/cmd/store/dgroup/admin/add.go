package snap

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
)

var addCmd = &cobra.Command{
	Use:   "add <dgroupId | dgroupName> <co-admin userId | email>",
	Short: "Add a co-administrator to a dgroup. Only the dgroup admin and Super Admin can add co-administrators.",
	Args:  cobra.ExactArgs(2),
	RunE:  runAddCmd,
}

func init() {
	fleetAdminCmd.AddCommand(addCmd)
}

func runAddCmd(cmd *cobra.Command, args []string) error {
	fleetId, err := gql.FleetIdByNameOrId(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	userId, err := gql.UserIdByEmailOrId(cmd.Context(), args[1])
	if err != nil {
		return err
	}

	// TODO: replace with generated code call!
	query, err := gql.RenderQueryOrMutation(addAdminTemplate, nil)
	if err != nil {
		return err
	}
	client, req := gql.PrepareClientAndRequest(cmd.Context(), query)
	req.Var("fleetId", fleetId)
	req.Var("userId", userId)

	var result map[string][]addAdminResult
	err = client.Run(cmd.Context(), req, &result)
	if err != nil {
		return err
	}

	if resultContent, ok := result["createFleetAdministrators"]; !ok {
		return fmt.Errorf("failed to add co-administrator: unexpected result: %s", resultContent)
	} else if len(resultContent) == 0 {
		return fmt.Errorf("failed to add co-administrator: no results")
	} else {
		return format.PrintFormattedOutput(cmd, resultContent[0], customFleetAdminAddFormatter)
	}
}

// "createFleetAdministrators":[{"id":"448ac4f2-dcfb-4f34-ebf2-08dcbc42b72d","fleetId":"688f76b6-7a76-4e1b-5ead-08dc49803b84","userId":"9283a07b-ee01-4304-a547-dffe0a3e26a6"}]}}

type addAdminResult struct {
	Id      gql.UUID `json:"id"`
	FleetId gql.UUID `json:"fleetId"`
	UserId  gql.UUID `json:"userId"`
}

const addAdminTemplate = `mutation($fleetId: UUID!, $userId: UUID!) {
  createFleetAdministrators(fleetAdministrators: [
    {
      fleetId: $fleetId
      userId: $userId
    }
  ]) {
    id
    fleetId
    userId
  }
}
`

func customFleetAdminAddFormatter(res addAdminResult) (string, error) {
	if res.Id == "" {
		return "failed to add co-administrator", nil
	} else {
		return "co-administrator added", nil
	}
}
