package firmwaredb

/**
Data types for syncing firmware from firmware-provider-snaps to m2cp-coap.
*/

// FirmwareSyncItem is used to offer firmware data to m2cp-coap
type FirmwareSyncItem struct {
	Fwt           uint
	Hwr           uint
	Fwr           uint
	MetaHash      string
	ManifestHash  string
	Firmware0Hash string
	Firmware1Hash string
}

// FirmwareSyncData is used to push firmware data to m2cp-coap
type FirmwareSyncData struct {
	Fwt       uint
	Hwr       uint
	Fwr       uint
	Meta      []byte
	Manifest  []byte
	Firmware0 []byte
	Firmware1 []byte
}
