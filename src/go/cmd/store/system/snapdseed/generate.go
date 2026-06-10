package snapdseed

import (
	"context"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/helper"
	"m2cpcli/tools"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Local aliases for generated backend types used in returns
type modelInfoT = backend.GetDeviceModelRevisionInfoByIdDeviceModelRevision
type appRevItemT = backend.GetAppRevisionsByAppIdAppRevisionsAppRevisionCollectionSegmentItemsAppRevision

var generateCmd = &cobra.Command{
	Use:   "generate [templateYamlPath outputPath]",
	Short: "Generate snapd seed artifacts from a template.yaml",
	Args:  cobra.ExactArgs(2),
	RunE:  runSnapdSeedGenerateCmd,
}

// cleanupOrphans removes files in assertions/ and snaps/ that do not correspond to the current template apps
func cleanupOrphans(seedRoot string, tpl *SeedTemplate) error {
	assertionsDir := filepath.Join(seedRoot, "assertions")
	snapsDir := filepath.Join(seedRoot, "snaps")

	// Build expected file name sets
	expectedAssertions := map[string]struct{}{
		"model.assert": {},
		"store.assert": {},
	}
	expectedSnaps := make(map[string]struct{}, len(tpl.Snaps))
	for _, s := range tpl.Snaps {
		expectedAssertions[fmt.Sprintf("%s_%d.assert", s.Name, s.Revision)] = struct{}{}
		expectedSnaps[fmt.Sprintf("%s_%d.snap", s.Name, s.Revision)] = struct{}{}
	}

	// Clean assertions dir
	if entries, err := os.ReadDir(assertionsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if _, ok := expectedAssertions[name]; !ok && strings.HasSuffix(name, ".assert") {
				_ = os.Remove(filepath.Join(assertionsDir, name))
			}
		}
	}

	// Clean snaps dir
	if entries, err := os.ReadDir(snapsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if _, ok := expectedSnaps[name]; !ok && strings.HasSuffix(name, ".snap") {
				_ = os.Remove(filepath.Join(snapsDir, name))
			}
		}
	}
	return nil
}

// downloadSnaps downloads each snap binary to snaps/<app-name>_<revision>.snap
func downloadSnaps(ctx context.Context, snapsDir string, tpl *SeedTemplate, appRevs map[string]*appRevItemT) error {
	for _, s := range tpl.Snaps {
		it, ok := appRevs[s.Name]
		if !ok {
			return fmt.Errorf("missing app revision info for %s", s.Name)
		}
		// Get download URL by app revision ID
		data, err := backend.GetAppDownloadUrl(ctx, it.Id)
		if err != nil {
			return err
		}
		downloadUrl := data.DownloadAppRevision.DownloadUrl
		// Determine expected filename and path
		outFile := filepath.Join(snapsDir, fmt.Sprintf("%s_%d.snap", s.Name, s.Revision))

		// If expected file already exists, skip download
		if st, statErr := os.Stat(outFile); statErr == nil && !st.IsDir() {
			if !format.JsonOutputMode {
				fmt.Println("Skipping download of", outFile, "as it already exists")
			}
			continue
		}

		// Remove stale revisions for the same app (e.g., <app>_*.snap except expected)
		pattern := filepath.Join(snapsDir, fmt.Sprintf("%s_*.snap", s.Name))
		if matches, globErr := filepath.Glob(pattern); globErr == nil {
			for _, m := range matches {
				if m != outFile {
					if !format.JsonOutputMode {
						fmt.Println("Removing stale revision", m)
					}
					_ = os.Remove(m)
				}
			}
		}

		// Download new file with progress and info
		if !format.JsonOutputMode {
			fmt.Printf("Downloading %s revision %d -> %s\n", s.Name, s.Revision, outFile)
		}
		showProgress := !format.JsonOutputMode
		if _, err := helper.DownloadSnap(downloadUrl, outFile, "", showProgress); err != nil {
			return err
		}
	}
	return nil
}

// getStoreAssertionsBody fetches store assertions and concatenates them into a single string
func getStoreAssertionsBody(ctx context.Context) (string, error) {
	resp, err := backend.GetStoreAssertions(ctx)
	if err != nil {
		return "", err
	}
	if resp == nil || resp.StoreAssertionsV2 == nil || resp.StoreAssertionsV2.Assertions == nil {
		return "", fmt.Errorf("no store assertions returned")
	}
	var b strings.Builder
	for _, it := range resp.StoreAssertionsV2.Assertions {
		if it.Assertion != nil {
			b.WriteString(*it.Assertion)
			if !strings.HasSuffix(*it.Assertion, "\n\n") {
				b.WriteString("\n")
			}
		}
	}
	// Ensure file ends with exactly one trailing newline
	body := b.String()
	body = strings.TrimRight(body, "\n") + "\n"
	return body, nil
}

