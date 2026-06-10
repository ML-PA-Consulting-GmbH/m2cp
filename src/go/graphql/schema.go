package graphql

// The AssertionPlain type is a return type that is used as result of two GraphQL queries: snapAssertions and storeAssertions.
// snapAssertions will probably soonish change to return the Assertion type instead of the AssertionPlain type.
type AssertionPlain struct {
	Assertion string `json:"assertion"`
	Revision  int    `json:"revision"`
}

type Assertion struct {
	Id            UUID          `json:"id"`
	AssertionBody string        `json:"assertionBody"`
	Revision      int           `json:"revision"`
	CreatedAt     DateTime      `json:"createdAt"`
	AssertionType AssertionType `json:"assertionType"`
}

type AssertionType struct {
	Id                UUID   `json:"id"`
	AssertionTypeName string `json:"assertionTypeName"`
}

// The DateTime scalar represents an ISO-8601 compliant date time type.
type DateTime string

type EdgeDevice struct {
	Id                 UUID             `json:"id" yaml:"-"` // The id must not be "-"! It is used for unmarshalling the queries for the `id`.
	DeviceSerial       string           `json:"deviceSerial" yaml:"deviceSerial,omitempty"`
	DeviceArchitecture string           `json:"deviceArchitecture" yaml:"deviceArchitecture,omitempty"`
	DeviceName         string           `json:"deviceName" yaml:"deviceName"`
	DeviceDescription  string           `json:"deviceDescription" yaml:"deviceDescription,omitempty"`
	DeviceLastActivity DateTime         `json:"deviceLastActivity" yaml:"deviceLastActivity,omitempty"`
	IsDeviceActivated  bool             `json:"isDeviceActivated" yaml:"isDeviceActivated,omitempty"`
	FleetId            UUID             `json:"fleetId" yaml:"-"`
	Fleet              *Fleet           `json:"fleet,omitempty" yaml:"fleet,flow"`
	EdgeDeviceModel    *EdgeDeviceModel `json:"edgeDeviceModel,omitempty" yaml:"edgeDeviceModel,flow"`
}

type EdgeDeviceModelAssertion struct {
	AssertionBody string `json:"assertionBody"`
}

type EdgeDeviceModel struct {
	Id                       UUID                      `json:"id"`
	Architecture             string                    `json:"modelArchitecture"`
	Name                     string                    `json:"modelName"`
	Revision                 int                       `json:"modelRevision"`
	Type                     string                    `json:"modelType"`
	IsTpmRequired            bool                      `json:"isTpmRequired" yaml:"isTpmRequired"`
	UploadMessage            string                    `json:"uploadMessage"`
	EdgeDeviceModelAssertion *EdgeDeviceModelAssertion `json:"modelAssertion,omitempty" yaml:"modelAssertion,omitempty"`
}

type EdgeDeviceModelBridgeSnapDeclaration struct {
	Id                UUID            `json:"id"`
	SnapDeclarationId UUID            `json:"snapDeclarationId"`
	SnapDeclaration   SnapDeclaration `json:"snapDeclaration"`
}

type EdgeDeviceModifySubset struct {
	DeviceSerial      string `json:"deviceSerial" yaml:"deviceSerial"`
	DeviceName        string `json:"deviceName" yaml:"deviceName"`
	DeviceDescription string `json:"deviceDescription" yaml:"deviceDescription"`
	FleetId           UUID   `json:"fleetId" yaml:"fleetId"`
}

type EdgeDeviceCollectionSegment struct {
	Items      []EdgeDevice          `json:"items"`
	TotalCount int                   `json:"totalCount"`
	PageInfo   CollectionSegmentInfo `json:"pageInfo"`
}

type EdgeDeviceOnlineStatus struct {
	DeviceSerial       string   `json:"deviceSerial"`
	ConnectionState    string   `json:"connectionState"`
	IsDeviceRegistered bool     `json:"isDeviceRegistered"`
	LastActivityTime   DateTime `json:"lastActivityTime"`
}

