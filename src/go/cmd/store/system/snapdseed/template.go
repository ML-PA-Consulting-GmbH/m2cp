package snapdseed

import (
	"context"
	"fmt"
	"m2cpcli/backend"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var templateCmd = &cobra.Command{
	Use:   "template <modelName> <modelRevision>",
	Short: "Print a YAML template for snapd seed generation",
	Args:  cobra.ExactArgs(2),
	RunE:  runSnapdSeedTemplateCmd,
}

func runSnapdSeedTemplateCmd(cmd *cobra.Command, args []string) error {
	modelName := args[0]
	rev, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("modelRevision must be an integer: %w", err)
	}

	tpl, err := GenerateSeedTemplate(cmd.Context(), modelName, rev)
	if err != nil {
		return err
	}
	out, err := yaml.Marshal(tpl)
	if err != nil {
		return err
	}
	cmd.Println(string(out))
	return nil
}

func GenerateSeedTemplate(ctx context.Context, modelName string, modelRevision int) (*SeedTemplate, error) {
	storeUrl := viper.Get("store").(string)
	// Resolve model revision by device type (edge), model name, and revision
	modelRevId, err := backend.TryGetDeviceModelRevisionByTypeNameRevision(ctx, "edge", modelName, modelRevision)
	if err != nil {
		return nil, err
	}

	// Fetch model revision details (includes bridge apps attached to this model)
	resp, err := backend.GetDeviceModelRevisionInfoById(ctx, modelRevId)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.DeviceModelRevision == nil {
		return nil, fmt.Errorf("device model revision '%s' not found", modelRevId)
	}
	info := resp.DeviceModelRevision

	arch := strings.ToLower(string(info.DeviceModel.Architecture))
	outRev := modelRevision
	if info.Revision != nil {
		outRev = *info.Revision
	}

	// Derive snap list from the model's bridge apps; resolve versions by
	// priority: stable > edge > experimental.
	snaps := make([]SeedSnap, 0, len(info.DeviceModelRevisionBridgeApps))
	for _, ba := range info.DeviceModelRevisionBridgeApps {
		if ba == nil || ba.App == nil {
			continue
		}
		appName := ba.App.AppName
		ver, err := resolveAppVersion(ctx, appName, arch)
		if err != nil {
			ver = "replaceme"
		}
		snaps = append(snaps, SeedSnap{Name: appName, Arch: arch, Version: ver})
	}

	tpl := &SeedTemplate{
		Store: SeedStore{URL: storeUrl},
		Model: SeedModel{Name: info.DeviceModel.ModelName, Revision: outRev},
		Snaps: snaps,
	}
	return tpl, nil
}

// resolveAppVersion returns the most recent version by status priority: stable, edge, experimental
func resolveAppVersion(ctx context.Context, appName, arch string) (string, error) {
	app, err := backend.GetAppByNameAndArch(ctx, appName, arch)
	if err != nil {
		return "", err
	}

	take := 100
	skip := 0
	var latestStable, latestEdge, latestExperimental string

	for {
		resp, err := backend.GetAppRevisionsByAppId(ctx, app.Id, take, skip)
		if err != nil {
			return "", err
		}
		if resp == nil || resp.AppRevisions == nil || resp.AppRevisions.Items == nil || len(resp.AppRevisions.Items) == 0 {
			break
		}

		for _, it := range resp.AppRevisions.Items {
			status := strings.ToLower(it.AppStatus.Name)
			switch status {
			case "stable":
				latestStable = it.Version
			case "edge":
				latestEdge = it.Version
			case "experimental":
				latestExperimental = it.Version
			}
		}

		// Continue paging until empty (query orders ASC by revision)
		skip += take
	}

	if latestStable != "" {
		return latestStable, nil
	}
	if latestEdge != "" {
		return latestEdge, nil
	}
	if latestExperimental != "" {
		return latestExperimental, nil
	}
	return "", fmt.Errorf("no usable revisions for app %s (%s)", appName, arch)
}

func init() {
	snapdSeedCmd.AddCommand(templateCmd)
}