// getModelAssertionBody fetches the model assertion by the assertion id in the model info
func getModelAssertionBody(ctx context.Context, model *modelInfoT) (string, error) {
	if model.ModelAssertionId == nil || *model.ModelAssertionId == "" {
		return "", fmt.Errorf("model assertion id is missing in model revision")
	}
	resp, err := backend.GetAssertionById(ctx, *model.ModelAssertionId)
	if err != nil {
		return "", err
	}
	if resp == nil || resp.Assertion == nil {
		return "", fmt.Errorf("model assertion not found: %s", *model.ModelAssertionId)
	}
	return resp.Assertion.AssertionBody, nil
}

// writeAssertions writes model and store assertions into the assertions directory
func writeAssertions(assertionsDir string, modelAssertionBody string, storeAssertionsBody string) error {
	// Write/replace only if content differs
	if err := writeIfChanged(filepath.Join(assertionsDir, "model.assert"), modelAssertionBody); err != nil {
		return err
	}
	if err := writeIfChanged(filepath.Join(assertionsDir, "store.assert"), storeAssertionsBody); err != nil {
		return err
	}
	return nil
}

// writeAppAssertions writes for each app a file `<app-name>_<revision>.assert` that contains
// the snap revision assertion and the snap declaration assertion separated by a single newline.
func writeAppAssertions(assertionsDir string, tpl *SeedTemplate) error {
	for _, s := range tpl.Snaps {
		expectedName := fmt.Sprintf("%s_%d.assert", s.Name, s.Revision)
		expectedPath := filepath.Join(assertionsDir, expectedName)
		// Build expected content: revision assertion + blank line + declaration assertion
		// Normalize trailing newlines to ensure exactly one blank line between blocks
		body := strings.TrimRight(s.snapRevisionAssertion, "\n") + "\n\n" + s.snapDeclarationAssertion

		// Remove stale revisions for the same app
		pattern := filepath.Join(assertionsDir, fmt.Sprintf("%s_*.assert", s.Name))
		if matches, globErr := filepath.Glob(pattern); globErr == nil {
			for _, m := range matches {
				if filepath.Base(m) != expectedName {
					_ = os.Remove(m)
				}
			}
		}

		// Write only if content differs or file missing
		if err := writeIfChanged(expectedPath, body); err != nil {
			return err
		}
	}
	return nil
}