type EdgeDevicePingOutput struct {
	Uptime int `json:"uptimeSeconds"`
}

type EdgeDeviceRefreshOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type EdgeDeviceSshOpenOutput struct {
	Host           string `json:"host"`
	RemoteSshPort  string `json:"remoteSshPort"`
	ReverseSshPort string `json:"reverseSshPort"`
	UsernameDevice string `json:"usernameDevice"`
	UsernameServer string `json:"usernameServer"`
}

type EdgeDeviceSshCloseOutput struct {
	Message string `json:"message"`
	Port    string `json:"port"`
}

type EdgeDeviceSystemStatsOutput struct {
	LogsUsageBytes uint64 `json:"LogsUsageBytes"`
	Statistics     struct {
		DiskInformation []struct {
			PartitionName      string `json:"partitionName"`
			PartitionSizeBytes uint64 `json:"partitionSizeBytes"`
			PartitionUsedBytes uint64 `json:"partitionUsedBytes"`
		} `json:"diskInformation"`
		OverallCpuUsagePercentage float64 `json:"overallCpuUsagePercentage"`
		OverallMemoryTotalBytes   uint64  `json:"overallMemoryTotalBytes"`
		OverallMemoryUsedBytes    uint64  `json:"overallMemoryUsedBytes"`
		ProcessInformation        []struct {
			ProcessCpuUsagePercentage float64 `json:"processCpuUsagePercentage"`
			ProcessMemoryUsedBytes    uint64  `json:"processMemoryUsedBytes"`
			ProcessName               string  `json:"processName"`
		} `json:"processInformation"`
	} `json:"statistics"`
}

type EdgeDeviceInstallationState struct {
	EdgeDevice     EdgeDevice   `json:"edgeDevice"`
	EdgeDeviceId   UUID         `json:"edgeDeviceId"`
	Id             UUID         `json:"id" yaml:"-"` // The id must not be "-"! It is used for unmarshalling the queries for the `id`.
	SnapRevision   SnapRevision `json:"snapRevision"`
	SnapRevisionId UUID         `json:"snapRevisionId"`
	TenantId       UUID         `json:"tenantId"`
}

type ExecuteRpcInput struct {
	Address    string                     `json:"address" yaml:"address"`
	Command    string                     `json:"command" yaml:"command"`
	Parameters []ExecuteRpcParameterInput `json:"parameters" yaml:"parameters"`
}

type ExecuteRpcParameterInput struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

type ExecuteRpcOutput struct {
	CommandId string                      `json:"commandId" yaml:"commandId"`
	Responses []ExecuteRpcCommandResponse `json:"responses" yaml:"responses"`
}

type ExecuteRpcCommandResponse struct {
	Error   int                       `json:"error" yaml:"error"`
	Message string                    `json:"message" yaml:"message"`
	Results []ExecuteRpcCommandResult `json:"results" yaml:"results"`
}

type ExecuteRpcCommandResult struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

type Fleet struct {
	Id            UUID              `json:"id" yaml:"-"`
	FleetName     string            `json:"fleetName"`
	Description   string            `json:"description"`
	Architecture  string            `json:"architecture"`
	Admin         User              `json:"ownerUser,omitempty"`
	CoAdmins      []User            `json:"fleetAdministrators,omitempty"`
	Model         EdgeDeviceModel   `json:"edgeDeviceModel,omitempty"`
	EdgeDevices   []EdgeDevice      `json:"edgeDevices,omitempty"`
	SnapRevisions []SnapRevision    `json:"snapRevisions,omitempty"`
	CoreSnaps     []SnapDeclaration `json:"coreSnapIds,omitempty"`
}

type FleetCollectionSegment struct {
	Items      []Fleet               `json:"items"`
	TotalCount int                   `json:"totalCount"`
	PageInfo   CollectionSegmentInfo `json:"pageInfo"`
}

