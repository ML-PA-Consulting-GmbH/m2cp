package manifest

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
	"m2cpcli/format"
	"m2cpcli/helper"
	"m2cpcli/structs"
	"m2cpcli/suit"
	"os"
	"os/exec"
)

var createCmd = &cobra.Command{
	Use:   "create -f <firmware.tar> -k <private-key> <manifest>",
	Short: "Create firmware manifest",
	Long:  "Create a new signed firmware manifest.",
	Args:  cobra.ExactArgs(1),
	Example: `To generate a signed manifest for a given firmware package, do:

  $ m2cp firmware manifest create -f /path/to/firmware.tar -k ed25519.pem signed-manifest.cbor

You may generate a valid private key with:

  $ openssl genpkey -algorithm Ed25519 -out ed25519-key.pem

Instead of passing a file, you may pass a string of colon-separated hexadecimal encoded bytes, as printed in:

  $ openssl pkey -in ed25519-key.pem -text -noout
  ...
  $ m2cp firmware manifest create -f /path/to/firmware.tar -k "00:ff:01:fe:..."
`,
	RunE: runCreateCmd,
}

func init() {
	ManifestCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("firmware-tar", "f", "", "path to the firmware TAR package")
	createCmd.Flags().StringP("private-key", "k", "", "path to the private key in PEM format. Valid types are defined by the suit-tool")
	createCmd.Flags().IntP("sequence-number", "s", 0, "sequence number to use. By default, UNIX seconds since epoch are used")
	createCmd.Flags().String("suit-tool", "", "path to the SUIT tool binary or executable Python script. By default, the executable is discovered")
	createCmd.Flags().String("vendor", "", "overwrite the tenant alias of the store owner. By default, the value will be retrieved")
	createCmd.Flags().Bool("keep", false, "keep temporary files")
}

type firmwareManifestCreateResult struct {
	DeviceModel                string           `json:"device-model" yaml:"device-model"`
	Vendor                     string           `json:"vendor" yaml:"vendor"`
	UriRoot                    string           `json:"uri-root" yaml:"uri-root"`
	SequenceNumber             int              `json:"sequence-number" yaml:"sequence-number"`
	SignedManifestFilepath     string           `json:"signed-manifest-filepath" yaml:"signed-manifest-filepath"`
	PublicKeyFingerprintSha256 string           `json:"public-key-fingerprint-sha256" yaml:"public-key-fingerprint-sha256"`
	KeepTemporaryFiles         bool             `json:"keep-temporary-files" yaml:"keep-temporary-files"`
	DebugInformation           DebugInformation `json:"debug-information" yaml:"-"`
}

type DebugInformation struct {
	CurrentWorkingDirectory   string                      `json:"current-working-directory"`
	ManifestCreationInputJson *suit.ManifestCreationInput `json:"manifest-creation-input-json"`
	Create                    *structs.CommandFeedback    `json:"create"`
	Sign                      *structs.CommandFeedback    `json:"sign"`
}

