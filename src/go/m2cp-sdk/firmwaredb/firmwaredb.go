package firmwaredb

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"m2cp"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/plgd-dev/go-coap/v3/message/codes"
)

const fileManifest = "manifest"
const fileFirmware0 = "firmware0"
const fileFirmware1 = "firmware1"

const revisionTimeOut = 5 * time.Minute

type FirmwareDb struct {
	path                string
	ctp                 m2cp.ContextPlus
	clientRevisionCache clientRevisionCache
}

type clientRevisionCache struct {
	revisionsByClient     map[string]*Revision
	revisionsByClientLock sync.Mutex
}

type MetaJSON struct {
	Fwt       uint      `json:"fwt"`
	Hwr       uint      `json:"hwr"`
	Fwr       uint      `json:"fwr"`
	Allowance Allowance `json:"allowance"`
	Hashes    Hashes    `json:"hashes"`
}

type Revision struct {
	FWT       int       `json:"fwt"`
	HWR       int       `json:"hwr"`
	FWR       int       `json:"fwr"`
	Allowance Allowance `json:"allowance"`
	Hashes    Hashes    `json:"hashes"`
	lock      sync.Mutex
	created   time.Time
	path      string
}

type Allowance struct {
	Default string            `json:"default"`
	Sensors map[string]string `json:"sensors,omitempty"`
}

type Hashes struct {
	Meta      string `json:"-"`
	Manifest  string `json:"manifest"`
	Firmware0 string `json:"firmware0"`
	Firmware1 string `json:"firmware1"`
}

func newClientRevisionCache() clientRevisionCache {
	cache := clientRevisionCache{
		revisionsByClient: make(map[string]*Revision),
	}
	// go routine to clean up old revisions
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			cache.revisionsByClientLock.Lock()
			for key, revision := range cache.revisionsByClient {
				if time.Since(revision.created) > revisionTimeOut {
					// log.Printf("INFO: removing revision for %s", key)
					delete(cache.revisionsByClient, key)
				}
			}
			cache.revisionsByClientLock.Unlock()
		}
	}()
	return cache
}

func (o *clientRevisionCache) Put(key string, revision *Revision) {
	o.revisionsByClientLock.Lock()
	defer o.revisionsByClientLock.Unlock()
	o.revisionsByClient[key] = revision
}

func (o *clientRevisionCache) Get(key string) (*Revision, bool) {
	o.revisionsByClientLock.Lock()
	defer o.revisionsByClientLock.Unlock()
	revision, ok := o.revisionsByClient[key]
	return revision, ok
}

func (o *clientRevisionCache) Delete(key string) {
	o.revisionsByClientLock.Lock()
	defer o.revisionsByClientLock.Unlock()
	delete(o.revisionsByClient, key)
}

func New(ctxParent m2cp.ContextPlus, dataPath string) *FirmwareDb {
	var err error
	ctx := ctxParent.SetModule("firmwaredb")

	if dataPath, err = filepath.Abs(dataPath); err != nil {
		log.Fatalf("Failed to get absolute path for %s: %s", dataPath, err.Error())
	}
	firmwarePath := path.Join(dataPath, "firmware")
	ctx.LogDebug("data path: %s", dataPath)
	ctx.LogDebug("firmware path: %s", firmwarePath)

	if err := os.MkdirAll(firmwarePath, 0755); err != nil {
		log.Fatalf("Failed to create firmware directory: %s", err.Error())
	}
	return &FirmwareDb{
		path:                firmwarePath,
		ctp:                 ctx,
		clientRevisionCache: newClientRevisionCache(),
	}
}

func (o *FirmwareDb) GetRevisionInfo(fwt, hwr, fwr int) (*Revision, error) {
	pathRevision := path.Join(o.path, fmt.Sprintf("%d", fwt), fmt.Sprintf("%d", hwr), fmt.Sprintf("%d", fwr))
	return o.loadRevision(pathRevision, fwt, hwr, fwr, nil)
}

func (o *FirmwareDb) FindNewerRevision(sensor m2cp.CoapPeer) (revision *Revision, err error) {
	if sensor == nil {
		return nil, fmt.Errorf("sensor is nil")
	}
	minFirmwareRevision := sensor.GetSequenceNumber() + 1

	pathCompatibleRevisions := path.Join(o.path, fmt.Sprintf("%d", sensor.GetFirmWareType()), fmt.Sprintf("%d", sensor.GetHardWareRevision()))
	o.ctp.LogDebug("Looking for compatible revisions in %s", pathCompatibleRevisions)

	// Read the directory to get all revisions
	_, err = os.Stat(pathCompatibleRevisions)
	if os.IsNotExist(err) {
		o.ctp.LogDebug("Path no existent -> no compatible revisions found in %s", pathCompatibleRevisions)
		return nil, fmt.Errorf("%d - No compatible and allowed revision found", int(codes.NotFound))
	}
	files, err := os.ReadDir(pathCompatibleRevisions)
	if err != nil {
		return nil, err
	}

	// Sort the revisions in descending order
	sort.Slice(files, func(i, j int) bool {
		revI, _ := strconv.Atoi(files[i].Name())
		revJ, _ := strconv.Atoi(files[j].Name())
		return revI > revJ
	})

	// Iterate through sorted revisions to find the highest compatible one
	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		o.ctp.LogDebug("Scanning %s", f.Name())
		if firmwareRevision, err := strconv.Atoi(f.Name()); err != nil || firmwareRevision < minFirmwareRevision {
			continue
		}
		if f.IsDir() {
			var firmwareRevision int
			if fNameInt, err := strconv.Atoi(f.Name()); err != nil {
				continue // dir name is not a number -> not a revision
			} else {
				firmwareRevision = fNameInt
			}

			revisionPath := path.Join(pathCompatibleRevisions, f.Name())

			revision, err = o.loadRevision(revisionPath, sensor.GetFirmWareType(), sensor.GetHardWareRevision(), firmwareRevision, sensor)
			if err != nil {
				return nil, err
			}
			if revision != nil {
				return revision, nil
			}
		}
	}

	return nil, fmt.Errorf("%d - No compatible and allowed revision found", int(codes.NotFound))
}

