package rate

import (
	"fmt"
	"m2cpcli/backend"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var syncCmd = &cobra.Command{
	Use:   "sync <release-bundle-path>",
	Short: "Sync app ratings from a release bundle's release_data.yaml files to the app store",
	Long: `Iterates over all packages in a release bundle, reads each package's release_data.yaml,
and sets the app store rating for each app listed in the appstore section.`,
	Args: cobra.ExactArgs(1),
	RunE: runSyncCmd,
}

func init() {
	rateCmd.AddCommand(syncCmd)
	syncCmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
}

type releaseData struct {
	Name     string           `yaml:"name"`
	Version  string           `yaml:"version"`
	Packages []bundledPackage `yaml:"packages"`
	Appstore *appStore        `yaml:"appstore"`
}

type bundledPackage struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type appStore struct {
	Apps []appStoreApp `yaml:"apps"`
}

type appStoreApp struct {
	Name         string `yaml:"name"`
	Type         string `yaml:"type"`
	Architecture string `yaml:"architecture"`
	Rate         string `yaml:"rate"`
}

func runSyncCmd(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	bundlePath := args[0]

	// Resolve to directory containing release_data.yaml
	info, err := os.Stat(bundlePath)
	if err != nil {
		return fmt.Errorf("cannot access %s: %w", bundlePath, err)
	}
	var bundleDir string
	if info.IsDir() {
		bundleDir = bundlePath
	} else {
		bundleDir = filepath.Dir(bundlePath)
	}

	// Load bundle release_data.yaml
	bundleData, err := loadReleaseData(filepath.Join(bundleDir, "release_data.yaml"))
	if err != nil {
		return fmt.Errorf("loading bundle release_data.yaml: %w", err)
	}

	if len(bundleData.Packages) == 0 {
		return fmt.Errorf("no packages found in bundle release_data.yaml")
	}

	cmd.Printf("Bundle: %s %s (%d packages)\n\n", bundleData.Name, bundleData.Version, len(bundleData.Packages))

	var warnings []string

	for _, pkg := range bundleData.Packages {
		pkgDir := filepath.Join(bundleDir, fmt.Sprintf("%s-%s", pkg.Name, pkg.Version))
		pkgDataPath := filepath.Join(pkgDir, "release_data.yaml")

		pkgData, err := loadReleaseData(pkgDataPath)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("  ✗ %s %s: %s", pkg.Name, pkg.Version, err))
			continue
		}

		if pkgData.Appstore == nil || len(pkgData.Appstore.Apps) == 0 {
			continue
		}

		bundleLabel := fmt.Sprintf("%s %s", bundleData.Name, bundleData.Version)
		for _, app := range pkgData.Appstore.Apps {
			if err := syncAppRating(cmd, app, pkg.Version, bundleLabel, dryRun); err != nil {
				warnings = append(warnings, fmt.Sprintf("  ✗ %s/%s (%s): %s", app.Name, app.Architecture, pkg.Version, err))
			}
		}
	}

	if len(warnings) > 0 {
		cmd.Printf("\nWarnings:\n%s\n", strings.Join(warnings, "\n"))
	}

	cmd.Printf("\n✅ Sync complete\n")
	return nil
}

func syncAppRating(cmd *cobra.Command, app appStoreApp, version string, bundleLabel string, dryRun bool) error {
	// Look up the app in the store
	storeApp, err := backend.GetAppByNameAndArch(cmd.Context(), app.Name, app.Architecture)
	if err != nil {
		return fmt.Errorf("app not found in store: %w", err)
	}

	// Find the revision matching this version by iterating revisions
	revisionId, currentRating, err := findRevisionIdByVersion(cmd, storeApp.Id, version)
	if err != nil {
		return err
	}

	if currentRating == app.Rate {
		cmd.Printf("  ▸ %s (%s) version %s — already %s\n", app.Name, app.Architecture, version, app.Rate)
		return nil
	}

	if dryRun {
		cmd.Printf("  ▸ [dry-run] %s (%s) version %s: %s → %s\n", app.Name, app.Architecture, version, currentRating, app.Rate)
		return nil
	}

	// Set the rating
	description := fmt.Sprintf("Release %s", bundleLabel)
	if err := backend.SetAppRevisionRating(cmd.Context(), revisionId, app.Rate, description); err != nil {
		return fmt.Errorf("failed setting rating: %w", err)
	}

	cmd.Printf("  ▸ %s (%s) version %s: %s → %s\n", app.Name, app.Architecture, version, currentRating, app.Rate)
	return nil
}

func findRevisionIdByVersion(cmd *cobra.Command, appId, version string) (string, string, error) {
	take := 50
	skip := 0
	for {
		resp, err := backend.GetAppRevisionsByAppId(cmd.Context(), appId, take, skip)
		if err != nil {
			return "", "", fmt.Errorf("failed listing revisions: %w", err)
		}
		if resp.AppRevisions == nil || len(resp.AppRevisions.Items) == 0 {
			break
		}
		for _, rev := range resp.AppRevisions.Items {
			if rev.Version == version {
				currentRating := ""
				if rev.AppStatus != nil {
					currentRating = rev.AppStatus.Name
				}
				return rev.Id, currentRating, nil
			}
		}
		if len(resp.AppRevisions.Items) < take {
			break
		}
		skip += take
	}
	return "", "", fmt.Errorf("no revision found for version %s", version)
}

func loadReleaseData(path string) (*releaseData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rd releaseData
	if err := yaml.Unmarshal(data, &rd); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &rd, nil
}
