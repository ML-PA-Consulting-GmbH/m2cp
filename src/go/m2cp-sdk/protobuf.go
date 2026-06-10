package m2cp

import "time"

type ProtobufOidParser interface {
	Parse(data []byte) ([]DataRow, error)
	GetOid() uint
}

type ProtobufParserService interface {
}

type RTDMessageForwarder interface {
	GetHeartBeatStats() HeartBeatStats
	EmitLogSignal(standard RTDDataStandard, name, content, level string)
	ForwardUnparsedDataMessage(standard RTDDataStandard, oid uint, rows []DataRow)
	EmitParsedData(standard RTDDataStandard, oid uint, rows []DataRow)
	IncReceived()
	IncParsed()
	IncForwarded()
	IncFailed()
}

type HeartBeatStats struct {
	Started            time.Time `json:"-"`
	Uptime             uint32    `json:"uptime"`
	Received           uint32    `json:"msg-received"`
	Parsed             uint32    `json:"rows-parsed"`
	Forwarded          uint32    `json:"rows-forwarded"`
	ReportedDeviceInfo bool      `json:"reportedDeviceInfo"`
	Failed             uint32    `json:"errors"`
}

const (
	StandardDIN1 string = "DIN1"
	StandardDIN2 string = "DIN2"
)

// RTDDataStandard is a struct to pass to ProtobufParserSender functions to change their behaviour. Either OSSerial
// or HWSerial must be set and not both.
type RTDDataStandard struct {
	Standard string
	OSSerial string
	HWSerial string
}

func DIN1(hwSerial string) RTDDataStandard {
	return RTDDataStandard{
		Standard: StandardDIN1,
		HWSerial: hwSerial,
	}
}
func DIN2(osSerial string) RTDDataStandard {
	return RTDDataStandard{
		Standard: StandardDIN2,
		OSSerial: osSerial,
	}
}