type FleetBridgeSnapRevision struct {
	Id             UUID         `json:"id,omitempty"`
	FleetId        UUID         `json:"fleetId,omitempty"`
	SnapRevisionId UUID         `json:"snapRevisionId,omitempty"`
	SnapRevision   SnapRevision `json:"snapRevision,omitempty"`
	Fleet          Fleet        `json:"fleet,omitempty"`
}

type FleetBridgeSnapRevisionCollectionSegment struct {
	Items      []FleetBridgeSnapRevision `json:"items"`
	TotalCount int                       `json:"totalCount"`
	PageInfo   CollectionSegmentInfo     `json:"pageInfo"`
}

type LogQueryOutput struct {
	OperationName             string   `json:"operationName"`
	OperationId               string   `json:"operationId"`
	OperationParentId         string   `json:"operationParentId"`
	Duration                  float32  `json:"duration"`
	SeverityLevel             int      `json:"severityLevel"`
	Message                   string   `json:"message"`
	ItemType                  string   `json:"itemType"`
	TaskId                    string   `json:"taskId"`
	ApplicationName           string   `json:"applicationName"`
	Name                      string   `json:"name"`
	Source                    string   `json:"source"`
	DateTime                  DateTime `json:"dateTime"`
	ExceptionType             string   `json:"exceptionType"`
	ExceptionOuterType        string   `json:"exceptionOuterType"`
	ExceptionOuterMessage     string   `json:"exceptionOuterMessage"`
	ExceptionFormattedMessage string   `json:"exceptionFormattedMessage"`
	Success                   string   `json:"success"`
}

type Model struct {
	Name         string `json:"name"`
	Revision     int    `json:"revision"`
	Architecture string `json:"architecture"`
}

type ModelsOutput struct {
	Models []Model `json:"models"`
}

type PushModelAssertionOutput struct {
	ModelAssertion string `json:"modelAssertion"`
	Revision       int    `json:"revision"`
}

type PushSnapOutput struct {
	Architecture      string `json:"architecture"`
	FileName          string `json:"fileName"`
	Revision          int    `json:"revision"`
	SnapDeclarationId string `json:"snapDeclarationId"`
	SnapName          string `json:"snapName"`
	SnapRevisionId    string `json:"snapRevisionId"`
	Version           string `json:"version"`
}

type SnapDeclaration struct {
	Id                     UUID           `json:"id" yaml:"-"` // The id must not be "-"! It is used for unmarshalling the queries for the `id`.
	SnapId                 string         `json:"snapId" yaml:"snapId,omitempty"`
	SnapBase               string         `json:"snapBase" yaml:"snapBase,omitempty"`
	SnapDeviceArchitecture string         `json:"snapDeviceArchitecture" yaml:"snapDeviceArchitecture,omitempty"`
	SnapName               string         `json:"snapName" yaml:"snapName,omitempty"`
	SnapDescription        string         `json:"snapDescription" yaml:"snapDescription,omitempty"`
	SnapSummary            string         `json:"snapSummary" yaml:"snapSummary,omitempty"`
	TenantId               UUID           `json:"tenantId" yaml:"tenantId,omitempty"`
	Tenant                 *Tenant        `json:"tenant" yaml:"tenant,omitempty"`
	CreatedAt              DateTime       `json:"createdAt" yaml:"createdAt,omitempty"`
	AssertionId            UUID           `json:"assertionId" yaml:"assertionId"`
	SnapRevisions          []SnapRevision `json:"snapRevisions,omitempty" yaml:"snapRevisions,omitempty"`
}

type SnapDeclarationCollectionSegment struct {
	Items      []SnapDeclaration     `json:"items"`
	TotalCount int                   `json:"totalCount"`
	PageInfo   CollectionSegmentInfo `json:"pageInfo"`
}

type SnapDownloadOutput struct {
	DownloadUrl string `json:"downloadUrl"`
}

