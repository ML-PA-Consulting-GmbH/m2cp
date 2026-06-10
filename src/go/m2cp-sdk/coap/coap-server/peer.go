package coap_server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
	"m2cp"
	"strconv"
	"time"
)

const (
	namespace = "32e58df2-3b30-11f0-9033-325096b39f47"
)

type Peer struct {
	OSSerial  string `json:"os_serial,omitempty"`
	Address   string `json:"address,omitempty"`
	EDAddress string `json:"ed_address,omitempty"`

	HardwareSerial string `json:"hw_serial,omitempty"`
	HardwareModel  string `json:"hw_model,omitempty"`
	HardwareVendor string `json:"hw_vendor,omitempty"`
	MCUID          string `json:"mcu_id,omitempty"`

	DeviceModel    string `json:"dev_model,omitempty"`
	AppName        string `json:"app_name,omitempty"`
	AppVersion     string `json:"app_version,omitempty"`
	SequenceNumber int    `json:"seq_no"`

	LastSeen     time.Time `json:"last_seen,omitempty"`
	Uptime       int64     `json:"uptime,omitempty"`
	FirmwareKeys []string  `json:"keys,omitempty"`

	UpdateResult     int       `json:"update_result,omitempty"`
	UpdateResultTime time.Time `json:"update_result_time,omitempty"`
	UpdateResultCode string    `json:"update_result_code,omitempty"`

	//Legacy fields for compatibility
	FirmWareType     int  `json:"firmware_type,omitempty"`
	HardWareRevision int  `json:"hardware_revision,omitempty"`
	Legacy           bool `json:"legacy,omitempty"`
}

func NewPeer() m2cp.CoapPeer {
	return &Peer{
		FirmWareType:     -1,
		HardWareRevision: -1,
		SequenceNumber:   -1,
		Uptime:           -1,
		FirmwareKeys:     make([]string, 0),
	}
}

func NewPeerFromRTDData(address string, ed_address string, rtdData []byte) (m2cp.CoapPeer, error) {

	if len(rtdData) == 0 {
		return nil, fmt.Errorf("body is empty")
	}

	p, err := deserializePeer(address, ed_address, rtdData)
	if err != nil {
		return nil, fmt.Errorf("cannot deserialize peer: %v", err)
	}

	return p, nil

}

func deserializePeer(address string, edAddress string, rtdData []byte) (*Peer, error) {
	if rtdData == nil || len(rtdData) == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	p, jsonErr := tryReadRTDDataFromJSON(rtdData)
	if jsonErr == nil {
		serial, err := p.GenerateOSSerial()
		if err != nil {
			return nil, err
		}
		if p.FirmWareType == 0 {
			p.FirmWareType = -1
		}
		p.OSSerial = serial
		p.Address = address
		p.EDAddress = edAddress
		cleanUpPeer(p)

		return p, nil
	}

	p, cborErr := tryReadRTDDataFromCBOR(rtdData)
	if cborErr == nil {
		serial, err := p.GenerateOSSerial()
		if err != nil {
			return nil, err
		}
		if p.FirmWareType == 0 {
			p.FirmWareType = -1
		}
		p.OSSerial = serial
		p.Address = address
		p.EDAddress = edAddress
		cleanUpPeer(p)

		return p, nil
	}
	return nil, fmt.Errorf("couldn't parse body. Neither as CBOR nor JSON \n"+
		"JSON: %v"+
		"CBOR: %v ", jsonErr, cborErr)

}

func (s *Peer) Serialize() ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("peer is nil")
	}
	return json.Marshal(s)
}

func (s *Peer) UpdatePeer(newPeer m2cp.CoapPeer) {

	if newPeer == nil {
		return
	}

	s.Legacy = newPeer.IsLegacy()

	if newPeer.GetAddress() != "" {
		s.Address = newPeer.GetAddress()
	}

	if newPeer.GetHardwareSerial() != "" {
		s.HardwareSerial = newPeer.GetHardwareSerial()
	}

	if newPeer.GetHardWareRevision() > -1 {
		s.HardWareRevision = newPeer.GetHardWareRevision()
	}

	// exists in some legacy variants as 'board'
	if newPeer.GetDeviceModel() != "" {
		s.DeviceModel = newPeer.GetDeviceModel()
	}

	if s.IsLegacy() {
		s.HardwareModel = ""
		s.HardwareVendor = ""
		s.AppName = ""
		s.AppVersion = ""
		if newPeer.GetFirmWareType() > -1 {
			s.FirmWareType = newPeer.GetFirmWareType()
		}
	} else {
		if newPeer.GetHardwareModel() != "" {
			s.HardwareModel = newPeer.GetHardwareModel()
		}
		if newPeer.GetHardwareVendor() != "" {
			s.HardwareVendor = newPeer.GetHardwareVendor()
		}
		if newPeer.GetAppName() != "" {
			s.AppName = newPeer.GetAppName()
		}
		if newPeer.GetAppVersion() != "" {
			s.AppVersion = newPeer.GetAppVersion()
		}
		s.FirmWareType = -1
	}

	// in legacy it's called 'fwr'
	if newPeer.GetSequenceNumber() > -1 {
		s.SequenceNumber = newPeer.GetSequenceNumber()
	}

	if newPeer.GetLastSeen().After(s.LastSeen) {
		s.LastSeen = newPeer.GetLastSeen()
	}

	if newPeer.GetUptime() > -1 {
		s.Uptime = newPeer.GetUptime()
	}

	if len(newPeer.GetFirmwareKeys()) > 0 {
		s.FirmwareKeys = newPeer.GetFirmwareKeys()
	}

	if newPeer.GetEDAddress() != "" || newPeer.GetEDAddress() != "::" || newPeer.GetEDAddress() != "[::]" {
		s.EDAddress = newPeer.GetEDAddress()
	}

	if newPeer.GetMCUID() != "" {
		s.MCUID = newPeer.GetMCUID()
	}

	if newPeer.IsLegacy() {
		s.Legacy = true
	} else {
		s.Legacy = false
	}

	// always generate the OS serial after changing the peer info. Should normally never change
	if serial, err := s.GenerateOSSerial(); err == nil {
		s.OSSerial = serial
	}

}

