package suit

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type ManifestCreationInput struct {
	ManifestVersion        int         `json:"manifest-version"`
	ManifestSequenceNumber int         `json:"manifest-sequence-number"`
	Components             []Component `json:"components"`
}

type Component struct {
	InstallId []string `json:"install-id"`
	//InstallDigest InstallDigest `json:"install-digest,omitempty"`
	InstallSize int    `json:"install-size,omitempty"`
	VendorId    string `json:"vendor-id"`
	ClassId     string `json:"class-id"`
	File        string `json:"file,omitempty"`
	Uri         string `json:"uri"`
	Bootable    bool   `json:"bootable"`
	Offset      int64  `json:"offset,omitempty"`
}

type InstallDigest struct {
	AlgorithmID string `json:"algorithm-id,omitempty"`
	DigestBytes string `json:"digest-bytes,omitempty"`
}

func NewInstallDigestFromBytes(data []byte) (*InstallDigest, error) {
	hash := sha256.Sum256(data)
	hashString := hex.EncodeToString(hash[:])

	result := InstallDigest{
		AlgorithmID: "sha256",
		DigestBytes: hashString,
	}
	return &result, nil
}

// UUIDv5 generates a UUIDv5 based on namespace and name using SHA-1.
// It is implemented manually using SHA-1 since the Go uuid package supports it only in some versions.
func UUIDv5(namespace uuid.UUID, name string) uuid.UUID {
	hasher := sha1.New()
	hasher.Write(namespace[:])
	hasher.Write([]byte(name))
	hash := hasher.Sum(nil)
	var uuidBytes [16]byte
	copy(uuidBytes[:], hash)
	uuidBytes[6] = (uuidBytes[6] & 0x0f) | 0x50
	uuidBytes[8] = (uuidBytes[8] & 0x3f) | 0x80
	return uuid.UUID(uuidBytes)
}

// ParseInt converts a string to an integer.
// Supports both decimal (e.g., "123") and hexadecimal (e.g., "0x7B") notations.
func ParseInt(input string) (int64, error) {
	// Use base 0 so strconv.ParseInt can infer the base from the prefix
	input = strings.TrimSpace(input)
	value, err := strconv.ParseInt(input, 0, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

type Image struct {
	Filename  string
	Offset    int64
	CompNames []string
}

func NewImage(filenameOffset string) (*Image, error) {
	split := strings.SplitN(filenameOffset, ":", 3)
	var filename string
	var offset int64
	var err error
	compName := []string{"00"}

	switch len(split) {
	case 1:
		filename = split[0]
		offset = 0
	case 2:
		filename = split[0]
		offset, err = ParseInt(split[1])
		if err != nil {
			return nil, err
		}
	case 3:
		filename = split[0]
		offset, err = ParseInt(split[1])
		if err != nil {
			return nil, err
		}
		compName = strings.Split(split[2], ":")
	default:
		return nil, fmt.Errorf("invalid image filename: %s", filenameOffset)
	}

	image := Image{filename, offset, compName}

	return &image, nil
}

func (i *InstallDigest) String() string {
	return fmt.Sprintf("%s:%s", i.AlgorithmID, i.DigestBytes)
}

// NewManifestCreationInputFromFiles aims to resemble https://github.com/RIOT-OS/RIOT/blob/master/dist/tools/suit/gen_manifest.py when
// rendered as a string.
func NewManifestCreationInputFromFiles(vendor string, class string, sequenceNumber int, slotfiles []string, urlRoot string) (*ManifestCreationInput, error) {
	result := ManifestCreationInput{
		ManifestVersion:        1,
		ManifestSequenceNumber: sequenceNumber,
	}

	uuidVendor := UUIDv5(uuid.NameSpaceDNS, vendor)
	uuidClass := UUIDv5(uuidVendor, class)

	var images []Image
	for _, filenameOffset := range slotfiles {
		image, err := NewImage(filenameOffset)
		if err != nil {
			return nil, err
		}
		images = append(images, *image)
	}

	for _, image := range images {
		uri, err := url.Parse(urlRoot)
		if err != nil {
			return nil, err
		}
		uri.Path = path.Join(uri.Path, image.Filename)

		component := Component{
			InstallId: image.CompNames,
			VendorId:  strings.ReplaceAll(uuidVendor.String(), "-", ""),
			ClassId:   strings.ReplaceAll(uuidClass.String(), "-", ""),
			File:      image.Filename,
			Uri:       uri.String(),
			Bootable:  false,
		}

		if image.Offset != 0 {
			component.Offset = image.Offset
		}
		result.Components = append(result.Components, component)
	}

	return &result, nil
}

func (i *ManifestCreationInput) AddComponent(component *Component) {
	i.Components = append(i.Components, *component)
}

func (i *ManifestCreationInput) String() string {
	jsonBytes, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return fmt.Sprintf("Failed to marshal JSON: %v", err)
	}
	return string(jsonBytes)
}
