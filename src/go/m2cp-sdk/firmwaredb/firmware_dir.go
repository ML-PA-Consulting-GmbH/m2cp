package firmwaredb

import (
	"fmt"
	"m2cp/tools"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type FirmwareDir struct {
	path          string
	fwt, hwr, fwr uint
	metaHash      string
	manifestHash  string
	firmwareHash  []string
}

func NewFirmwareDir(basePath string, fwt, hwr, fwr uint) FirmwareDir {
	firmwareDir := FirmwareDir{
		path: path.Join(basePath, fmt.Sprint(fwt), fmt.Sprint(hwr), fmt.Sprint(fwr)),
		fwt:  fwt,
		hwr:  hwr,
		fwr:  fwr,
	}
	return firmwareDir
}

func NewFirmwareDirFromPath(fullPath string) (*FirmwareDir, error) {
	if !strings.HasSuffix(fullPath, "/meta.json") {
		if !strings.HasSuffix(string(os.PathSeparator), fullPath) {
			fullPath += string(os.PathSeparator)
		}
	}
	var err error
	if fullPath, err = filepath.Abs(path.Dir(fullPath)); err != nil {
		return nil, err
	}

	pathComponents := strings.Split(fullPath, string(os.PathSeparator))
	if len(pathComponents) < 4 {
		return nil, fmt.Errorf("invalid firmware path: %s", fullPath)
	}
	var fwt, hwr, fwr int64
	if fwr, err = strconv.ParseInt(pathComponents[len(pathComponents)-1], 10, 32); err != nil {
		return nil, fmt.Errorf("invalid firmware path: %s", fullPath)
	}
	if hwr, err = strconv.ParseInt(pathComponents[len(pathComponents)-2], 10, 32); err != nil {
		return nil, fmt.Errorf("invalid firmware path: %s", fullPath)
	}
	if fwt, err = strconv.ParseInt(pathComponents[len(pathComponents)-3], 10, 32); err != nil {
		return nil, fmt.Errorf("invalid firmware path: %s", fullPath)
	}
	return &FirmwareDir{
		path: fullPath,
		fwt:  uint(fwt),
		hwr:  uint(hwr),
		fwr:  uint(fwr),
	}, nil
}

func (o *FirmwareDir) Path() string {
	return o.path
}

func (o *FirmwareDir) Exists() bool {
	info, err := os.Stat(o.path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}

func (o *FirmwareDir) Verify() error {
	// TODO: verify file hashes against hashes in meta.json
	return nil
}

func (o *FirmwareDir) Fwt() uint {
	return o.fwt
}

func (o *FirmwareDir) Hwr() uint {
	return o.hwr
}

func (o *FirmwareDir) Fwr() uint {
	return o.fwr
}

func (o *FirmwareDir) MetaPath() string {
	return path.Join(o.path, "meta.json")
}

func (o *FirmwareDir) MetaHash() (hash string, err error) {
	if o.metaHash == "" {
		o.metaHash, err = tools.Sha256Sum(o.MetaPath())
	}
	return o.metaHash, err
}

func (o *FirmwareDir) ManifestPath() string {
	return path.Join(o.path, "manifest")
}

func (o *FirmwareDir) ManifestHash() (hash string, err error) {
	if o.manifestHash == "" {

		o.manifestHash, err = tools.Sha256Sum(o.ManifestPath())
	}
	return o.manifestHash, err
}

func (o *FirmwareDir) FirmwarePath(index int) string {
	return path.Join(o.path, fmt.Sprintf("firmware%d", index))
}

func (o *FirmwareDir) FirmwareHash(index int) (hash string, err error) {
	if o.firmwareHash == nil {
		o.firmwareHash = []string{"", ""}
	}
	if o.firmwareHash[index] == "" {
		o.firmwareHash[index], err = tools.Sha256Sum(o.FirmwarePath(index))
	}
	return o.firmwareHash[index], err
}

func (o *FirmwareDir) GetSyncItem() (item *FirmwareSyncItem, err error) {
	item = &FirmwareSyncItem{
		Fwt: o.Fwt(),
		Hwr: o.Hwr(),
		Fwr: o.Fwr(),
	}

	if item.MetaHash, err = o.MetaHash(); err != nil {
		return nil, fmt.Errorf("error calculating hash: %s", err)
	}
	if item.ManifestHash, err = o.ManifestHash(); err != nil {
		return nil, fmt.Errorf("error calculating hash: %s", err)
	}
	for i := 0; i < 2; i++ {
		if item.Firmware0Hash, err = o.FirmwareHash(0); err != nil {
			return nil, fmt.Errorf("error calculating hash: %s", err)
		}
	}
	return item, nil
}

func (o *FirmwareDir) GetSyncData() (*FirmwareSyncData, error) {
	data := &FirmwareSyncData{
		Fwt: o.fwt,
		Hwr: o.hwr,
		Fwr: o.fwr,
	}

	var err error

	data.Meta, err = os.ReadFile(o.MetaPath())
	if err != nil {
		return nil, fmt.Errorf("error reading meta file: %s", err)
	}

	data.Manifest, err = os.ReadFile(o.ManifestPath())
	if err != nil {
		return nil, fmt.Errorf("error reading manifest file: %s", err)
	}

	data.Firmware0, err = os.ReadFile(o.FirmwarePath(0))
	if err != nil {
		return nil, fmt.Errorf("error reading firmware0 file: %s", err)
	}

	data.Firmware1, err = os.ReadFile(o.FirmwarePath(1))
	if err != nil {
		return nil, fmt.Errorf("error reading firmware1 file: %s", err)
	}

	return data, nil
}

func (o *FirmwareDir) Store(meta []byte, manifest []byte, firmware0 []byte, firmware1 []byte) (err error) {
	// Ensure the directory exists
	err = os.MkdirAll(o.path, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %s", err)
	}

	// Function to write data to a file, overwrite if exists
	writeFile := func(filePath string, data []byte) error {
		return os.WriteFile(filePath, data, 0644)
	}

	// Store each file
	if err = writeFile(o.MetaPath(), meta); err != nil {
		return fmt.Errorf("error writing meta.json: %s", err)
	}
	if err = writeFile(o.ManifestPath(), manifest); err != nil {
		return fmt.Errorf("error writing manifest: %s", err)
	}
	if err = writeFile(o.FirmwarePath(0), firmware0); err != nil {
		return fmt.Errorf("error writing firmware0: %s", err)
	}
	if err = writeFile(o.FirmwarePath(1), firmware1); err != nil {
		return fmt.Errorf("error writing firmware1: %s", err)
	}

	return nil
}
