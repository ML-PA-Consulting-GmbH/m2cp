package suit

import (
	"archive/tar"
	"bytes"
	"fmt"
	"gopkg.in/square/go-jose.v2/json"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type FirmwareMetadata struct {
	AppName      string `json:"app-name" yaml:"app-name"`
	AppSummary   string `json:"app-summary" yaml:"app-summary"`
	AppVersion   string `json:"app-version" yaml:"app-version"`
	DeviceModel  string `json:"device-model" yaml:"device-model"`
	Architecture string `json:"architecture" yaml:"architecture"`

	// optional fields for expert developers
	ClassId  string          `json:"class-id" yaml:"class-id"`
	Bootable bool            `json:"bootable" yaml:"bootable"`
	Offset   map[string]uint `json:"offset" yaml:"offset"`
}

// ExtractFileFromTar will extract `targetFilename` only from the given `tarFilepath`.
func ExtractFileFromTar(tarFilepath, targetFilename string) ([]byte, error) {
	file, err := os.Open(tarFilepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tarReader := tar.NewReader(file)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag == tar.TypeReg && header.Name == targetFilename {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tarReader); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}
	}

	return nil, fmt.Errorf("file %q not found in archive", targetFilename)
}

func ValidateFirmwareTar(filepath string) error {
	_, err := ExtractFirmwareMetadataFromTar(filepath)
	if err != nil {
		return err
	}

	err = ValidateFirmwareBlobFromTar(filepath, "./fw0.bin")
	if err != nil {
		return err
	}

	err = ValidateFirmwareBlobFromTar(filepath, "./fw1.bin")
	if err != nil {
		return err
	}

	return nil
}

func ValidateFirmwareMetadata(metadata FirmwareMetadata) error {
	if metadata.AppName == "" {
		return fmt.Errorf("app name must not be empty")
	}
	if metadata.AppSummary == "" {
		return fmt.Errorf("app summary must not be empty")

	}
	if metadata.AppVersion == "" {
		return fmt.Errorf("app version must not be empty")

	}
	if metadata.DeviceModel == "" {
		return fmt.Errorf("device model must not be empty")

	}
	if metadata.Architecture != "arm32" {
		return fmt.Errorf("unexpected architecture: \"%s\"", metadata.Architecture)
	}
	return nil
}

// ValidateFirmwareBlobFromTar checks if the firmware size is larger than zero.
func ValidateFirmwareBlobFromTar(tarFilepath string, blobFilepath string) error {
	binaryBlob, err := ExtractFileFromTar(tarFilepath, blobFilepath)
	if err != nil {
		return fmt.Errorf("could not extract blob \"%s\": %w", blobFilepath, err)
	}
	if len(binaryBlob) == 0 {
		return fmt.Errorf("empty blob \"%s\"", blobFilepath)
	}
	return nil
}

// ExtractFirmwareMetadataFromTar extracts and validates the metadata
func ExtractFirmwareMetadataFromTar(filepath string) (*FirmwareMetadata, error) {
	firmwareYamlBytes, err := ExtractFileFromTar(filepath, "./meta/firmware.yaml")
	if err != nil {
		return nil, fmt.Errorf("at extracting firmware metadata: %v", err)
	}
	var firwareMetadata FirmwareMetadata
	err = yaml.Unmarshal(firmwareYamlBytes, &firwareMetadata)
	if err != nil {
		return nil, fmt.Errorf("at parsing firmware metadata: %v", err)
	}

	err = ValidateFirmwareMetadata(firwareMetadata)
	if err != nil {
		return nil, fmt.Errorf("at validating firmware metadata: %v", err)
	}

	return &firwareMetadata, nil
}

func (f *FirmwareMetadata) String() string {
	jsonData, err := json.Marshal(f)
	if err != nil {
		fmt.Println(err)
		return "error"
	}
	return string(jsonData)
}