func (s *Peer) String() string {
	if s == nil {
		return "Peer{empty}"
	}
	if s.OSSerial != "" {

		if s.MCUID != "" {
			return fmt.Sprintf("Peer{OS Serial: %s, HW Serial: %s , Address: %s, Device Model: %s, MCU ID: %s, Sequence Number: %d}", s.OSSerial, s.HardwareSerial, s.Address, s.GetDeviceModel(), s.MCUID, s.SequenceNumber)
		}

		return fmt.Sprintf("Peer{OS Serial: %s, HW Serial: %s , Address: %s, Device Model: %s, Sequence Number: %d}", s.OSSerial, s.HardwareSerial, s.Address, s.GetDeviceModel(), s.SequenceNumber)

	}

	return fmt.Sprintf("Peer{ HW Serial %s,  Address: %s, Firmware Type: %d, Hardware Revision: %d, Sequence Number: %d}", s.HardwareSerial, s.Address, s.FirmWareType, s.HardWareRevision, s.SequenceNumber)
}

func (s *Peer) HasDetails() bool {
	if s == nil {
		return false
	}

	if s.Legacy {
		if s.DeviceModel != "" {
			return s.OSSerial != "" && s.Address != "" && s.HardwareSerial != "" && s.SequenceNumber != -1
		} else if s.FirmWareType != -1 {
			return s.OSSerial != "" && s.Address != "" && s.HardwareSerial != "" && s.SequenceNumber != -1
		}
		return false

	}

	// everything must be set excluding legacy fields, uptime, firmware keys, last seen
	return s.OSSerial != "" && s.Address != "" && s.HardwareSerial != "" && s.HardwareModel != "" && s.HardwareVendor != "" && s.DeviceModel != "" && s.SequenceNumber != -1 && s.AppName != "" && s.AppVersion != ""

}

func (s *Peer) GenerateOSSerial() (string, error) {

	if s == nil {
		return "", fmt.Errorf("peer is nil")
	}
	if s.HardwareSerial == "" {
		return "", fmt.Errorf("hardware serial is not set or invalid")
	} else if hws, err := strconv.ParseUint(s.HardwareSerial, 10, 32); err == nil && uint32(hws) == ^uint32(0) {
		return "00000000-0000-0000-0000-000000000000", nil
	}

	if s.Legacy {

		nsp, err := uuid.Parse(namespace)
		if err != nil {
			return "", err
		}

		serial := uuid.NewSHA1(nsp, []byte(fmt.Sprintf("mlpa-legacy$%s", s.HardwareSerial)))

		s.OSSerial = serial.String()

		return serial.String(), nil

	}

	if s.HardwareVendor == "" {
		return "", fmt.Errorf("hardware vendor is not set")
	}
	if s.HardwareModel == "" {
		return "", fmt.Errorf("hardware model is not set")
	}
	if s.DeviceModel == "" {
		return "", fmt.Errorf("device model is not set")
	}

	nsp, err := uuid.Parse(namespace)
	if err != nil {
		return "", err
	}

	var serial uuid.UUID
	if s.MCUID != "" {
		serial = uuid.NewSHA1(nsp, []byte(fmt.Sprintf("%s$%s$%s$%s$%s", s.HardwareVendor, s.HardwareModel, s.HardwareSerial, s.MCUID, s.DeviceModel)))
	} else {
		serial = uuid.NewSHA1(nsp, []byte(fmt.Sprintf("%s$%s$%s$%s", s.HardwareVendor, s.HardwareModel, s.HardwareSerial, s.DeviceModel)))
	}

	s.OSSerial = serial.String()

	return serial.String(), nil
}

func (s *Peer) GetOSSerial() string {
	if s.OSSerial == "" {
		// try to generate it
		serial, err := s.GenerateOSSerial()
		if err != nil {
			return ""
		}
		s.OSSerial = serial
	}
	return s.OSSerial
}

