package structs

import (
	"time"
)

type App struct {
	Id           string
	Name         string
	Architecture string
	Description  string
}

type AppRevision struct {
	Id            string    `json:"id"`
	Revision      int       `json:"revision"`
	Version       string    `json:"version"`
	Description   *string   `json:"description"`
	DownloadSize  int64     `json:"downloadSize"`
	AppId         string    `json:"appId"`
	App           App       `json:"app"`
	UploadMessage string    `json:"uploadMessage"`
	Rating        string    `json:"rating"`
	RatingId      string    `json:"ratingId"`
	CreatedAt     time.Time `json:"createdAt"`
	CreatedBy     User      `json:"createdBy"`
}

type Assertion struct {
	Id                string `json:"id,omitempty"`
	AssertionBody     string `json:"assertionBody,omitempty"`
	AssertionRevision int    `json:"assertionRevision,omitempty"`
	AssertionTypeId   string `json:"assertionTypeId,omitempty"`
}

type AssertionModel struct {
	Type                string    `json:"type"`
	Series              string    `json:"series"`
	AuthorityId         string    `json:"authority-id"`
	SystemUserAuthority []string  `json:"system-user-authority,omitempty"`
	BrandId             string    `json:"brand-id"`
	Model               string    `json:"model"`
	Store               string    `json:"store"`
	Architecture        string    `json:"architecture,omitempty"` // omitempty needed for generic models
	Classic             string    `json:"classic,omitempty"`      // omitempty needed for generic models
	Timestamp           time.Time `json:"timestamp"`
	Base                string    `json:"base,omitempty"`  // omitempty needed for generic models
	Grade               string    `json:"grade,omitempty"` // omitempty needed for generic models
	Snaps               []struct {
		Name           string `json:"name"`
		Type           string `json:"type"`
		DefaultChannel string `json:"default-channel"`
		Id             string `json:"id"`
	} `json:"snaps,omitempty"` // omitempty needed for generic models
}

