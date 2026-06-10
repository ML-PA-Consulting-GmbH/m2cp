package snap

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
)

var removeCmd = &cobra.Command{
	Use:   "remove <dgroupId | dgroupName> <co-admin userId | email>",
	Short: "Remove a co-administrators from a dgroup. Only the dgroup's admin and Super Admin can remove co-administrators.",
	Args:  cobra.ExactArgs(2),
	RunE:  runRemoveCmd,
}

func init() {
	fleetAdminCmd.AddCommand(removeCmd)
}

func runRemoveCmd(cmd *cobra.Command, args []string) (err error) {
	fleetId, err := gql.FleetIdByNameOrId(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	userId, err := gql.UserIdByEmailOrId(cmd.Context(), args[1])
	if err != nil {
		return err
	}

	// we need to find out the id of the fleet-admin relationship first
	// TODO: replace with generated code calls!
	query, err := gql.RenderQueryOrMutation(queryAdminTemplate, nil)
	if err != nil {
		return err
	}
	client, req := gql.PrepareClientAndRequest(cmd.Context(), query)
	req.Var("fleetId", fleetId)
	req.Var("userId", userId)

	var resFetchId struct {
		FleetAdministrators struct {
			Items []struct {
				Id gql.UUID `json:"id"`
			}
		} `json:"fleetAdministrators"`
	}
	err = client.Run(cmd.Context(), req, &resFetchId)
	if err != nil {
		return err
	}
	if len(resFetchId.FleetAdministrators.Items) == 0 {
		return fmt.Errorf("co-administrator not found")
	}
	relationId := resFetchId.FleetAdministrators.Items[0].Id

	// now we can remove the relationship
	query, err = gql.RenderQueryOrMutation(removeAdminTemplate, nil)
	if err != nil {
		return err
	}
	client, req = gql.PrepareClientAndRequest(cmd.Context(), query)
	req.Var("fleetAdminId", relationId)

	// {"data":{"deleteFleetAdministrators":[{"id":"80cd7b25-b635-4395-371e-08dc8c4c92cf"}]}}
	var resDelete struct {
		FleetAdministrators []struct {
			Id gql.UUID `json:"id"`
		} `json:"deleteFleetAdministrators"`
	}
	err = client.Run(cmd.Context(), req, &resDelete)
	if err != nil {
		return err
	}
	if len(resDelete.FleetAdministrators) == 0 {
		return fmt.Errorf("failed to remove co-administrator")
	}
	if resDelete.FleetAdministrators[0].Id != relationId {
		panic("unexpected result - id mismatch")
	}

	// add some detail to the result to have nicer output
	result := removeAdminResult{
		Id:      relationId,
		UserId:  userId,
		FleetId: fleetId,
	}

	return format.PrintFormattedOutput(cmd, result, customFleetAdminRemoveFormatter)
}

type removeAdminResult struct {
	Id      gql.UUID `json:"id"`
	FleetId gql.UUID `json:"fleetId"`
	UserId  gql.UUID `json:"userId"`
}

const queryAdminTemplate = `query($fleetId: UUID!, $userId: UUID!) {
	  fleetAdministrators(where: {fleetId: {eq: $fleetId}, userId: {eq: $userId}}) {
	    items {
			id
		}
	  }
	}`

const removeAdminTemplate = `mutation($fleetAdminId: UUID!) {
  deleteFleetAdministrators(ids: [$fleetAdminId]) {
    id
  }
}`

func customFleetAdminRemoveFormatter(res removeAdminResult) (string, error) {
	if res.Id == "" {
		return "failed to remove co-administrator", nil
	} else {
		return "co-administrator removed", nil
	}
}
