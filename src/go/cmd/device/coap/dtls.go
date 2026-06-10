package coap

import (
	"context"
	"encoding/json"
	"fmt"
	"m2cp/crypto"
	"m2cpcli/backend"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var dtlsCmd = &cobra.Command{
	Use:   "dtls <deviceId|deviceSerial|deviceName> <enable|disable>",
	Short: "Configure DTLS for CoAP communication on a device's Edge Device",
	Long: `Configure DTLS for CoAP communication on the Edge Device associated with the given device.

When enabling DTLS, a PSK config file may be provided via --config. If omitted,
the existing configuration stored on the device is used (a prior upload with
--config is required in that case).

The PSK config file format is:
  <id>: <hex-key>
  <id>: <hex-key>
  ...`,
	Args: cobra.ExactArgs(2),
	RunE: runDtlsCmd,
}

func init() {
	dtlsCmd.Flags().StringP("config", "c", "", "path to PSK config file (only valid when enabling DTLS)")
	CoapCmd.AddCommand(dtlsCmd)
}

func runDtlsCmd(cmd *cobra.Command, args []string) error {
	action := strings.ToLower(args[1])
	if action != "enable" && action != "disable" {
		return fmt.Errorf("invalid action '%s': must be 'enable' or 'disable'", args[1])
	}
	enable := action == "enable"

	configFile, _ := cmd.Flags().GetString("config")
	if !enable && configFile != "" {
		return fmt.Errorf("--config is only valid when enabling DTLS")
	}

	ctx := cmd.Context()

	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(ctx, args[0])
	if err != nil {
		return err
	}
	dev, err := backend.GetDeviceInfoById(ctx, deviceId)
	if err != nil {
		return err
	}
	if dev.IsRTD() {
		return fmt.Errorf("DTLS configuration is only applicable to Edge Devices, but '%s' is not an ED", args[0])
	}

	node := fmt.Sprintf("rpc.m2cp-coap.%s", dev.DeviceSerial)

	// Build RPC parameters.
	params := map[string]string{
		"enable": fmt.Sprintf("%v", enable),
	}

	if enable && configFile != "" {

		plaintext, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config file '%s': %w", configFile, err)
		}

		pubKeyPEM, err := fetchPublicEncryptionKey(ctx, node)
		if err != nil {
			return fmt.Errorf("failed to fetch public encryption key from %s: %w", node, err)
		}

		fmt.Printf("encrypting config with device public key\n")
		envelope, err := crypto.EncryptForSecureChannel(plaintext, pubKeyPEM)
		if err != nil {
			return fmt.Errorf("failed to encrypt config: %w", err)
		}

		envelopeJSON, err := json.Marshal(envelope)
		if err != nil {
			return fmt.Errorf("failed to marshal encrypted config: %w", err)
		}

		params["dtlsConfig"] = string(envelopeJSON)
	} else {
		params["dtlsConfig"] = ""
	}

	fmt.Printf("calling ConfigureDTLS on %s (enable=%v)\n", node, enable)
	res, err := backend.DeviceRpc(ctx, node, "ConfigureDTLS", params)
	if err != nil {
		if strings.Contains(err.Error(), "timer expired") {
			fmt.Printf("ConfigureDTLS timed out, retrying once...\n")
			res, err = backend.DeviceRpc(ctx, node, "ConfigureDTLS", params)
		}
		if err != nil {
			return fmt.Errorf("RPC failed: %w", err)
		}
	}
	if res.Error != 0 {
		return fmt.Errorf("ConfigureDTLS failed: %s", res.Message)
	}

	fmt.Printf("success: %s\n", res.Message)
	return nil
}

// fetchPublicEncryptionKey calls GetPublicEncryptionKey on the given RPC node
// and returns the parsed RSA public key.
func fetchPublicEncryptionKey(ctx context.Context, node string) (string, error) {
	fmt.Printf("fetching public encryption key from %s\n", node)
	res, err := backend.DeviceRpc(ctx, node, "GetPublicEncryptionKey", nil)
	if err != nil {
		return "", err
	}
	if res.Error != 0 {
		return "", fmt.Errorf("GetPublicEncryptionKey RPC failed: %s", res.Message)
	}
	pemStr, ok := res.Result["publicKey"]
	if !ok {
		return "", fmt.Errorf("response missing 'publicKey' field")
	}
	return pemStr, nil
}