type SnapRevision struct {
	Id                UUID            `json:"id" yaml:"-" `
	UploadMessage     string          `json:"uploadMessage"`
	Revision          int32           `json:"revision"`
	SnapDownloadSize  int64           `json:"snapDownloadSize"`
	CreatedAt         DateTime        `json:"createdAt"`
	SnapVersion       string          `json:"snapVersion"`
	SnapSha3          string          `json:"snapSha3"`
	SnapDeclarationId UUID            `json:"snapDeclarationId"`
	SnapDeclaration   SnapDeclaration `json:"snapDeclaration"`
	SnapRatingId      UUID            `json:"snapStatusId"`
	SnapRating        SnapRating      `json:"snapStatus,omitempty"`
	Description       string          `json:"description,omitempty"`

	// fleetBridgeSnapRevision is for receiving DB output, Fleets for reformated info
	FleetBridgeSnapRevision []FleetBridgeSnapRevision `json:"fleetBridgeSnapRevisions,omitempty"`
	Fleets                  []Fleet                   `json:"fleets,omitempty"`
}

type SnapRevisionsOutput struct {
	SnapRevisions []SnapRevision `json:"items"`
}

type SnapRating struct {
	Id            UUID           `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	SnapRevisions []SnapRevision `json:"snapRevisions,omitempty"`
}

type SnapStatusesOutput struct {
	SnapRatings []SnapRating `json:"items"`
}

type SnapStoreVersionOutput struct {
	Environment               string `json:"environment"`
	Version                   string `json:"version"`
	CoreVersion               string `json:"coreVersion"`
	DatabaseVersion           string `json:"databaseVersion"`
	CompatibleDatabaseVersion string `json:"compatibleDatabaseVersion"`
	FriendlyVersion           string `json:"friendlyVersion"`
}

type StoreInitOutput struct {
	Messages []string `json:"messages"`
}

type SystemUserAssertionCreateInput struct {
	Email          string   `json:"email"`
	Name           string   `json:"name"`
	Username       string   `json:"username"`
	HashedPassword string   `json:"hashedPassword"`
	Models         []string `json:"models"`
	SshKeys        []string `json:"sshKeys"`
}

type Tenant struct {
	Id         UUID   `json:"id"`
	TenantName string `json:"tenantName"`
	Alias      string `json:"alias"`
}

type TenantCollectionSegment struct {
	Items      []Tenant              `json:"items"`
	TotalCount int                   `json:"totalCount"`
	PageInfo   CollectionSegmentInfo `json:"pageInfo"`
}

//type SnapUploadInput struct {
//	ContinuesToken string `json:"continuesToken"`
//	EncodedData    string `json:"encodedData"`
//	FileSize       int    `json:"fileSize"`
//}

type SnapUploadUrl struct {
	UploadUrl      string `json:"uploadUrl"`
	ContinuesToken string `json:"continuesToken"`
}

type UploadSnapOutput struct {
	ContinuesToken string `json:"continuesToken"`
}

type User struct {
	Id           UUID   `json:"id"`
	DisplayName  string `json:"displayName"`
	Email        string `json:"email"`
	SshPublicKey string `json:"sshPublicKey"`
}

type UserLoginOutput struct {
	Challenge    string `json:"challenge"`
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Success      bool   `json:"success"`
}

type UserLoginChallengeResponseOutput struct {
	Token        string `json:"token"`
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Success      bool   `json:"success"`
}

type UserLogoutOutput struct {
	ErrorMessage string `json:"errorMessage"`
	Success      bool   `json:"success"`
}

type UserSetting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UserSwitchTenantOutput struct {
	Success      bool   `json:"success"`
	Token        string `json:"token"`
	ErrorMessage string `json:"errorMessage"`
}

type VirtualDeviceContainerRegistryCredentialsOutput struct {
	ContainerRegistryUri string `json:"containerRegistryUri"`
	Username             string `json:"username"`
	Password             string `json:"password"`
}

type UUID string
