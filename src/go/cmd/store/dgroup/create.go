package dgroup

import (
	"context"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [name] [modelId | modelType modelName modelRevision | deviceName or osSerial]",
	Short: "Create a Deployment Group from a model or device",
	Long: `Create a Deployment Group using one of three approaches:

1. From a model ID:
   m2cp store dgroup create <name> <modelId>

2. From model details:
   m2cp store dgroup create <name> <modelType> <modelName> <modelRevision>
   Valid device types: 'm2cp' or 'edge' for edge devices, 'rtd' for real-time devices

3. From an existing device:
   m2cp store dgroup create <name> <deviceName or osSerial>

When providing a UUID, the command automatically determines whether it's a device ID 
or model ID.`,
	Args: validateDeploymentGroupCreateArgs,
	RunE: runCreateCmd,
}

func init() {
	createCmd.Flags().StringP("description", "d", "", "description of the new Deployment Group")
	createCmd.Flags().StringP("auto-update-mode", "m", "off", "auto update mode (off, stable, edge)")
	FleetCmd.AddCommand(createCmd)
	DGroupCmd.AddCommand(createCmd)
}

func validateDeploymentGroupCreateArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.RangeArgs(2, 4)(cmd, args); err != nil {
		return err
	}

	switch len(args) {
	case 2:
		// name := args[0]
		// second argument could be modelId or deviceName/osSerial - we'll validate at runtime
	case 4:
		// name := args[0]
		//modelType := args[1]
		//modelName := args[2]
		modelRevision := args[3]
		if !tools.IsValidRevision(modelRevision) {
			return fmt.Errorf("invalid model revision '%s'", modelRevision)
		}
	default:
		return fmt.Errorf("invalid number of arguments")
	}

	return nil
}

type StoreDeploymentGroupCreateResult struct {
	Message                string      `json:"message"`
	CreatedDeploymentGroup interface{} `json:"createdDeploymentGroup"`
}

func runCreateCmd(cmd *cobra.Command, args []string) error {
	var deploymentGroupName string
	var err error

	deploymentGroupName = args[0]

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	autoUpdateModeName, err := cmd.Flags().GetString("auto-update-mode")
	if err != nil {
		return err
	}

	if autoUpdateModeName == "" {
		autoUpdateModeName = "off"
	}

	autoUpdateModeId, found := backend.GetAutoUpdateModeIdByName(autoUpdateModeName)
	if !found {
		return fmt.Errorf("unknown auto update mode '%s'", autoUpdateModeName)
	}

	switch len(args) {
	case 2:
		// Second argument could be modelId or deviceName/osSerial
		// Try model first, then device if model lookup fails
		secondArg := args[1]
		return createFromModelOrDevice(cmd, deploymentGroupName, secondArg, &description, &autoUpdateModeId)
	case 4:
		// Model details approach
		modelType := args[1]
		modelName := args[2]
		modelRevision, err := strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("invalid model revision number: \"%s\": %s", args[3], err)
		}

		modelRevisionId, err := backend.TryGetDeviceModelRevisionByTypeNameRevision(cmd.Context(), modelType, modelName, modelRevision)
		if err != nil {
			return err
		}

		return createFromModelId(cmd, deploymentGroupName, modelRevisionId, &description, &autoUpdateModeId)
	default:
		return fmt.Errorf("invalid number of arguments")
	}
}

func createFromModelOrDevice(cmd *cobra.Command, deploymentGroupName, identifier string, description, autoUpdateModeId *string) error {
	// Check if it's a modelId
	if tools.IsValidUuid(identifier) {
		if modelId, found := TryGetModelById(cmd.Context(), identifier); found {
			return createFromModelId(cmd, deploymentGroupName, modelId, description, autoUpdateModeId)
		}
	}

	// Check if it's a device (bys osSerial or name)
	if deviceId, found := TryGetDeviceByNameOrSerial(cmd.Context(), identifier); found {
		return createFromDeviceWithId(cmd, deploymentGroupName, deviceId, description, autoUpdateModeId)
	}

	return fmt.Errorf("'%s' is not a valid device name/ID or model ID", identifier)
}

// TryGetDeviceByNameOrSerial checks if the given identifier corresponds to an existing device and returns the device ID
func TryGetDeviceByNameOrSerial(ctx context.Context, deviceIdOrName string) (string, bool) {
	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(ctx, deviceIdOrName)
	if err != nil {
		return "", false
	}
	return deviceId, true
}

// TryGetModelById checks if the given ID corresponds to an existing model revision and returns the model revision ID
func TryGetModelById(ctx context.Context, modelId string) (string, bool) {
	return backend.TryGetDeviceModelRevisionById(ctx, modelId)
}

func createFromModelId(cmd *cobra.Command, deploymentGroupName, modelId string, description, autoUpdateModeId *string) error {
	if modelId == "" || !tools.IsValidUuid(modelId) {
		return fmt.Errorf("invalid model ID: %s", modelId)
	}

	isDeltaUpdateOnly := false

	// Create deployment group input
	input := &backend.DeploymentGroupCreateInput{
		DeviceModelRevisionId: modelId,
		Name:                  deploymentGroupName,
		Description:           description,
		AutoUpdateModeId:      autoUpdateModeId,
		IsDeltaUpdateOnly:     &isDeltaUpdateOnly,
	}

	createResult, err := backend.CreateDeploymentGroups(cmd.Context(), []*backend.DeploymentGroupCreateInput{input})
	if err != nil {
		return err
	}

	if len(createResult.CreateDeploymentGroups) != 1 {
		return fmt.Errorf("failed to create deployment group")
	}

	msg := StoreDeploymentGroupCreateResult{
		Message:                "Deployment Group created from model",
		CreatedDeploymentGroup: createResult.CreateDeploymentGroups[0],
	}

	return format.PrintFormattedOutput(cmd, msg, customStoreDeploymentGroupFormatter)
}

func createFromDeviceWithId(cmd *cobra.Command, deploymentGroupName, deviceId string, description, autoUpdateModeId *string) error {
	// Prepare the input for device-based creation
	isDeltaUpdateOnly := false
	createInput := &backend.DeploymentGroupCreateFromDeviceInput{
		DeviceId:          deviceId,
		Name:              deploymentGroupName,
		Description:       description,
		AutoUpdateModeId:  autoUpdateModeId,
		IsDeltaUpdateOnly: &isDeltaUpdateOnly,
	}

	// Create deployment group from device
	createResult, err := backend.CreateDeploymentGroupFromDevice(cmd.Context(), createInput)
	if err != nil {
		return fmt.Errorf("failed to create deployment group from device: %w", err)
	}

	msg := StoreDeploymentGroupCreateResult{
		Message:                "Deployment Group created from device",
		CreatedDeploymentGroup: createResult.CreateDeploymentGroupFromDevice,
	}

	return format.PrintFormattedOutput(cmd, msg, customStoreDeploymentGroupFormatter)
}

func customStoreDeploymentGroupFormatter(res StoreDeploymentGroupCreateResult) (string, error) {
	return res.Message, nil
}
