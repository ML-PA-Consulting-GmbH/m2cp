package legacy

import (
	"context"
	"fmt"
	"m2cpcli/structs"
	"strings"
)

func GetAppByNameAndArch(ctx context.Context, name, arch string) (*GetAppByNameAndArchitectureAppsAppCollectionSegmentItemsApp, error) {
	archObj := Architecture(strings.ToUpper(arch))
	resp, err := GetAppByNameAndArchitecture(ctx, &name, &archObj)
	if err != nil {
		return nil, err
	}

	if resp.Apps == nil || resp.Apps.Items == nil || len(resp.Apps.Items) == 0 {
		return nil, fmt.Errorf("App with name %s and arch %s not found", name, arch)
	}

	return resp.Apps.Items[0], nil
}

func GetAppRevision(ctx context.Context, appId string, revision int) (*structs.AppRevision, error) {
	if revision == 0 {
		resp, err := getAppRevisionLatest(ctx, appId)
		if err != nil {
			return nil, err
		}

		if resp.AppRevisions == nil || resp.AppRevisions.Items == nil || len(resp.AppRevisions.Items) != 1 {
			return nil, fmt.Errorf("App Revision 'latest' for App %s not found", appId)
		}

		return &structs.AppRevision{
			Id:       resp.AppRevisions.Items[0].Id,
			AppId:    appId,
			Revision: resp.AppRevisions.Items[0].Revision,
			Version:  resp.AppRevisions.Items[0].Version,
			Rating:   resp.AppRevisions.Items[0].AppStatus.Name,
			RatingId: resp.AppRevisions.Items[0].AppStatus.Id,
		}, nil
	} else {
		resp, err := getAppRevisionSpecific(ctx, appId, &revision)
		if err != nil {
			return nil, err
		}

		if resp.AppRevisions == nil || resp.AppRevisions.Items == nil || len(resp.AppRevisions.Items) != 1 {
			return nil, fmt.Errorf("App Revision %d for App %s not found", revision, appId)
		}

		return &structs.AppRevision{
			Id:       resp.AppRevisions.Items[0].Id,
			AppId:    appId,
			Revision: resp.AppRevisions.Items[0].Revision,
			Version:  resp.AppRevisions.Items[0].Version,
			Rating:   resp.AppRevisions.Items[0].AppStatus.Name,
			RatingId: resp.AppRevisions.Items[0].AppStatus.Id,
		}, nil
	}
}

func SetAppRevisionRating(ctx context.Context, appRevisionId, rating, description string) error {
	appStatusId, found := GetAppStatusIdByName(rating)
	if !found {
		return fmt.Errorf("invalid rating '%s'", rating)
	}

	_, err := setAppRevisionStatus(ctx, appRevisionId, appStatusId, description)
	if err != nil {
		return err
	}

	return nil
}