func runCreateCmd(cmd *cobra.Command, args []string) error {
	var err error
	result := &firmwareManifestCreateResult{}

	if cmd.Flags().Changed("keep") {
		result.KeepTemporaryFiles, err = cmd.Flags().GetBool("keep")
		if err != nil {
			return err
		}
	}

	if !cmd.Flags().Changed("firmware-tar") {
		return fmt.Errorf("missing required flag: firmware-tar")
	}
	firmwareTarFilepath, err := cmd.Flags().GetString("firmware-tar")
	if err != nil {
		return err
	}

	if !cmd.Flags().Changed("private-key") {
		return fmt.Errorf("missing required flag: private-key")
	}
	privateKeyFilepathOrEncodedData, err := cmd.Flags().GetString("private-key")
	if err != nil {
		return err
	}

	var privateKeyPemBytes []byte
	if helper.IsValidColonSeparatedHexadecimalEncodedBytes(privateKeyFilepathOrEncodedData) {
		privateKeyPemBytes, err = helper.ColonSeparatedHexadecimalEncodedBytesToPemFormatedPrivateKeyBytes(privateKeyFilepathOrEncodedData)
		if err != nil {
			return fmt.Errorf("could not parse private key: %v", err)
		}
	} else {
		privateKeyPemBytes, err = os.ReadFile(privateKeyFilepathOrEncodedData)
		if err != nil {
			return fmt.Errorf("could not read private key: %v", err)
		}
	}

	privateKey, err := helper.LoadEd25519PrivateKeyFromPEM(privateKeyPemBytes)
	if err != nil {
		return fmt.Errorf("could not load private key: %v", err)
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	result.PublicKeyFingerprintSha256 = helper.Sha256Fingerprint(publicKey)

	var suitToolExecutablePath string
	if cmd.Flags().Changed("suit-tool") {
		suitToolExecutablePath, err = cmd.Flags().GetString("suit-tool")
		if err != nil {
			return err
		}
	} else {
		const toolName = "m2cp-suit-tool"
		suitToolExecutablePath, err = exec.LookPath(toolName)
		if err != nil {
			return fmt.Errorf("could not find \"%s\" executable: %v", toolName, err)
		}
	}

	var metadata *suit.FirmwareMetadata
	metadata, err = suit.ExtractFirmwareMetadataFromTar(firmwareTarFilepath)
	if err != nil {
		return fmt.Errorf("could not extract firmware metadata: %v", err)
	}
	result.DeviceModel = metadata.DeviceModel

	// TODO: note, due to the way `suit-tool` works, we must unpack the firmware to the exact relative path where `m2cp-coap` will serve it later!
	result.DebugInformation.CurrentWorkingDirectory, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %v", err)
	}
	firmwareTempfile := make([]*os.File, 2)
	for idx := range firmwareTempfile {
		filepathTemplate := fmt.Sprintf("fw%d.bin", idx)
		firmwareBytes, err := suit.ExtractFileFromTar(firmwareTarFilepath, fmt.Sprintf("./%s", filepathTemplate))
		if err != nil {
			return fmt.Errorf("could not extract firmware for slot %d: %v", idx, err)
		}

		firmwareTempfile[idx], err = os.Create(fmt.Sprintf("%d", idx))
		if err != nil {
			return err
		}
		defer func() {
			_ = firmwareTempfile[idx].Close()
			if !result.KeepTemporaryFiles {
				_ = os.Remove(firmwareTempfile[idx].Name())
			}
		}()

		// Write data to the file
		if _, err := firmwareTempfile[idx].Write(firmwareBytes); err != nil {
			return err
		}
	}

	// vendor-id ist bei uns im Backend jetzt der tenant alias vom store owner
	if cmd.Flags().Changed("vendor") {
		// local definition overrides retrieval from the backend
		result.Vendor, err = cmd.Flags().GetString("vendor")
		if err != nil {
			return err
		}
	} else {
		// TODO: vendor retrieval not implemented.
		// TODO: there is one exception for the ML!PA internal store!
		return fmt.Errorf("vendor retrieval not implemented")
	}

	class := result.DeviceModel // TODO a.k.a. model name

	if !cmd.Flags().Changed("sequence-number") {
		return fmt.Errorf("missing required flag: sequence-number")
	}
	result.SequenceNumber, err = cmd.Flags().GetInt("sequence-number")
	if err != nil {
		return err
	}

	slotfile := []string{
		fmt.Sprintf("%s:%d", firmwareTempfile[0].Name(), 0),
		fmt.Sprintf("%s:%d", firmwareTempfile[1].Name(), 0),
	}

	result.UriRoot = fmt.Sprintf("coap://[fdea:dbee:f::1]/suit/fw")

	result.DebugInformation.ManifestCreationInputJson, err = suit.NewManifestCreationInputFromFiles(result.Vendor, class, result.SequenceNumber,
		slotfile, result.UriRoot)
	if err != nil {
		return err
	}

	var manifest []byte
	manifest, result.DebugInformation.Create, err = suit.CreateWithExternalBinary([]byte(result.DebugInformation.ManifestCreationInputJson.String()),
		suitToolExecutablePath, result.KeepTemporaryFiles)
	if err != nil {
		return err
	}

	var signedManifest []byte
	signedManifest, result.DebugInformation.Sign, err = suit.SignWithExternalBinary(manifest, privateKeyPemBytes,
		suitToolExecutablePath, result.KeepTemporaryFiles)
	if err != nil {
		return err
	}

	result.SignedManifestFilepath = args[0]
	err = os.WriteFile(result.SignedManifestFilepath, signedManifest, 0644) // Overwrites if file exists
	if err != nil {
		return fmt.Errorf("could not write signed manifest to \"%s\": %v", result.SignedManifestFilepath, err)
	}

	return format.PrintFormattedOutput(cmd, result, nil)
}

//func firmwareManifestCreateFormatter(result any) (string, error) {
//	userStatus := result.(CreateResult)
//
//	return outputStr, nil
//}
