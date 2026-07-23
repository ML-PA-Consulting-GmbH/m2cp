package device

import (
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var claimCmd = &cobra.Command{
	Use:   "claim <claimingToken>",
	Short: "Claim a device for the store using a claiming token",
	Long: "Claim a device for the current store by redeeming a claiming token.\n" +
		"Claiming tokens are primarily a devkit mechanism; once redeemed the device is\n" +
		"claimed by the tenant and appears in the devices list.",
	Args: cobra.ExactArgs(1),
	RunE: runClaimCmd,
}

type claimResult struct {
	ClaimId    string `json:"claimId"`
	DeviceId   string `json:"deviceId,omitempty"`
	ExpiresAt  string `json:"expiresAt,omitempty"`
	ConsumedAt string `json:"consumedAt,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

func init() {
	DeviceCmd.AddCommand(claimCmd)
}

func runClaimCmd(cmd *cobra.Command, args []string) error {
	claim, err := backend.ClaimDevice(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	result := claimResult{
		ClaimId:    claim.ClaimId,
		DeviceId:   claim.DeviceId,
		ExpiresAt:  claim.ExpiresAt,
		ConsumedAt: claim.ConsumedAt,
		CreatedAt:  claim.CreatedAt,
	}

	return format.PrintFormattedOutput(cmd, result, customClaimFormatter)
}

func customClaimFormatter(res claimResult) (string, error) {
	list := format.NewList()
	list.Add("Message", "Device claimed successfully")
	list.Add("Claim Id", res.ClaimId)
	if res.DeviceId != "" {
		list.Add("Device Id", res.DeviceId)
	}
	if res.ExpiresAt != "" {
		list.Add("Expires At", res.ExpiresAt)
	}
	return list.String(), nil
}