// writeIfChanged writes content to path only if the file does not exist or content differs
func writeIfChanged(path string, content string) error {
	if data, err := os.ReadFile(path); err == nil {
		if string(data) == content {
			return nil
		}
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func runSnapdSeedGenerateCmd(cmd *cobra.Command, args []string) error {
	tplPath := args[0]
	outPath := args[1]

	data, err := tools.ReadLocalFile(tplPath)
	if err != nil {
		return err
	}

	var tpl SeedTemplate
	if err := yaml.Unmarshal(data, &tpl); err != nil {
		return fmt.Errorf("invalid template yaml: %w", err)
	}

	// Validate store URL matches current login/store
	currentStore := viper.Get("store").(string)
	if tpl.Store.URL != "" && tpl.Store.URL != currentStore {
		return fmt.Errorf("store mismatch: logged into '%s' but seed template expects store '%s'", currentStore, tpl.Store.URL)
	}

	// Verify model and snaps exist, and annotate internal revisions; also retrieve model/app revision details
	model, appRevs, err := getAndVerifyTemplate(cmd.Context(), &tpl)
	if err != nil {
		return err
	}

	seedRoot, err := createFolderStructure(outPath)
	if err != nil {
		return err
	}

	if err := createSeedYaml(&tpl, seedRoot); err != nil {
		return err
	}

	// Fetch and write assertions
	modelAssertBody, err := getModelAssertionBody(cmd.Context(), model)
	if err != nil {
		return err
	}
	storeAssertBody, err := getStoreAssertionsBody(cmd.Context())
	if err != nil {
		return err
	}
	assertionsDir := filepath.Join(seedRoot, "assertions")
	// Full cleanup of orphan artifacts for apps no longer present in template
	if err := cleanupOrphans(seedRoot, &tpl); err != nil {
		return err
	}
	if err := writeAssertions(assertionsDir, modelAssertBody, storeAssertBody); err != nil {
		return err
	}
	if err := writeAppAssertions(assertionsDir, &tpl); err != nil {
		return err
	}

	// Download snap binaries into seed/snaps
	snapsDir := filepath.Join(seedRoot, "snaps")
	if err := downloadSnaps(cmd.Context(), snapsDir, &tpl, appRevs); err != nil {
		return err
	}

	cmd.Println("Seed structure created at:", seedRoot)
	return nil
}

// createFolderStructure ensures <outPath>/seed/{assertions,snaps} exists and returns the seed root path
func createFolderStructure(outPath string) (string, error) {
	seedRoot := filepath.Join(outPath, "seed")
	assertionsDir := filepath.Join(seedRoot, "assertions")
	snapsDir := filepath.Join(seedRoot, "snaps")

	if err := os.MkdirAll(assertionsDir, 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(snapsDir, 0o755); err != nil {
		return "", err
	}
	return seedRoot, nil
}

// createSeedYaml converts our template to snapd seed.yaml format and writes it to <seedRoot>/seed.yaml
func createSeedYaml(tpl *SeedTemplate, seedRoot string) error {
	type seedSnap struct {
		Name    string `yaml:"name"`
		Channel string `yaml:"channel"`
		File    string `yaml:"file"`
	}
	type seedOut struct {
		Snaps []seedSnap `yaml:"snaps"`
	}

	var outDoc seedOut
	for _, s := range tpl.Snaps {
		filename := fmt.Sprintf("%s_%d.snap", s.Name, s.Revision)
		outDoc.Snaps = append(outDoc.Snaps, seedSnap{
			Name:    s.Name,
			Channel: "stable",
			File:    filename,
		})
	}

	out, err := yaml.Marshal(&outDoc)
	if err != nil {
		return err
	}

	// Write seed.yaml idempotently
	seedYamlPath := filepath.Join(seedRoot, "seed.yaml")
	if err := writeIfChanged(seedYamlPath, string(out)); err != nil {
		return err
	}
	return nil
}

// getAndVerifyTemplate checks model and snap versions exist in backend, fills internal snap revision field,
// and returns the model info and a map of app name -> app revision item selected
func getAndVerifyTemplate(ctx context.Context, tpl *SeedTemplate) (*modelInfoT, map[string]*appRevItemT, error) {
	// Verify model exists (device type fixed to edge)
	modelRevId, err := backend.TryGetDeviceModelRevisionByTypeNameRevision(ctx, "edge", tpl.Model.Name, tpl.Model.Revision)
	if err != nil {
		return nil, nil, err
	}
	resp, err := backend.GetDeviceModelRevisionInfoById(ctx, modelRevId)
	if err != nil {
		return nil, nil, err
	}
	if resp == nil || resp.DeviceModelRevision == nil {
		return nil, nil, fmt.Errorf("device model revision '%s' not found", modelRevId)
	}
	model := resp.DeviceModelRevision
	modelArch := strings.ToLower(string(model.DeviceModel.Architecture))

	// Verify snaps and record revisions
	appRevs := make(map[string]*appRevItemT, len(tpl.Snaps))
	for i := range tpl.Snaps {
		s := &tpl.Snaps[i]
		hasVersion := s.Version != "" && s.Version != "replaceme"
		hasRevision := s.Revision > 0
		if hasVersion && hasRevision {
			return nil, nil, fmt.Errorf("snap %s: specify either version or revision, not both", s.Name)
		}
		if !hasVersion && !hasRevision {
			return nil, nil, fmt.Errorf("snap %s: must specify either version or revision", s.Name)
		}
		if s.Arch == "" {
			s.Arch = modelArch
		}
		if strings.ToLower(s.Arch) != modelArch {
			return nil, nil, fmt.Errorf("architecture mismatch for snap %s: template=%s, model=%s", s.Name, s.Arch, modelArch)
		}

		app, err := backend.GetAppByNameAndArch(ctx, s.Name, modelArch)
		if err != nil {
			return nil, nil, fmt.Errorf("app not found: %s (%s)", s.Name, modelArch)
		}

		// Page through revisions and find one matching the version or revision
		take := 100
		skip := 0
		found := false
		for {
			r, err := backend.GetAppRevisionsByAppId(ctx, app.Id, take, skip)
			if err != nil {
				return nil, nil, err
			}
			if r == nil || r.AppRevisions == nil || r.AppRevisions.Items == nil || len(r.AppRevisions.Items) == 0 {
				break
			}
			for _, it := range r.AppRevisions.Items {
				matches := (hasVersion && it.Version == s.Version) ||
					(hasRevision && it.Revision == s.Revision)
				if matches {
					s.Revision = it.Revision
					s.Version = it.Version
					// Fetch and set assertion bodies for snap declaration and app revision snap
					if it.App != nil && it.App.AppSnap != nil {
						declId := it.App.AppSnap.AssertionId
						if declId != "" {
							declResp, err := backend.GetAssertionById(ctx, declId)
							if err != nil {
								return nil, nil, err
							}
							if declResp != nil && declResp.Assertion != nil {
								s.snapDeclarationAssertion = declResp.Assertion.AssertionBody
							}
						}
					}
					if it.AppRevisionSnap != nil {
						revId := it.AppRevisionSnap.AssertionId
						if revId != "" {
							revResp, err := backend.GetAssertionById(ctx, revId)
							if err != nil {
								return nil, nil, err
							}
							if revResp != nil && revResp.Assertion != nil {
								s.snapRevisionAssertion = revResp.Assertion.AssertionBody
							}
						}
					}
					appRevs[s.Name] = it
					found = true
				}
			}
			if found {
				break
			}
			skip += take
		}
		if !found {
			if hasRevision {
				return nil, nil, fmt.Errorf("no revision found for app %s revision %d", s.Name, s.Revision)
			}
			return nil, nil, fmt.Errorf("no revision found for app %s version %s", s.Name, s.Version)
		}
	}
	return model, appRevs, nil
}

func init() {
	snapdSeedCmd.AddCommand(generateCmd)
}
