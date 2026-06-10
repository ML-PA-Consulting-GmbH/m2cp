package app

import (
	"context"
	"fmt"
	"m2cpcli/backend"
	v5 "m2cpcli/backend/v5"
	"m2cpcli/cmd/store/dgroup"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var fleetSnapCmd = &cobra.Command{
	Use:     "snap",
	Aliases: []string{"s"},
	Short:   "Manage Apps of a Deployment Group",
	Hidden:  true,
}

var dgroupAppCmd = &cobra.Command{
	Use:     "app",
	Aliases: []string{"a"},
	Short:   "Manage Apps of a Deployment Group",
}

func init() {
	dgroup.FleetCmd.AddCommand(fleetSnapCmd)
	dgroup.DGroupCmd.AddCommand(dgroupAppCmd)
}

func validateAppAddOrModifyCmdArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(3)(cmd, args); err != nil {
		return err
	}

	if !tools.IsValidAppId(args[1]) && !tools.IsValidAppName(args[1]) {
		return fmt.Errorf("invalid appName or appId '%s'", args[1])
	}

	return nil
}

type DGroupArgType int

const (
	DGroupArgTypeId DGroupArgType = iota
	DGroupArgTypeName
)

func ParseDGroupArgType(str string) DGroupArgType {
	if tools.IsValidUuid(str) {
		return DGroupArgTypeId
	}
	return DGroupArgTypeName
}

type AppArgType int

const (
	AppArgTypeId AppArgType = iota
	AppArgTypeName
)

func ParseAppArgType(str string) AppArgType {
	if tools.IsValidUuid(str) {
		return AppArgTypeId
	}
	return AppArgTypeName
}

type RevisionOrVersionArgType int

const (
	RevisionOrVersionArgTypeRevision RevisionOrVersionArgType = iota
	RevisionOrVersionArgTypeVersion
	RevisionOrVersionArgTypeLatest
)

func isPositiveInteger(s string) bool {
	n, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return n > 0
}

func ParseRevisionOrVersionArgType(str string) RevisionOrVersionArgType {
	if str == "latest" {
		return RevisionOrVersionArgTypeLatest
	}
	if isPositiveInteger(str) {
		return RevisionOrVersionArgTypeRevision
	}
	return RevisionOrVersionArgTypeVersion
}

func RetrieveAppIdByAnyType(ctx context.Context, dgroupArch string, appNameOrId string) (appId string, err error) {
	switch ParseAppArgType(appNameOrId) {
	case AppArgTypeId:
		appId = appNameOrId
	case AppArgTypeName:
		app, err := backend.GetAppByNameAndArch(ctx, appNameOrId, dgroupArch)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve app %s for arch %s: %s",
				appNameOrId, dgroupArch, err)
		}
		appId = app.Id
	default:
		return "", fmt.Errorf("app name or id not supported: %s", appNameOrId)
	}
	return appId, nil
}

func RetrieveAppRevisionByAnyType(ctx context.Context, appId string, revisionOrVersion string) (appRevision *structs.AppRevision, err error) {

	switch ParseRevisionOrVersionArgType(revisionOrVersion) {
	case RevisionOrVersionArgTypeLatest:
		resp, err := v5.GetAppRevisionByAppIdLatest(ctx, appId)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve latest app revision")
		}

		isPopulated := func(resp *v5.GetAppRevisionByAppIdLatestResponse) bool {
			return resp.AppRevisions != nil &&
				len(resp.AppRevisions.Items) > 0
		}
		// TODO: implement constructor from GetAppRevisionLatestAppRevisionsAppRevisionCollectionSegmentItemsAppRevision?
		if isPopulated(resp) {
			appRevision = &structs.AppRevision{
				Id:       resp.AppRevisions.Items[0].Id,
				Revision: resp.AppRevisions.Items[0].Revision,
				Version:  resp.AppRevisions.Items[0].Version,
			}
		} else {
			return nil, fmt.Errorf("failed to retrieve app revision by latest: %s", revisionOrVersion)
		}
	case RevisionOrVersionArgTypeVersion:
		// TODO: this deserves it own GraphQL call
		resp, err := v5.GetAppRevisionByAppIdAndVersion(ctx, appId, &revisionOrVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve latest app revision")
		}

		isPopulated := func(resp *v5.GetAppRevisionByAppIdAndVersionResponse) bool {
			return resp.AppRevisions != nil &&
				len(resp.AppRevisions.Items) > 0
		}
		if isPopulated(resp) {
			appRevision = &structs.AppRevision{
				Id:       resp.AppRevisions.Items[0].Id,
				Revision: resp.AppRevisions.Items[0].Revision,
				Version:  resp.AppRevisions.Items[0].Version,
			}
		} else {
			return nil, fmt.Errorf("failed to retrieve app revision by version: %s", revisionOrVersion)
		}
	case RevisionOrVersionArgTypeRevision:
		// TODO: is the convention 0 == "latest"?
		var revision int
		revision, err = strconv.Atoi(revisionOrVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid revision number: %s", revisionOrVersion)
		}
		resp, err := v5.GetAppRevisionByAppIdAndRevisionNumber(ctx, appId, &revision)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve app revision by number: %d", revision)
		}

		isPopulated := func(resp *v5.GetAppRevisionByAppIdAndRevisionNumberResponse) bool {
			return resp.AppRevisions != nil &&
				len(resp.AppRevisions.Items) > 0
		}
		if isPopulated(resp) {
			appRevision = &structs.AppRevision{
				Id:       resp.AppRevisions.Items[0].Id,
				Revision: resp.AppRevisions.Items[0].Revision,
				Version:  resp.AppRevisions.Items[0].Version,
			}
		} else {
			return nil, fmt.Errorf("failed to retrieve app revision by revision number: %s", revisionOrVersion)
		}
	default:
		return nil, fmt.Errorf("revision or version not supported: %s", revisionOrVersion)
	}
	return appRevision, nil
}

func resolve(ctx context.Context, dgroupNameOrId, appNameOrId, revisionOrVersion string) (dgroupId, appId, appRevisionId string, err error) {
	dgroup, err := backend.GetDeploymentGroupByNameOrId(ctx, dgroupNameOrId)
	if err != nil {
		return "", "", "", err
	}
	dgroupId = dgroup.Id
	appId, err = RetrieveAppIdByAnyType(ctx, dgroup.Architecture, appNameOrId)
	if err != nil {
		return "", "", "", err
	}
	appRevision, err := RetrieveAppRevisionByAnyType(ctx, appId, revisionOrVersion)
	if err != nil {
		return "", "", "", err
	}
	appRevisionId = appRevision.Id
	return dgroupId, appId, appRevisionId, nil
}

func resolveDGroupAndApp(ctx context.Context, groupNameOrId, appNameOrId string) (dgroupId, appId string, err error) {
	// first we need details on the target Deployment Group
	dgroup, err := backend.GetDeploymentGroupByNameOrId(ctx, groupNameOrId)
	if err != nil {
		return "", "", err
	}
	dgroupId = dgroup.Id

	appId, err = RetrieveAppIdByAnyType(ctx, dgroup.Architecture, appNameOrId)
	if err != nil {
		return "", "", err
	}

	return dgroupId, appId, nil
}