type AssertionType struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"assertionTypeName,omitempty"`
}

type DeviceLegacy struct {
	Id                  string `json:"id,omitempty"`
	Serial              string `json:"deviceSerial,omitempty"`
	LastActivity        string `json:"deviceLastActivity,omitempty"`
	LastKnownIp         string `json:"deviceLastKnownIp"`
	HubOverride         string `json:"deviceDeviceHubOverride"`
	Hub                 string `json:"deviceDeviceHub,omitempty"`
	BrandId             string `json:"brandId,omitempty"`
	IsActivated         bool   `json:"deviceIsActivated"`
	InstallationState   string `json:"deviceLastKnownInstallationState"`
	SshKey              string `json:"deviceSshKey"`
	Name                string `json:"deviceName"`
	Description         string `json:"deviceDescription"`
	Architecture        string `json:"deviceArchitecture"`
	SnapdActionRequired bool   `json:"deviceSnapdActionRequired"`
}

type SystemAssetProvisionInput struct {
	TenantAlias          string                             `json:"tenantAlias"`
	SystemAssetModelName string                             `json:"systemAssetModelName"`
	SystemAssetName      *string                            `json:"systemAssetName"`
	Devices              []*SystemAssetProvisionInputDevice `json:"devices,omitempty"`
}

type SystemAssetProvisionInputDevice struct {
	Type               string  `json:"type"`
	SerialNumber       string  `json:"serialNumber"`
	McuId              *string `json:"mcuId"`
	AssetModelName     string  `json:"assetModelName"`
	AssetModelRevision *int    `json:"assetModelRevision"`
	AttestationKey     *string `json:"attestationKey"`
}
type SystemAssetProvisionOutput struct {
	SystemAsset SystemAssetProvisionOutputSystem `json:"provisionSystemAsset"`
}
type SystemAssetProvisionOutputSystem struct {
	Id          string                                  `json:"id"`
	AssetName   string                                  `json:"assetName"`
	SerialNo    string                                  `json:"serialNo"`
	ChildAssets []*SystemAssetProvisionOutputChildAsset `json:"childAssets"`
}

type SystemAssetProvisionOutputChildAsset struct {
	Id                 string  `json:"id"`
	AssetName          string  `json:"assetName"`
	SerialNo           string  `json:"serialNo"`
	McuId              *string `json:"mcuId"`
	AttestationKey     *string `json:"attestationKey"`
	AttestationKeySha3 *string `json:"attestationKeySha3"`
}

type DeviceType string

const (
	DeviceTypeRealTimeDevice DeviceType = "Real-time Device"
	DeviceTypeEdgeDevice     DeviceType = "Edge Device"
	DeviceTypeUnknown        DeviceType = "Unknown"
)

var DeviceTypeAliases = map[string]DeviceType{
	"Real-time Device": DeviceTypeRealTimeDevice,
	"Real Time Device": DeviceTypeRealTimeDevice,
	"Edge Device":      DeviceTypeEdgeDevice,
}

type DeviceArchitecture string

const (
	DeviceArchitectureAmd64 DeviceArchitecture = "AMD64"
	DeviceArchitectureArm64 DeviceArchitecture = "ARM64"
	DeviceArchitectureArm32 DeviceArchitecture = "ARM32"
)

var DeviceArchitectures = []DeviceArchitecture{
	DeviceArchitectureAmd64,
	DeviceArchitectureArm64,
	DeviceArchitectureArm32,
}

type Device struct {
	DeviceId                string     `json:"deviceId"`
	DeviceType              DeviceType `json:"deviceType"`
	DeviceSerial            string     `json:"deviceSerial"`
	DeviceArchitecture      string     `json:"deviceArchitecture"`
	DeviceName              *string    `json:"deviceName"`
	Description             *string    `json:"description"`
	DeviceModelRevision     *DeviceModelRevision
	DeploymentGroup         *DeploymentGroup
	DeviceInstallStates     []DeploymentGroupAppRevision `json:"deviceInstallStates"`
	DevicePendingActions    []DevicePendingAction        `json:"devicePendingActions"`
	LastAppstoreActivity    *time.Time
	LastMessagingActivity   *time.Time
	LastUptime              *int64
	LastEdgeDevice          *Device  `json:"lastEdgeDevice,omitempty"`      // for RTDs: which ED was last seen as an uplink
	LastRealTimeDevices     []Device `json:"lastRealTimeDevices,omitempty"` // for EDs: which RTDs were observed recently?
	DeviceAssetId           *string  `json:"assetId,omitempty"`
	DeviceAssetName         *string  `json:"assetName,omitempty"`
	DeviceAttestationKey    *string  `json:"attestationKey,omitempty"`
	DeviceAssetAccess       *bool
	DeviceAssetSerial       *string
	LastHubEndpoint         *string
	UplinkMode              *string
	LastUplinkSignalContent *DeviceUplinkSignalContent
	IsOnline                *bool
}

type DeviceUplinkSignalContent struct {
	N          int `json:"n"`
	MessageHub struct {
		Up *bool `json:"up"`
	} `json:"mh"`
	Gateway struct {
		FileSystemFreeMb *int `json:"fs-free-mb"`
		SystemCpuUsage   *int `json:"cpu-pct"`
		UptimeSeconds    *int `json:"up-s"`
		MemGatewayMB     *int `json:"mem-mb"`
		SentKB           *int `json:"sent-kb"`
		BufferCount      *int `json:"q-n"`
		BufferBadKB      *int `json:"q-bad-kb"`
		BufferSizeKB     *int `json:"q-kb"`
		BufferAgeSeconds *int `json:"q-age-s"`
	} `json:"gw"`
}

func (d *Device) IsRTD() bool {
	return d.DeviceModelRevision != nil && (d.DeviceModelRevision.Type == "Real Time Device" || d.DeviceModelRevision.Type == "Real-time Device")
}

func (d *Device) IsED() bool {
	return d.DeviceModelRevision != nil && (d.DeviceModelRevision.Type == "Edge Device")
}

type DeviceInfoLegacy struct {
	Device            DeviceLegacy `json:"device,omitempty"`
	SnapSets          []SnapSet    `json:"snapSets,omitempty"`
	Snaps             []Snap       `json:"snaps,omitempty"`
	InstallationState []Snap       `json:"installationState,omitempty"`
	Action            string       `json:"action,omitempty"`
}

type DeviceModelRevision struct {
	Id           string
	Name         string
	Architecture string
	Type         string
	Revision     int
	SystemApps   []App
}
type DeploymentGroup struct {
	Id                  string
	Name                string
	Description         *string
	Architecture        string
	DeviceType          string
	Owner               User
	CoOwners            []User
	Devices             []Device
	AppRevisions        []DeploymentGroupAppRevision
	ModelRevision       *DeviceModelRevision
	PendingActionsTotal int
	AutoUpdate          string
}

type DevicePendingAction struct {
	AppName           string
	DeltaFileSize     *int64
	DownloadSize      *int64
	InstalledRevision *int
	InstalledVersion  *string
	IsCoreApp         bool
	PendingActionName string
	TargetRevision    *int
	TargetVersion     *string
}

type DeploymentGroupAppRevision struct {
	Id                       string
	AppId                    string
	AppName                  string
	AppDescription           string
	AppRevisionId            string
	AppRevision              int
	AppVersion               string
	AppRevisionDownloadSize  int64
	AppRating                string
	InstalledCount           *int
	AppRevisionUploadMessage string
	IsSystemApp              *bool
}

type Snap struct {
	Download          SnapDownload
	Declaration       SnapDeclaration
	Revision          *SnapRevision
	RevisionsConflict []SnapRevisionConflict `json:"snapRevisionsConflict,omitempty"`
	SnapSetId         string                 `json:"snapSetId,omitempty"`
	SnapSetName       string                 `json:"snapSetName,omitempty"`
	Type              string
}

type SnapRevisionConflict struct {
	SnapSetName string `json:"snapSetName"`
	SnapSetId   string `json:"snapSetId"`
	Revision    int    `json:"revision"`
}

type SnapDownload struct {
	Size int64
	Url  string
	Sha3 string
}

type SnapDeclaration struct {
	Id                 string    `json:"id,omitempty"`
	Description        string    `json:"snapDescription,omitempty"`
	SnapId             string    `json:"snapId,omitempty"`
	Name               string    `json:"snapName,omitempty"`
	Summary            string    `json:"snapSummary,omitempty"`
	Base               string    `json:"snapBase,omitempty"`
	AssertionId        string    `json:"assertionId,omitempty"`
	TypeId             string    `json:"snapTypeId,omitempty"`
	DeviceArchitecture string    `json:"snapDeviceArchitecture,omitempty"`
	BrandId            string    `json:"brandId,omitempty"`
	BrandName          string    `json:"brandName,omitempty"`
	CreationDate       time.Time `json:"snapDeclarationCreationDate,omitempty"`
}

type SnapRevision struct {
	Id                  string    `json:"id,omitempty"`
	Revision            int       `json:"snapRevision1,omitempty"`
	Version             string    `json:"snapVersion,omitempty"`
	SnapSha3            string    `json:"snapSha3,omitempty"`
	DownloadSha3        string    `json:"snapDownloadSha3,omitempty"`
	DownloadSize        int64     `json:"snapDownloadSize,omitempty"`
	DeclarationId       string    `json:"snapDeclarationId,omitempty"`
	SnapId              string    `json:"snapId,omitempty"`
	Confinement         string    `json:"snapRevisionConfinement,omitempty"`
	Summary             string    `json:"snapSummary,omitempty"`
	DeviceArchitecture  string    `json:"snapDeviceArchitecture,omitempty"`
	DownloadUrl         string    `json:"download-url,omitempty"`
	FileName            string    `json:"snapRevisionFileName,omitempty"`
	RevisionAssertionId string    `json:"revisionAssertionId,omitempty"`
	UploadMessage       string    `json:"snapRevisionUploadMessage"`
	UploadDate          time.Time `json:"snapRevisionUploadDate,omitempty"`
}

type SnapDelta struct {
	Id               string `json:"id,omitempty"`
	Format           string `json:"snapDeltaFormat,omitempty"`
	Filename         string `json:"snapDeltaFilename,omitempty"`
	Sha3             string `json:"snapDeltaSha3,omitempty"`
	Size             int64  `json:"snapDeltaSize,omitempty"`
	SourceRevisionId string `json:"snapDeltaRevisionIdSource,omitempty"`
	TargetRevisionId string `json:"snapDeltaRevisionIdTarget,omitempty"`
}

type SnapSet struct {
	Id          string        `json:"id,omitempty"`
	Name        string        `json:"snapSetName"`
	Description string        `json:"snapSetDescription"`
	BrandId     string        `json:"brandId"`
	Items       []SnapSetItem `json:"snapSetItems,omitempty"`
}

type SnapSetItem struct {
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	BrandName          string                 `json:"brandName,omitempty"`
	SnapId             string                 `json:"snapId"`
	Architecture       string                 `json:"architecture"`
	Base               string                 `json:"base"`
	Revision           int                    `json:"revision"`
	RevisionId         string                 `json:"revisionId"`
	RevisionVersion    string                 `json:"revisionVersion"`
	UploadMessage      string                 `json:"revisionUploadMessage"`
	CreationDate       time.Time              `json:"creationDate"`
	RevisionUploadDate time.Time              `json:"revisionUploadDate"`
	RevisionsConflict  []SnapRevisionConflict `json:"snapRevisionsConflict,omitempty"`
	SnapSetName        string                 `json:"snapSetName,omitempty"`
	SnapSetId          string                 `json:"snapSetId,omitempty"`
}

type SnapType struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"snapTypeName,omitempty"`
}

