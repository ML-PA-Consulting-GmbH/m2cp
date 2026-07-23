package model

import (
	"github.com/spf13/cobra"
	"m2cpcli/backend"
	"m2cpcli/format"
	"strconv"
)

var modifyCmd = &cobra.Command{
	Use:   "modify [modelId | modelType modelName modelRevision]",
	Short: "Modify a model in your store",
	Args:  validateModelModifyArgs,
	RunE:  runModifyCmd,
}

func validateModelModifyArgs(cmd *cobra.Command, args []string) error {
	return validateModelArgs(cmd, args)
}

func init() {
	ModelCmd.AddCommand(modifyCmd)
	modifyCmd.Flags().String("tpm", "false", "if the model requires a Trusted Platform Module (TPM) (true or false)")
	modifyCmd.Flags().String("pre-reg", "false", "if the model requires hardware pre-registration (true or false)")
	modifyCmd.Flags().String("provisioning-claim", "false", "if the model requires a device provisioning claim (true or false); requires backend >= 5.2.0")
}

type ModelModifyResult struct {
	Response *backend.DeviceModelRevisionUpdateResult
}

func runModifyCmd(cmd *cobra.Command, args []string) error {
	modelId, err := modelIdFromArgs(cmd, args)
	deviceModelRevisionId := string(modelId)

	var isTpmRequired *bool
	var isPreRegistrationRequired *bool
	var isDeviceProvisioningClaimRequired *bool

	if cmd.Flags().Changed("tpm") {
		var tpm string
		tpm, err = cmd.Flags().GetString("tpm")
		if err != nil {
			return err
		}

		value, err := strconv.ParseBool(tpm)
		if err != nil {
			return err
		}

		isTpmRequired = &value
	}

	if cmd.Flags().Changed("pre-reg") {
		var preReg string
		preReg, err = cmd.Flags().GetString("pre-reg")
		if err != nil {
			return err
		}

		value, err := strconv.ParseBool(preReg)
		if err != nil {
			return err
		}

		isPreRegistrationRequired = &value
	}

	if cmd.Flags().Changed("provisioning-claim") {
		var provisioningClaim string
		provisioningClaim, err = cmd.Flags().GetString("provisioning-claim")
		if err != nil {
			return err
		}

		value, err := strconv.ParseBool(provisioningClaim)
		if err != nil {
			return err
		}

		isDeviceProvisioningClaimRequired = &value
	}

	result, err := backend.UpdateDeviceModelRevision(cmd.Context(), deviceModelRevisionId, isTpmRequired, isPreRegistrationRequired, isDeviceProvisioningClaimRequired)
	if err != nil {
		return err
	}

	msg := ModelModifyResult{
		Response: result,
	}

	return format.PrintFormattedOutput(cmd, msg, customModelModifyFormatter)
}

func customModelModifyFormatter(_ ModelModifyResult) (string, error) {
	return "Success", nil
}