func (s *Peer) GetHardwareModel() string {
	return s.HardwareModel
}

func (s *Peer) GetHardwareVendor() string {
	return s.HardwareVendor
}

func (s *Peer) GetHardwareSerial() string {
	return s.HardwareSerial
}

func (s *Peer) SetHardwareSerial(serial string) {
	s.HardwareSerial = serial
}

func (s *Peer) SetDeviceModel(model string) {
	s.DeviceModel = model
}
func (s *Peer) GetDeviceModel() string {
	return s.DeviceModel
}

func (s *Peer) SetAddress(address string) {
	s.Address = address
}

func (s *Peer) GetAddress() string {
	return s.Address
}

func (s *Peer) GetMCUID() string {
	return s.MCUID
}

func (s *Peer) SetMCUID(mcuID string) {
	s.MCUID = mcuID
}

func (s *Peer) SetEDAddress(edAddress string) {
	s.EDAddress = edAddress
}

func (s *Peer) GetEDAddress() string {
	return s.EDAddress
}

func (s *Peer) SetUpdateResult(result int, resultCode string) {
	s.UpdateResult = result
	s.UpdateResultCode = resultCode
	s.UpdateResultTime = time.Now()
}

func (s *Peer) GetAppName() string {
	return s.AppName
}

func (s *Peer) GetAppVersion() string {
	return s.AppVersion
}

func (s *Peer) GetFirmWareType() int {
	return s.FirmWareType
}

func (s *Peer) GetHardWareRevision() int {
	return s.HardWareRevision
}

func (s *Peer) SetSequenceNumber(seq int) {
	s.SequenceNumber = seq
}

func (s *Peer) GetSequenceNumber() int {
	return s.SequenceNumber
}

func (s *Peer) GetLastSeen() time.Time {
	return s.LastSeen
}

func (s *Peer) SetLastSeenNow() {
	s.LastSeen = time.Now()
}

func (s *Peer) GetUptime() int64 {
	return s.Uptime
}

func (s *Peer) SetUptime(uptime int64) {
	s.Uptime = uptime
}

func (s *Peer) SetLegacy(legacy bool) {
	s.Legacy = legacy
}

func (s *Peer) IsLegacy() bool {
	return s.Legacy
}

func (s *Peer) AddFirmwareKey(key string) {
	if s.FirmwareKeys == nil {
		s.FirmwareKeys = make([]string, 0)
	}
	s.FirmwareKeys = append(s.FirmwareKeys, key)
}

func (s *Peer) SetFirmwareKeys(keys []string) {
	s.FirmwareKeys = keys
}

func (s *Peer) GetFirmwareKeys() []string {
	return s.FirmwareKeys
}

func tryReadRTDDataFromJSON(body []byte) (*Peer, error) {
	if body == nil || len(body) == 0 {
		return nil, fmt.Errorf("body is empty")
	}

	var peer Peer
	if err := json.Unmarshal(body, &peer); err != nil {
		return nil, fmt.Errorf("cannot parse body: %v", err)
	}

	return &peer, nil
}

func tryReadRTDDataFromCBOR(body []byte) (*Peer, error) {

	type intermediatePeer struct {
		HardwareSerial string   `cbor:"0,keyasint"`
		HardwareModel  string   `cbor:"1,keyasint"`
		HardwareVendor string   `cbor:"2,keyasint"`
		DeviceModel    string   `cbor:"3,keyasint"`
		AppName        string   `cbor:"4,keyasint"`
		AppVersion     string   `cbor:"5,keyasint"`
		SequenceNumber uint     `cbor:"6,keyasint"`
		Uptime         uint     `cbor:"7,keyasint"`
		FirmwareKeys   [][]byte `cbor:"8,keyasint"`
		MCUID          string   `cbor:"9,keyasint"`
	}

	if body == nil || len(body) == 0 {
		return nil, fmt.Errorf("body is empty")
	}

	var peerData intermediatePeer
	if err := cbor.Unmarshal(body, &peerData); err != nil {
		return nil, fmt.Errorf("cannot parse body: %v", err)
	}

	keys := make([]string, 0)
	for _, key := range peerData.FirmwareKeys {
		if key != nil && len(key) > 0 {
			keys = append(keys, base64.StdEncoding.EncodeToString(key))
		}
	}

	p := &Peer{
		HardwareSerial: peerData.HardwareSerial,
		HardwareModel:  peerData.HardwareModel,
		HardwareVendor: peerData.HardwareVendor,
		MCUID:          peerData.MCUID,
		DeviceModel:    peerData.DeviceModel,
		AppName:        peerData.AppName,
		AppVersion:     peerData.AppVersion,
		SequenceNumber: int(peerData.SequenceNumber),
		Uptime:         int64(peerData.Uptime),
		FirmwareKeys:   keys,
	}

	return p, nil
}

func cleanUpPeer(p *Peer) {

	if p == nil {
		return
	}

	newKeys := make([]string, 0)
	for _, key := range p.FirmwareKeys {
		if key != "" {
			newKeys = append(newKeys, key)
		}
	}

	p.FirmwareKeys = newKeys

}
