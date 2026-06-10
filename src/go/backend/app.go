package backend

import (
	"context"
	"m2cpcli/backend/legacy"
	"m2cpcli/structs"
)

const (
	BrokenAppStatusId       = legacy.BrokenAppStatusId
	DeniedAppStatusId       = legacy.DeniedAppStatusId
	DeprecatedAppStatusId   = legacy.DeprecatedAppStatusId
	EdgeAppStatusId         = legacy.EdgeAppStatusId
	ExperimentalAppStatusId = legacy.ExperimentalAppStatusId
	StableAppStatusId       = legacy.StableAppStatusId
	UnratedAppStatusId      = legacy.UnratedAppStatusId
	ArchitectureAmd64       = legacy.ArchitectureAmd64
	ArchitectureArm32       = legacy.ArchitectureArm32
	ArchitectureArm64       = legacy.ArchitectureArm64
)

type (
	GetSnapRevisionsByNameArchRevisionSnapRevisionsSnapRevisionCollectionSegmentItemsSnapRevision = legacy.GetSnapRevisionsByNameArchRevisionSnapRevisionsSnapRevisionCollectionSegmentItemsSnapRevision
	Architecture                                                                                  = legacy.Architecture
	AppFilterInput                                                                                = legacy.AppFilterInput
	GetAppInfoResponse                                                                            = legacy.GetAppInfoResponse
	GetAppRevisionsByAppIdResponse                                                                = legacy.GetAppRevisionsByAppIdResponse
	GetAppRevisionsByAppIdAppRevisionsAppRevisionCollectionSegmentItemsAppRevision                = legacy.GetAppRevisionsByAppIdAppRevisionsAppRevisionCollectionSegmentItemsAppRevision
	SnapRevisionUpdateStatusInput                                                                 = legacy.SnapRevisionUpdateStatusInput
	AppRevisionCompleteUploadInput                                                                = legacy.AppRevisionCompleteUploadInput
	AppRevisionInitiateUploadInput                                                                = legacy.AppRevisionInitiateUploadInput
	ArchitectureOperationFilterInput                                                              = legacy.ArchitectureOperationFilterInput
)

func GetAppStatusIdByName(appStatusName string) (string, bool) {
	return legacy.GetAppStatusIdByName(appStatusName)
}

func GetAppStatusNameById(appStatusId string) string {
	return legacy.GetAppStatusNameById(appStatusId)
}

func GetSnapRevisionByNameArchRevision(ctx context.Context, snapName, snapArchitecture string, revision int) (*legacy.GetSnapRevisionsByNameArchRevisionSnapRevisionsSnapRevisionCollectionSegmentItemsSnapRevision, error) {
	return legacy.GetSnapRevisionByNameArchRevision(ctx, snapName, snapArchitecture, revision)
}

func GetAppByNameAndArch(ctx context.Context, name, arch string) (*legacy.GetAppByNameAndArchitectureAppsAppCollectionSegmentItemsApp, error) {
	return legacy.GetAppByNameAndArch(ctx, name, arch)
}

func GetAppRevision(ctx context.Context, appId string, revision int) (*structs.AppRevision, error) {
	return legacy.GetAppRevision(ctx, appId, revision)
}

func SetAppRevisionRating(ctx context.Context, appRevisionId, rating, description string) error {
	return legacy.SetAppRevisionRating(ctx, appRevisionId, rating, description)
}

func GetAppIdsByAppNameAndArchitecture(ctx context.Context, appName string, architecture Architecture) (*legacy.GetAppIdsByAppNameAndArchitectureResponse, error) {
	return legacy.GetAppIdsByAppNameAndArchitecture(ctx, appName, architecture)
}

func GetAppInfo(ctx context.Context, appId string) (*legacy.GetAppInfoResponse, error) {
	return legacy.GetAppInfo(ctx, appId)
}

func GetAppRevisionsByAppId(ctx context.Context, appId string, take int, skip int) (*legacy.GetAppRevisionsByAppIdResponse, error) {
	return legacy.GetAppRevisionsByAppId(ctx, appId, take, skip)
}

func GetAppDownloadUrl(ctx context.Context, revisionId string) (*legacy.GetAppDownloadUrlResponse, error) {
	return legacy.GetAppDownloadUrl(ctx, revisionId)
}

func GetAppByNameAndArchitectureAndVersion(ctx context.Context, name *string, arch *Architecture, version *string) (*legacy.GetAppByNameAndArchitectureAndVersionResponse, error) {
	return legacy.GetAppByNameAndArchitectureAndVersion(ctx, name, (*legacy.Architecture)(arch), version)
}

func UpdateSnapRevisionStatus(ctx context.Context, input *SnapRevisionUpdateStatusInput) (*legacy.UpdateSnapRevisionStatusResponse, error) {
	return legacy.UpdateSnapRevisionStatus(ctx, (*legacy.SnapRevisionUpdateStatusInput)(input))
}

func GetStoreAssertions(ctx context.Context) (*legacy.GetStoreAssertionsResponse, error) {
	return legacy.GetStoreAssertions(ctx)
}

func GetAssertionById(ctx context.Context, id string) (*legacy.GetAssertionByIdResponse, error) {
	return legacy.GetAssertionById(ctx, id)
}

func GetAppList(ctx context.Context, filter *AppFilterInput, take *int, skip *int) (*legacy.GetAppListResponse, error) {
	return legacy.GetAppList(ctx, (*legacy.AppFilterInput)(filter), take, skip)
}

func InitiateAppRevisionUpload(ctx context.Context, input *AppRevisionInitiateUploadInput) (*legacy.InitiateAppRevisionUploadResponse, error) {
	return legacy.InitiateAppRevisionUpload(ctx, (*legacy.AppRevisionInitiateUploadInput)(input))
}

func CompleteAppRevisionUpload(ctx context.Context, input *AppRevisionCompleteUploadInput) (*legacy.CompleteAppRevisionUploadResponse, error) {
	return legacy.CompleteAppRevisionUpload(ctx, (*legacy.AppRevisionCompleteUploadInput)(input))
}

func GetAppsByRating(ctx context.Context, rating string) (*legacy.GetAppsByRatingResponse, error) {
	return legacy.GetAppsByRating(ctx, rating)
}