func (o *FirmwareDb) loadRevision(revisionPath string, fwt, hwr, fwr int, sensor m2cp.CoapPeer) (revision *Revision, err error) {
	// TODO: move this function to firmwareDir class

	metaPath := path.Join(revisionPath, "meta.json")

	var revisionRaw []byte
	revisionRaw, err = os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening meta.json at %s: %s", revisionPath, err.Error())
	}
	revision = &Revision{}
	if err = json.Unmarshal(revisionRaw, &revision); err != nil {
		return nil, fmt.Errorf("failed parsing meta.json at %s: %s", revisionPath, err.Error()) //
	}
	revision.path = revisionPath
	revision.Hashes.Meta, err = revision.sha256(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed calculating hash for meta.json at %s: %s", revisionPath, err.Error())
	}

	o.ctp.LogDebug("Checking %s", revision.path)

	if revision.IsMetadataValid(fwt, hwr, fwr) &&
		revision.IsCompatible(fwt, hwr) &&
		(sensor == nil || revision.IsAllowed(sensor.GetHardwareSerial())) &&
		revision.IsHashesValid() {
		o.ctp.LogDebug("Found highest compatible revision %s", revision.path)
		revision.created = time.Now()
		if sensor != nil {
			o.clientRevisionCache.Put(sensor.GetAddress(), revision)
		}
		return revision, nil // Return the manifest of the highest compatible revision
	}
	return nil, nil
}

func (o *FirmwareDb) GetRevisionForKnownDevice(sensor m2cp.CoapPeer) (*Revision, error) {
	if sensor == nil {
		return nil, fmt.Errorf("%d - sensor is nil", int(codes.InternalServerError))
	}
	if revision, ok := o.clientRevisionCache.Get(sensor.GetAddress()); ok {
		return revision, nil
	}
	return nil, fmt.Errorf("%d - no revision found for %s", int(codes.NotFound), sensor)
}

func (o *FirmwareDb) Put(update FirmwareSyncData) (err error) {
	firmwareDir := NewFirmwareDir(o.path, update.Fwt, update.Hwr, update.Fwr)
	err = firmwareDir.Store(update.Meta, update.Manifest, update.Firmware0, update.Firmware1)
	return err
}

func (o *Revision) IsCompatible(firmwareType, hardwareRevision int) bool {
	return o.FWT == firmwareType && o.HWR == hardwareRevision
}

func (o *Revision) IsAllowed(sensorId string) bool {
	if allowance, ok := o.Allowance.Sensors[sensorId]; ok {
		return allowance == "allow"
	}
	return o.Allowance.Default == "allow"
}

func (o *Revision) IsHashesValid() bool {
	return o.verifyHash(path.Join(o.path, fileManifest), o.Hashes.Manifest) &&
		o.verifyHash(path.Join(o.path, fileFirmware0), o.Hashes.Firmware0) &&
		o.verifyHash(path.Join(o.path, fileFirmware1), o.Hashes.Firmware1)
}

func (o *Revision) verifyHash(path, hash string) bool {
	if h, err := o.sha256(path); err != nil {
		// log.Printf("ERROR: failed checking hash for %s: %s", path, err.Error())
		return false
	} else if hash != h {
		// log.Printf("ERROR: manifest hash mismatch in %s: %s != %s", path, h, hash)
		return false
	}
	return true
}

func (o *Revision) GetManifest() ([]byte, error) {
	return os.ReadFile(path.Join(o.path, fileManifest))
}

func (o *Revision) GetFirmware(slot uint8) ([]byte, error) {
	if slot == 0 {
		return os.ReadFile(path.Join(o.path, fileFirmware0))
	} else if slot == 1 {
		return os.ReadFile(path.Join(o.path, fileFirmware1))
	} else {
		return nil, fmt.Errorf("%d - invalid slot: %d", slot, int(codes.BadRequest))
	}
}

func (o *Revision) sha256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := hasher.Sum(nil)
	// encode with hex
	return hex.EncodeToString(hash), nil
	//return base64.StdEncoding.EncodeToString(hash), nil
}

func (o *Revision) IsMetadataValid(firmwareType, hardwareRevision, firmwareRevision int) bool {
	valid := o.FWT == firmwareType && o.HWR == hardwareRevision && o.FWR == firmwareRevision
	return valid
}
