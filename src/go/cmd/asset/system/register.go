package system

import (
	"encoding/json"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "for installation-server to register a complete system with a single command. For manual registration please use 'asset create', 'asset system create', 'asset system add' commands",
	Args:  cobra.ExactArgs(1),
	RunE:  runRegisterCmd,
}

func init() {
	AssetSystemCmd.AddCommand(registerCmd)
}

type registration struct {
	System  systemSpec   `json:"system"`
	Devices []deviceSpec `json:"devices"`
}

type systemSpec struct {
	TenantAlias    string `json:"tenantAlias"`
	AssetName      string `json:"assetName"`
	AssetModelName string `json:"assetModelName"`
}

type deviceSpec struct {
	Type       string `json:"type"`
	HwSerial   string `json:"hwSerial"`
	HwMcuId    string `json:"hwMcuId,omitempty"`
	TpmPubKey  string `json:"tpmPubKey,omitempty"`
	HwModel    string `json:"hwModel"`
	HwRevision int    `json:"hwRevision,omitempty"`
}

func runRegisterCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid number of arguments. Expected 1, got %d", len(args))
	}

	var registrationInput registration

	jsonString := args[0]
	if err := json.Unmarshal([]byte(jsonString), &registrationInput); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	mutationInput := structs.SystemAssetProvisionInput{
		TenantAlias:          registrationInput.System.TenantAlias,
		SystemAssetModelName: registrationInput.System.AssetModelName,
		SystemAssetName:      &registrationInput.System.AssetName,
	}

	mutationInput.Devices = []*structs.SystemAssetProvisionInputDevice{}

	for _, device := range registrationInput.Devices {
		deviceInput := structs.SystemAssetProvisionInputDevice{
			Type:               device.Type,
			AssetModelName:     device.HwModel,
			SerialNumber:       device.HwSerial,
			McuId:              &device.HwMcuId,
			AssetModelRevision: &device.HwRevision,
			AttestationKey:     &device.TpmPubKey,
		}
		mutationInput.Devices = append(mutationInput.Devices, &deviceInput)
	}

	response, err := backend.ProvisionSystemAsset(cmd.Context(), mutationInput)
	if err != nil {
		jsonBytes, e := json.MarshalIndent(err, "", "  ")
		if e != nil {
			return err
		} else {
			return fmt.Errorf("%s", string(jsonBytes))
		}
	}

	return format.PrintFormattedOutput(cmd, response, nil)
}
