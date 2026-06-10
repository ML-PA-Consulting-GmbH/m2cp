package firmware

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/suit"
)

import (
	"m2cpcli/format"
)

var infoCmd = &cobra.Command{
	Use:   "info <firmware.tar>",
	Short: "Info on firmware package",
	Long:  "Print information on a given firmware package",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfoCmd,
}

func init() {
	FirmwareCmd.AddCommand(infoCmd)
}

type firmwareInfoResult struct {
	Metadata *suit.FirmwareMetadata `json:"firmware-metadata" yaml:"firmware-metadata"`
	Firmware []FirmwareMetadata     `json:"firmware" yaml:"firmware"`
}

type FirmwareMetadata struct {
	Filename string `json:"filename" yaml:"filename"`
	Bytes    []byte `json:"-" yaml:"-"`
	Size     int64  `json:"size" yaml:"size"`
}

func runInfoCmd(cmd *cobra.Command, args []string) error {
	var err error
	result := &firmwareInfoResult{}

	firmwareTarFilepath := args[0]

	result.Metadata, err = suit.ExtractFirmwareMetadataFromTar(firmwareTarFilepath)
	if err != nil {
		return fmt.Errorf("could not extract firmware metadata: %v", err)
	}

	result.Firmware = make([]FirmwareMetadata, 2)
	for idx := range result.Firmware {
		result.Firmware[idx] = FirmwareMetadata{
			Filename: fmt.Sprintf("fw%d.bin", idx),
		}

		result.Firmware[idx].Bytes, err = suit.ExtractFileFromTar(firmwareTarFilepath, fmt.Sprintf("./%s", result.Firmware[idx].Filename))
		if err != nil {
			return fmt.Errorf("could not extract firmware for slot %d: %v", idx, err)
		}

		result.Firmware[idx].Size = int64(len(result.Firmware[idx].Bytes))
		//result.Firmware[idx].Header = string(result.Firmware[idx].Bytes[:8]) // TODO: is there any file magic, we could do?
	}

	return format.PrintFormattedOutput(cmd, result, nil)
}
