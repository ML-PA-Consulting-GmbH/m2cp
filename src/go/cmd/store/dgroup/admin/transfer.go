package snap

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
)

var transferCmd = &cobra.Command{
	Use:   "transfer <dgroupId | dgroupName> <admin userId | email>",
	Short: "Transfer administrator privilege from current user to a new user. Only the dgroup's admin and Super Admin can do this.",
	Args:  cobra.ExactArgs(2),
	RunE:  runTransferCmd,
}

func init() {
	fleetAdminCmd.AddCommand(transferCmd)
}

func runTransferCmd(cmd *cobra.Command, args []string) error {
	fleetId, err := gql.FleetIdByNameOrId(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	userId, err := gql.UserIdByEmailOrId(cmd.Context(), args[1])
	if err != nil {
		return err
	}

	// TODO: replace with generated code call!
	query, err := gql.RenderQueryOrMutation(transferAdminTemplate, nil)
	if err != nil {
		return err
	}
	client, req := gql.PrepareClientAndRequest(cmd.Context(), query)
	req.Var("fleetId", fleetId)
	req.Var("userId", userId)

	// {"data":{"updateFleets":[{"id":"8b2d67a0-caea-4ce9-045a-08dc6f6964e6","ownerUserId":"839fb7ff-1e13-452d-8e0d-df076c2b0b6f"}]}}
	var resTransfer struct {
		UpdateFleets []struct {
			FleetId     gql.UUID `json:"id"`
			OwnerUserId gql.UUID `json:"ownerUserId"`
		} `json:"updateFleets"`
	}
	err = client.Run(cmd.Context(), req, &resTransfer)
	if err != nil {
		return err
	}
	if len(resTransfer.UpdateFleets) == 0 {
		return fmt.Errorf("failed to transfer administrator privilege")
	}
	if fleetId != resTransfer.UpdateFleets[0].FleetId {
		panic("fleetId mismatch")
	}
	if userId != resTransfer.UpdateFleets[0].OwnerUserId {
		panic("userId mismatch")
	}

	result := transferAdminResult{
		FleetId: fleetId,
		UserId:  userId,
	}

	return format.PrintFormattedOutput(cmd, result, customFleetAdminTransferFormatter)
}

type transferAdminResult struct {
	FleetId gql.UUID `json:"id"`
	UserId  gql.UUID `json:"userId"`
}

const transferAdminTemplate = `mutation($fleetId: UUID!, $userId: UUID!) {
  updateFleets(fleets: [{
      id: $fleetId
      ownerUserId: $userId
    }
  ]) {
    id
    ownerUserId
  }
}
`

func customFleetAdminTransferFormatter(res transferAdminResult) (string, error) {
	if res.FleetId == "" {
		return "failed to transfer administrator privilege", nil
	} else {
		return "administrator privilege transferred", nil
	}
}
