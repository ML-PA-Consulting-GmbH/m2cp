package coap_server

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"
)

// TestSerializePeer tests the serialization of a Peer struct to JSON.
// The test was written before the MCU ID was introduced and didn't need to be changed (proof: git history)
// Thus MCU ID is backwards compatible and doesn't change the OS Serial
func (t *TestSuite) TestSerializePeer() {

	p := Peer{
		Address:        "[fd00::100]",
		HardwareSerial: "1",
		HardwareModel:  "PMIC 1.08",
		HardwareVendor: "mlpa",
		DeviceModel:    "pmic",
		AppName:        "fu-test-rtd-a",
		AppVersion:     "0.1.0",
		SequenceNumber: 23,
		LastSeen:       time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		Uptime:         50,
		FirmwareKeys:   []string{base64.StdEncoding.EncodeToString([]byte("test-key-1"))},
		FirmWareType:   0,
		Legacy:         false,
	}
	_, err := p.GenerateOSSerial()
	t.NoError(err)

	res, err := json.Marshal(p)
	t.NoError(err)

	t.Equal(`{"os_serial":"1dbccf89-1027-5a29-8395-fbfe6320e371","address":"[fd00::100]","hw_serial":"1","hw_model":"PMIC 1.08","hw_vendor":"mlpa","dev_model":"pmic","app_name":"fu-test-rtd-a","app_version":"0.1.0","seq_no":23,"last_seen":"2009-11-10T23:00:00Z","uptime":50,"keys":["dGVzdC1rZXktMQ=="],"update_result_time":"0001-01-01T00:00:00Z"}`, string(res))
}

func (t *TestSuite) TestDeserializePeer() {
	data := `{"address":"[fd00::100]","hw_serial":"1","hw_model":"PMIC 1.08","hw_vendor":"mlpa","dev_model":"pmic","app_name":"fu-test-rtd-a","app_version":"0.1.0","seq_no":23,"uptime":50,"keys":["dGVzdC1rZXktMQ=="]}`

	var p Peer
	err := json.Unmarshal([]byte(data), &p)
	t.NoError(err)

	_, err = p.GenerateOSSerial()
	t.NoError(err)

	t.Equal("1dbccf89-1027-5a29-8395-fbfe6320e371", p.OSSerial)
	t.Equal("[fd00::100]", p.Address)
	t.Equal("1", p.HardwareSerial)
	t.Equal("PMIC 1.08", p.HardwareModel)
	t.Equal("mlpa", p.HardwareVendor)
	t.Equal("pmic", p.DeviceModel)
	t.Equal("fu-test-rtd-a", p.AppName)
	t.Equal("0.1.0", p.AppVersion)
	t.Equal(23, p.SequenceNumber)
	t.Equal(int64(50), p.Uptime)
	t.Len(p.FirmwareKeys, 1)
	t.Equal("dGVzdC1rZXktMQ==", p.FirmwareKeys[0])
}
func (t *TestSuite) TestDeserializePeerWithMCUID() {
	data := `{"address":"[fd00::100]","hw_serial":"1","hw_model":"PMIC 1.08","hw_vendor":"mlpa", "mcu_id": "mcu-2", "dev_model":"pmic","app_name":"fu-test-rtd-a","app_version":"0.1.0","seq_no":23,"uptime":50,"keys":["dGVzdC1rZXktMQ=="]}`

	var p Peer
	err := json.Unmarshal([]byte(data), &p)
	t.NoError(err)

	_, err = p.GenerateOSSerial()
	t.NoError(err)

	t.Equal("249dfacf-8010-59f8-bcca-2697180602e8", p.OSSerial)
	t.Equal("mcu-2", p.MCUID)
	t.Equal("[fd00::100]", p.Address)
	t.Equal("1", p.HardwareSerial)
	t.Equal("PMIC 1.08", p.HardwareModel)
	t.Equal("mlpa", p.HardwareVendor)
	t.Equal("pmic", p.DeviceModel)
	t.Equal("fu-test-rtd-a", p.AppName)
	t.Equal("0.1.0", p.AppVersion)
	t.Equal(23, p.SequenceNumber)
	t.Equal(int64(50), p.Uptime)
	t.Len(p.FirmwareKeys, 1)
	t.Equal("dGVzdC1rZXktMQ==", p.FirmwareKeys[0])
}

// TestTryReadRTDDataFromCBOR tests the CBOR deserialization of RTD data into a Peer struct.
// Here the equality check for the OS Serial was introduced in the same commit as the MCU ID as the check was overlooked before.
// However, the matching content of the CBOR and OS Serial to TestDeserializePeer are proof that the introduction of MCU ID are non-breaking
func (t *TestSuite) TestTryReadRTDDataFromCBOR() {

	hexData := "A961306131613169504D494320312E30386132646D6C7061613364706D696361346D66752D746573742D7274642D61613565302E312E306136176137183261388244DEADBEEF44CAFEBABE"
	bytes, err := hex.DecodeString(hexData)
	t.NoError(err)

	p, err := tryReadRTDDataFromCBOR(bytes)
	t.NoError(err)
	t.NotNil(p)
	if p == nil {
		return
	}

	_, err = p.GenerateOSSerial()
	t.NoError(err)
	t.Equal("1dbccf89-1027-5a29-8395-fbfe6320e371", p.OSSerial)
	t.Equal("1", p.HardwareSerial)
	t.Equal("PMIC 1.08", p.HardwareModel)
	t.Equal("mlpa", p.HardwareVendor)
	t.Equal("pmic", p.DeviceModel)
	t.Equal("fu-test-rtd-a", p.AppName)
	t.Equal("0.1.0", p.AppVersion)
	t.Equal(23, p.SequenceNumber)
	t.Equal(int64(50), p.Uptime)
	t.Len(p.FirmwareKeys, 2)
	if len(p.FirmwareKeys) < 2 {
		return
	}
	t.Equal("3q2+7w==", p.FirmwareKeys[0])
	t.Equal("yv66vg==", p.FirmwareKeys[1])
}

func (t *TestSuite) TestTryReadRTDDataFromCBORWithMCUID() {

	//Uses the same content as TestTryReadRTDDataFromCBOR, just added mcu_id: "mcu-2"
	hexData := "AA61306131613169504D494320312E30386132646D6C7061613364706D696361346D66752D746573742D7274642D61613565302E312E306136176137183261388244DEADBEEF44CAFEBABE6139656D63752D32"

	bytes, err := hex.DecodeString(hexData)
	t.NoError(err)

	p, err := tryReadRTDDataFromCBOR(bytes)
	t.NoError(err)
	t.NotNil(p)
	if p == nil {
		return
	}

	_, err = p.GenerateOSSerial()
	t.NoError(err)

	//note that OSSerial is different to TestTryReadRTDDataFromCBOR
	t.Equal("249dfacf-8010-59f8-bcca-2697180602e8", p.OSSerial)
	t.Equal("mcu-2", p.MCUID)
	t.Equal("1", p.HardwareSerial)
	t.Equal("PMIC 1.08", p.HardwareModel)
	t.Equal("mlpa", p.HardwareVendor)
	t.Equal("pmic", p.DeviceModel)
	t.Equal("fu-test-rtd-a", p.AppName)
	t.Equal("0.1.0", p.AppVersion)
	t.Equal(23, p.SequenceNumber)
	t.Equal(int64(50), p.Uptime)
	t.Len(p.FirmwareKeys, 2)
	if len(p.FirmwareKeys) < 2 {
		return
	}
	t.Equal("3q2+7w==", p.FirmwareKeys[0])
	t.Equal("yv66vg==", p.FirmwareKeys[1])

}
