package suit

import (
	"bytes"
	"m2cpcli/structs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//// TODO: a map is not ordered! The original uses an OrderedDict!
//type SuitManifest map[int][]byte
//
//func NewIntermediateFormat(input ManifestCreationInput) (*IntermediateFormat, error) {
//	componentID := input.Components[0].InstallId // assuming all components share the same ID (e.g., ["00"])
//
//	// Build common.components
//	commonComponents := [][]string{componentID}
//
//	// Build common-sequence
//	commonSequence := []Command{}
//
//	// First directive-override-parameters (common)
//	commonSequence = append(commonSequence, Command{
//		CommandID: "directive-override-parameters",
//		CommandArg: map[string]interface{}{
//			"vendor-id":  input.Components[0].VendorId,
//			"class-id":   input.Components[0].ClassId,
//			"image-size": input.Components[0].InstallSize,
//		},
//		ComponentID: componentID,
//	})
//
//	// directive-try-each for override parameters, offset & digest
//	tryEachList := []interface{}{}
//	for _, comp := range input.Components {
//		seq := []Command{
//			{
//				CommandID: "directive-override-parameters",
//				CommandArg: map[string]interface{}{
//					"offset": comp.Offset,
//				},
//				ComponentID: comp.InstallId,
//			},
//			{
//				CommandID:   "condition-component-offset",
//				CommandArg:  5,
//				ComponentID: comp.InstallId,
//			},
//			{
//				CommandID: "directive-override-parameters",
//				//CommandArg: map[string]interface{}{
//				//	"image-digest": Digest{
//				//		AlgorithmID: comp.InstallDigest.AlgorithmID,
//				//		DigestBytes: comp.InstallDigest.DigestBytes,
//				//	},
//				//},
//				ComponentID: comp.InstallId,
//			},
//		}
//		tryEachList = append(tryEachList, seq)
//	}
//
//	commonSequence = append(commonSequence, Command{
//		CommandID:   "directive-try-each",
//		CommandArg:  tryEachList,
//		ComponentID: componentID,
//	})
//
//	// Vendor and class condition checks
//	commonSequence = append(commonSequence, []Command{
//		{
//			CommandID:   "condition-vendor-identifier",
//			CommandArg:  15,
//			ComponentID: componentID,
//		},
//		{
//			CommandID:   "condition-class-identifier",
//			CommandArg:  15,
//			ComponentID: componentID,
//		},
//	}...)
//
//	// Build install
//	installTryEach := []interface{}{}
//	for _, comp := range input.Components {
//		seq := []Command{
//			{
//				CommandID: "directive-set-parameters",
//				CommandArg: map[string]interface{}{
//					"offset": comp.Offset,
//				},
//				ComponentID: comp.InstallId,
//			},
//			{
//				CommandID:   "condition-component-offset",
//				CommandArg:  5,
//				ComponentID: comp.InstallId,
//			},
//			{
//				CommandID: "directive-set-parameters",
//				CommandArg: map[string]interface{}{
//					"uri": comp.Uri,
//				},
//				ComponentID: comp.InstallId,
//			},
//		}
//		installTryEach = append(installTryEach, seq)
//	}
//
//	install := []Command{
//		{
//			CommandID:   "directive-try-each",
//			CommandArg:  installTryEach,
//			ComponentID: componentID,
//		},
//		{
//			CommandID:   "directive-fetch",
//			CommandArg:  2,
//			ComponentID: componentID,
//		},
//		{
//			CommandID:   "condition-image-match",
//			CommandArg:  15,
//			ComponentID: componentID,
//		},
//	}
//
//	// Build validate
//	validate := []Command{
//		{
//			CommandID:   "condition-image-match",
//			CommandArg:  15,
//			ComponentID: componentID,
//		},
//	}
//
//	result := IntermediateFormat{
//		AuthenticationWrapper: []interface{}{},
//		Manifest: Manifest{
//			ManifestVersion:        input.ManifestVersion,
//			ManifestSequenceNumber: input.ManifestSequenceNumber,
//			Common: Common{
//				Components:     commonComponents,
//				CommonSequence: commonSequence,
//			},
//			Install:  install,
//			Validate: validate,
//		},
//	}
//
//	return &result, nil
//}

//func (c *IntermediateFormat) ToSuitManifest() (*SuitManifest, error) {
//	// TODO: hardcoded is bad!
//	suitManifest := SuitManifest{
//		2: []byte{0x80},
//		3: []byte{
//			0xa5, 0x01, 0x01, 0x02, 0x1a, 0x00, 0x1e, 0x9b, 0xf4, 0x03, 0x58, 0xa6, 0xa2, 0x02, 0x81,
//			0x81, 0x41, 0x00, 0x04, 0x58, 0x9d, 0x88, 0x14, 0xa3, 0x01, 0x50, 0x95, 0xb2, 0xa5, 0xc5,
//			0x12, 0x5f, 0x5a, 0x84, 0xad, 0xa0, 0x11, 0x61, 0x39, 0xc7, 0x02, 0x07, 0x02, 0x50, 0x0b,
//			0x83, 0x10, 0x11, 0x35, 0x3f, 0x5b, 0x12, 0x84, 0x2e, 0xed, 0x2e, 0x22, 0x49, 0x61, 0x84,
//			0x0e, 0x1a, 0x00, 0x03, 0x71, 0x38, 0x0f, 0x82, 0x58, 0x32, 0x86, 0x14, 0xa1, 0x05, 0x19,
//			0x40, 0x00, 0x05, 0x05, 0x14, 0xa1, 0x03, 0x58, 0x24, 0x82, 0x02, 0x58, 0x20, 0x98, 0x23,
//			0xd8, 0x04, 0x3c, 0xe0, 0x24, 0x3b, 0x46, 0x3c, 0xd4, 0xe6, 0xde, 0x49, 0x59, 0x33, 0x0d,
//			0xaf, 0x66, 0x39, 0xe4, 0x47, 0xdc, 0xf3, 0xbb, 0xd7, 0x1f, 0x37, 0xbb, 0x32, 0x84, 0x54,
//			0x58, 0x34, 0x86, 0x14, 0xa1, 0x05, 0x1a, 0x00, 0x08, 0x20, 0x00, 0x05, 0x05, 0x14, 0xa1,
//			0x03, 0x58, 0x24, 0x82, 0x02, 0x58, 0x20, 0x62, 0x86, 0x9b, 0x55, 0xfa, 0x1d, 0x26, 0xa3,
//			0xd7, 0x57, 0xab, 0x93, 0x0d, 0x52, 0x94, 0xb9, 0x84, 0xaa, 0x6f, 0xd1, 0x07, 0xf4, 0xb9,
//			0x8d, 0x7a, 0x48, 0xd2, 0x98, 0x4d, 0x33, 0x95, 0xf8, 0x01, 0x0f, 0x02, 0x0f, 0x09, 0x58,
//			0x69, 0x86, 0x0f, 0x82, 0x58, 0x2e, 0x86, 0x13, 0xa1, 0x05, 0x19, 0x40, 0x00, 0x05, 0x05,
//			0x13, 0xa1, 0x15, 0x78, 0x20, 0x63, 0x6f, 0x61, 0x70, 0x3a, 0x2f, 0x2f, 0x5b, 0x66, 0x64,
//			0x65, 0x61, 0x3a, 0x64, 0x62, 0x65, 0x65, 0x3a, 0x66, 0x3a, 0x3a, 0x31, 0x5d, 0x2f, 0x66,
//			0x77, 0x2f, 0x31, 0x2f, 0x64, 0x2f, 0x30, 0x58, 0x30, 0x86, 0x13, 0xa1, 0x05, 0x1a, 0x00,
//			0x08, 0x20, 0x00, 0x05, 0x05, 0x13, 0xa1, 0x15, 0x78, 0x20, 0x63, 0x6f, 0x61, 0x70, 0x3a,
//			0x2f, 0x2f, 0x5b, 0x66, 0x64, 0x65, 0x61, 0x3a, 0x64, 0x62, 0x65, 0x65, 0x3a, 0x66, 0x3a,
//			0x3a, 0x31, 0x5d, 0x2f, 0x66, 0x77, 0x2f, 0x31, 0x2f, 0x64, 0x2f, 0x31, 0x15, 0x02, 0x03,
//			0x0f, 0x0a, 0x43, 0x82, 0x03, 0x0f,
//		},
//	}
//
//	return &suitManifest, nil
//}

//// Create resembles `suit-tool create -f suit -i IFILE -o OFILE`, where the return value is a CBOR-encoded SUIT manifest.
//func Create(manifestCreationInputJson []byte) ([]byte, error) {
//	var manifestCreationInput ManifestCreationInput
//
//	err := json.Unmarshal(manifestCreationInputJson, &manifestCreationInput)
//	if err != nil {
//		return nil, fmt.Errorf("error unmarshalling manifest creation input: %w", err)
//	}
//
//	intermediateFormat, err := NewIntermediateFormat(manifestCreationInput)
//	if err != nil {
//		return nil, fmt.Errorf("error creating compiled manifest: %w", err)
//	}
//
//	suitManifest, err := intermediateFormat.ToSuitManifest()
//	if err != nil {
//		return nil, fmt.Errorf("error creating suit manifest: %w", err)
//	}
//
//	cborData, err := cbor.Marshal(suitManifest)
//	if err != nil {
//		return nil, fmt.Errorf("error marshalling CBOR: %w", err)
//	}
//	//fmt.Printf("CBOR Data: %x\n", cborData)
//
//	return cborData, nil
//}

func m2cpCliTempDir() string {
	dirPath := filepath.Join(os.TempDir(), "m2cp-cli")

	err := os.MkdirAll(dirPath, 0700)
	if err != nil {
		// it will crash again in os.CreateTemp() soon
	}

	return dirPath
}

func CreateWithExternalBinary(manifestCreationInputJson []byte, suitToolExecutableFilepath string, keepTemporaryFiles bool) ([]byte, *structs.CommandFeedback, error) {
	jsonFile, err := os.CreateTemp(m2cpCliTempDir(), "jsonInputFile-*.json")
	if err != nil {
		return nil, nil, err
	}
	if !keepTemporaryFiles {
		defer os.Remove(jsonFile.Name())
	}

	// Write data to the file
	if _, err := jsonFile.Write(manifestCreationInputJson); err != nil {
		return nil, nil, err
	}

	cborFile, err := os.CreateTemp(m2cpCliTempDir(), "cborOutputFile-*.cborBytes")
	if err != nil {
		return nil, nil, err
	}
	err = cborFile.Close()
	if err != nil {
		return nil, nil, err
	}
	if !keepTemporaryFiles {
		defer os.Remove(cborFile.Name())
	}

	cmdAndArgs := []string{
		suitToolExecutableFilepath,
		//"python3",
		//"/home/jakob.duebel@ml-pa.loc/Downloads/git/RIOT/dist/tools/suit/suit-manifest-generator/bin/suit-tool",
		//"/home/jakob.duebel@ml-pa.loc/Downloads/git/RIOT/dist/tools/suit/suit-manifest-generator/dist/suit-tool",
		"--log-level", "debug",
		"create",
		"--format", "suit",
		"--input-file", jsonFile.Name(),
		"--output-file", cborFile.Name(),
	}
	cmd := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Run the command
	err = cmd.Run()
	if err != nil {
		return nil, nil, err
	}

	cfb := structs.CommandFeedback{
		Call:   strings.Join(cmdAndArgs, " "),
		Stdout: stdoutBuf.String(),
		Stderr: stderrBuf.String(),
	}

	cborBytes, err := os.ReadFile(cborFile.Name())
	if err != nil {
		return nil, nil, err
	}

	if err := jsonFile.Close(); err != nil {
		return nil, nil, err
	}
	return cborBytes, &cfb, nil
}

// SignWithExternalBinary resembles `suit-tool sign -m MANIFEST -k PRIVKEY`, where the return value is a signed CBOR-encoded SUIT manifest.
func SignWithExternalBinary(manifest []byte, privateKeyPem []byte, suitToolExecutableFilepath string, keepTemporaryFiles bool) ([]byte, *structs.CommandFeedback, error) {

	manifestFile, err := os.CreateTemp(m2cpCliTempDir(), "manifestFile-*.bin")
	if err != nil {
		return nil, nil, err
	}
	if !keepTemporaryFiles {
		defer os.Remove(manifestFile.Name())
	}
	if _, err := manifestFile.Write(manifest); err != nil {
		return nil, nil, err
	}

	// Note, the file mode is `0o600`, so the key is protected as good as in the $HOME directory
	keyFile, err := os.CreateTemp(m2cpCliTempDir(), "keyFile-*.pem")
	if err != nil {
		return nil, nil, err
	}
	if !keepTemporaryFiles {
		defer os.Remove(keyFile.Name())
	}

	if _, err := keyFile.Write(privateKeyPem); err != nil {
		return nil, nil, err
	}

	signedManifestFile, err := os.CreateTemp(m2cpCliTempDir(), "signedManifestFile-*.bin")
	if err != nil {
		return nil, nil, err
	}
	if !keepTemporaryFiles {
		defer os.Remove(signedManifestFile.Name())
	}

	cmdAndArgs := []string{
		suitToolExecutableFilepath,
		//"python3",
		//"/home/jakob.duebel@ml-pa.loc/Downloads/git/RIOT/dist/tools/suit/suit-manifest-generator/bin/suit-tool",
		//"/home/jakob.duebel@ml-pa.loc/Downloads/git/RIOT/dist/tools/suit/suit-manifest-generator/dist/suit-tool",
		"--log-level", "debug",
		"sign",
		"--manifest", manifestFile.Name(),
		"--private-key", keyFile.Name(),
		"--output-file", signedManifestFile.Name(),
	}
	cmd := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Run the command
	err = cmd.Run()
	if err != nil {
		return nil, nil, err
	}

	cfb := structs.CommandFeedback{
		Call:   strings.Join(cmdAndArgs, " "),
		Stdout: stdoutBuf.String(),
		Stderr: stderrBuf.String(),
	}

	signedManifestBytes, err := os.ReadFile(signedManifestFile.Name())
	if err != nil {
		return nil, nil, err
	}

	if err := signedManifestFile.Close(); err != nil {
		return nil, nil, err
	}
	return signedManifestBytes, &cfb, nil
}
