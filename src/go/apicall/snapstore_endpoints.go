package apicall

const (
	MlpaDeviceInfoCmd          = "/mlpa/device/info"
	MlpaDeviceModifyCmd        = "/mlpa/device/modify"
	MlpaDeviceSnapSetAddCmd    = "/mlpa/device/snapset/add"
	MlpaDeviceSnapSetRemoveCmd = "/mlpa/device/snapset/remove"
	MlpaDevicesList            = "/mlpa/devices/list"
	MlpaLogs                   = "/mlpa/logs"
	MlpaSnapSetAdd             = "/mlpa/set/add"
	MlpaSnapSetRemove          = "/mlpa/set/remove"
	MlpaSnapInfoCmd            = "/mlpa/snap/info"
	DownloadCmd                = "/download/{filename}"
	AssertionDeclarationCmd    = "/v2/assertions/snap-declaration/{mucVersion}/{snapId}"
	AssertionRevisionCmd       = "/v2/assertions/snap-revision/{snapSha3}"
	MlpaSnapsList              = "/mlpa/snaps/list"
)