type Brand struct {
	Id                             string `json:"id,omitempty"`
	Name                           string `json:"brandName"`
	ManualDeviceActivationRequired bool   `json:"brandManualDeviceActivationRequired,omitempty"`
	DeviceHubDefault               string `json:"brandDeviceHubDefault,omitempty"`
	AssertionId                    string `json:"assertionId,omitempty"`
}

type User struct {
	Id         string  `json:"id,omitempty"`
	SshKey     *string `json:"developerAccountSshKey,omitempty"`
	Email      string  `json:"developerAccountSshKeyEmail,omitempty"`
	Name       string  `json:"developerAccountSshKeyname,omitempty"`
	TenantId   string  `json:"developerAccountTenantId,omitempty"`
	TenantName string  `json:"developerAccountTenantName,omitempty"`
}

type Tenant struct {
	Id    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type DeveloperAccountBridgeBrand struct {
	BrandId string `json:"brandId,omitempty"`
}

type BrandSigningKey struct {
	Id          string `json:"id,omitempty"`
	BrandId     string `json:"brandId,omitempty"`
	Name        string `json:"brandSigningKeyName,omitempty"`
	Sha3        string `json:"brandSigningKeyPublicSignKeySha3,omitempty"`
	AssertionId string `json:"assertionId,omitempty"`
}

type Log struct {
	Id      string    `json:"id,omitempty"`
	Source  string    `json:"logSource,omitempty"`
	BrandId string    `json:"brandId,omitempty"`
	Date    time.Time `json:"logDate,omitempty"`
	TaskId  string    `json:"logTaskId,omitempty"`
	Action  string    `json:"logAction,omitempty"`
	Success bool      `json:"logSuccess"`
	Result  string    `json:"logResult,omitempty"`
	// an actor is of type developer or a device with a uuid performing an action
	// on 0-2 targets. Example: A developer links a revision to a snap-set. The
	// actor is the developer, the target1 is the revision and the target2 is the
	// snap-set. Target types are then: "revision", "snap-set"
	ActorType   string `json:"logActorType,omitempty"`
	ActorId     string `json:"logActorId,omitempty"`
	Target1Type string `json:"logTarget1Type,omitempty"`
	Target1Id   string `json:"logTarget1Id,omitempty"`
	Target2Type string `json:"logTarget2Type,omitempty"`
	Target2Id   string `json:"logTarget2Id,omitempty"`
}

type CommandFeedback struct {
	Call   string `json:"call"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}
