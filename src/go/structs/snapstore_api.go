package structs

import "time"

type PostDeviceInfoResponse struct {
	Result     string           `json:"result,omitempty"`
	Message    string           `json:"message,omitempty"`
	DeviceInfo DeviceInfoLegacy `json:"deviceInfo,omitempty"`
}

type PostDeviceInfoRequest struct {
	Device string `json:"deviceSerial,omitempty"`
}

type PostMlpaDeviceModifyRequest struct {
	Serial      string   `json:"serial"`
	Fields      []string `json:"fields"` // which fields to update
	Name        string   `json:"name"`
	Description string   `json:"description"`
}

type PostMlpaDeviceModifyResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

type PostDeviceSnapSetAddRequest struct {
	DeviceSerial string `json:"deviceSerial"`
	SnapSetName  string `json:"snapSetName"`
}

type PostDeviceSnapSetAddResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

type PostDeviceSnapSetRemoveRequest struct {
	DeviceSerial string `json:"deviceSerial"`
	SnapSetName  string `json:"snapSetName"`
}

type PostDeviceSnapSetRemoveResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

type PostDevicesListResponse struct {
	Result  string         `json:"result,omitempty"`
	Message string         `json:"message,omitempty"`
	Devices []DeviceLegacy `json:"version,omitempty"`
}

type PostDevicesListRequest struct {
	LastActiveBefore string `json:"lastActiveBefore,omitempty"`
	LastActiveAfter  string `json:"lastActiveAfter,omitempty"`
	Activated        string `json:"activated,omitempty"`
	Architecture     string `json:"architecture,omitempty"`
}

type PostBindIpResult struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

type PostLogsRequest struct {
	FromTime     time.Time
	ToTime       time.Time
	Success      int    `json:"success,omitempty"`
	TaskId       string `json:"taskId,omitempty"`
	SearchAction string `json:"action,omitempty"`
	SearchResult string `json:"result,omitempty"`
	Limit        int    `json:"limit,omitempty"`
}

type PostLogsResponse struct {
	Result  string `json:"result,omitempty"`
	Message string `json:"message,omitempty"`
	Logs    []Log
}

type PostSnapIdResponse struct {
	SnapId  string `json:"snapId,omitempty"`
	Version string `json:"version,omitempty"`
}

type PostSnapIdRequest struct {
	Name         string `json:"name,omitempty"`
	Architecture string `json:"architecture,omitempty"`
}

type PostStoreInfoResult struct {
	Result             string `json:"result,omitempty"`
	Error              string `json:"error,omitempty"`
	StoreAssertionBody string `json:"store-assertion,omitempty"`
}

type PostInitStoreResult struct {
	Result string   `json:"result,omitempty"`
	Error  string   `json:"error,omitempty"`
	Log    []string `json:"log,omitempty"`
}

type PostSnapPushResponse struct {
	Result  string `json:"result"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type PostSnapSetInfoRequest struct {
	SnapSetId   string `json:"id,omitempty"`
	SnapSetName string `json:"name,omitempty"`
}

type PostSnapSetInfoResponse struct {
	Result  string  `json:"result,omitempty"`
	Message string  `json:"message,omitempty"`
	SnapSet SnapSet `json:"snapSet,omitempty"`
}

type PostSnapSetAddRequest struct {
	SnapSetId    string `json:"snapSetId,omitempty"`
	SnapSetName  string `json:"SnapSetName,omitempty"`
	SnapId       string `json:"snapId,omitempty"`
	SnapName     string `json:"snapName,omitempty"`
	SnapRevision int    `json:"snapRevision,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Mlpa         bool   `json:"mlpa,omitempty"`
	Modify       bool   `json:"modify,omitempty"`
}

type PostSnapSetAddResponse struct {
	Result  string `json:"result,omitempty"`
	Message string `json:"message,omitempty"`
}

type PostSnapSetCreateRequest struct {
	SnapSetName        string `json:"name,omitempty"`
	SnapSetDescription string `json:"description,omitempty"`
}

type PostSnapSetCreateResponse struct {
	SnapSetId string `json:"setId,omitempty"`
	Result    string `json:"result,omitempty"`
	Message   string `json:"message,omitempty"`
}

type PostSnapSetDeleteRequest struct {
	SnapSetId   string `json:"id,omitempty"`
	SnapSetName string `json:"name,omitempty"`
}

type PostSnapSetDeleteResponse struct {
	Result  string `json:"result,omitempty"`
	Message string `json:"message,omitempty"`
}

type PostSnapSetRemoveRequest struct {
	SnapSetId        string `json:"snapSetId,omitempty"`
	SnapSetName      string `json:"SnapSetName,omitempty"`
	SnapId           string `json:"snapId,omitempty"`
	SnapName         string `json:"snapName,omitempty"`
	SnapArchitecture string `json:"architecture,omitempty"`
}

type PostSnapSetRemoveResponse struct {
	Result  string `json:"result,omitempty"`
	Message string `json:"message,omitempty"`
}

type PostSnapSetsListResponse struct {
	Result   string    `json:"result,omitempty"`
	Message  string    `json:"message,omitempty"`
	SnapSets []SnapSet `json:"snapSets,omitempty"`
}

type PostSnapInfoResponse struct {
	Result      string           `json:"result,omitempty"`
	Message     string           `json:"message,omitempty"`
	Declaration *SnapDeclaration `json:"declaration,omitempty"`
	Revisions   []SnapRevision   `json:"revisions,omitempty"`
}

type PostSnapInfoRequest struct {
	SnapId       string `json:"snapId,omitempty"`
	Name         string `json:"name,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Revision     int    `json:"revision,omitempty"`
	Mlpa         bool   `json:"mlpa,omitempty"`
}

type PostSnapsListResponse struct {
	Result           string `json:"result,omitempty"`
	Message          string `json:"message,omitempty"`
	SnapDeclarations []SnapDeclaration
}

type GetVersionResult struct {
	Result      string `json:"result,omitempty"`
	Message     string `json:"message,omitempty"`
	Version     string `json:"version"`
	Build       string `json:"build"`
	CurrentDate string `json:"dateCurrent"`
}

type DeviceInfoRequest struct {
	Macaroon     string `json:"macaroon"`
	DeviceSerial string `json:"deviceSerial"`
}
type DeviceInfoResult struct {
	Users              []interface{}            `json:"users"`
	Result             string                   `json:"result"`
	Message            string                   `json:"message"`
	Nodes              []DeviceInfoNode         `json:"nodes"`
	IpAddresses        map[string]interface{}   `json:"ipaddresses"`
	ListInstalledSnaps []map[string]interface{} `json:"listedsnaps"`
}

type DeviceInfoNode struct {
	Address string `json:"address"`
}

type GetSnapSetAndRevisionListResponse struct {
	Result              string         `json:"result,omitempty"`
	SnapSetAndRevisions map[string]int `json:"snapSetAndRevision"`
}

type DeviceRefreshRequest struct {
	Macaroon     string `json:"macaroon"`
	DeviceSerial string `json:"deviceSerial"`
}

type DeviceRefreshResult struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}
