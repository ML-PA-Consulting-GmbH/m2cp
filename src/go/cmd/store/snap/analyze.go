package snap

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/helper"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// snapBootstrapFields are the labels we surface from a kernel snap's
// initrd snap-bootstrap binary (matched as "<field>:").
var snapBootstrapFields = []string{
	"account-id:",
	"public-key-sha3-384:",
	"sign-key-sha3-384:",
	"display-name:",
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze (<snapName> <version> | <snapFile>)",
	Short: "Inspect a snap; for kernel snaps extract initrd identity fields",
	Long: `Inspect a snap and print its snap type. Two input modes:

  analyze <snapName> <version>   download from the store (a snap is a subtype
                                 of app) and inspect it
  analyze <snapFile>             inspect a local .snap file (no download)

The snap is unsquashed to a temporary directory and the identity fields baked
into its snap-bootstrap/snapd binary are reported (account-id,
public-key-sha3-384, sign-key-sha3-384, display-name):

  kernel snap   the initrd is extracted (unmkinitramfs) and the initrd's
                snap-bootstrap is inspected
  snapd snap    the snap-bootstrap/snapd binary in the snap is inspected
                directly

For any other type the command stops after printing the type.

Requires unsquashfs, strings (and, for kernel snaps, unmkinitramfs and cpio)
on PATH.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAnalyzeCmd,
}

func init() {
	SnapCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().String("arch", "arm64", "device architecture")
	analyzeCmd.Flags().Bool("no-progress-bar", false, "disable download progress bar")
}

type SnapAnalyzeResult struct {
	SnapName string   `json:"snapName"`
	Version  string   `json:"version,omitempty"`
	Revision int      `json:"revision,omitempty"`
	Arch     string   `json:"arch,omitempty"`
	SnapType string   `json:"snapType"`
	Findings []string `json:"findings,omitempty"`
}

func runAnalyzeCmd(cmd *cobra.Command, args []string) error {
	// Everything happens in a self-cleaning temp workspace.
	workDir, err := os.MkdirTemp("", "m2cp-snap-analyze-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	var result SnapAnalyzeResult
	var snapPath string

	if len(args) == 2 {
		// Store mode: <snapName> <version> -> resolve + download.
		snapName, version := args[0], args[1]
		arch, _ := cmd.Flags().GetString("arch")
		if !tools.IsValidAppName(snapName) {
			return fmt.Errorf("invalid snap name '%s'", snapName)
		}
		if !tools.IsValidDeviceArchitecture(arch) {
			return fmt.Errorf("invalid architecture '%s'", arch)
		}

		revision, downloadURL, err := resolveSnapDownload(cmd, snapName, version, arch)
		if err != nil {
			return err
		}

		snapPath = filepath.Join(workDir, snapName+".snap")
		withProgress := !format.JsonOutputMode
		if noPB, _ := cmd.Flags().GetBool("no-progress-bar"); noPB {
			withProgress = false
		}
		if _, err := helper.DownloadSnap(downloadURL, snapPath, snapPath, withProgress); err != nil {
			return err
		}

		result = SnapAnalyzeResult{
			SnapName: snapName,
			Version:  version,
			Revision: revision,
			Arch:     strings.ToLower(arch),
		}
	} else {
		// Local mode: <snapFile>.
		snapPath = args[0]
		if info, err := os.Stat(snapPath); err != nil || info.IsDir() {
			return fmt.Errorf("snap file not found: %s", snapPath)
		}
	}

	squashDir := filepath.Join(workDir, "squashfs-root")
	if err := runUnsquashfs(snapPath, squashDir); err != nil {
		return err
	}

	meta, err := readSnapMeta(filepath.Join(squashDir, "meta", "snap.yaml"))
	if err != nil {
		return err
	}
	result.SnapType = meta.Type
	// For a local file the store metadata is unknown; take it from snap.yaml.
	if len(args) == 1 {
		result.SnapName = meta.Name
		result.Version = meta.Version
		result.Arch = strings.Join(meta.Architectures, ",")
	}

	// Extract the trusted identity fields from the relevant binary.
	switch meta.Type {
	case "kernel":
		findings, err := analyzeKernel(workDir, squashDir)
		if err != nil {
			return err
		}
		result.Findings = findings
	case "snapd":
		findings, err := analyzeSnapd(squashDir)
		if err != nil {
			return err
		}
		result.Findings = findings
	}

	return format.PrintFormattedOutput(cmd, result, customAnalyzeFormatter)
}

// resolveSnapDownload resolves a snap (a subtype of app) by name + version to
// its revision number and download URL, assuming the given architecture.
func resolveSnapDownload(cmd *cobra.Command, snapName, version, arch string) (int, string, error) {
	apps, err := backend.GetAppIdsByAppNameAndArchitecture(cmd.Context(), snapName, backend.Architecture(strings.ToUpper(arch)))
	if err != nil {
		return 0, "", err
	}
	switch len(apps.Apps.Items) {
	case 0:
		return 0, "", fmt.Errorf("no snap found with name '%s' and architecture '%s'", snapName, arch)
	case 1:
		// ok
	default:
		return 0, "", fmt.Errorf("multiple snaps found with name '%s' and architecture '%s'", snapName, arch)
	}
	appID := apps.Apps.Items[0].Id

	// Auto-detect: parsable as int -> revision number, otherwise a version string.
	revisionInt, err := strconv.Atoi(version)
	if err != nil {
		revs, err := backend.GetAppRevisionsByAppId(cmd.Context(), appID, 1000, 0)
		if err != nil {
			return 0, "", err
		}
		found := false
		for _, item := range revs.AppRevisions.Items {
			if item.Version == version {
				revisionInt = item.Revision
				found = true
				break
			}
		}
		if !found {
			return 0, "", fmt.Errorf("version '%s' not found for snap '%s' (%s)", version, snapName, arch)
		}
	}

	revision, err := backend.GetAppRevision(cmd.Context(), appID, revisionInt)
	if err != nil {
		return 0, "", err
	}
	urlData, err := backend.GetAppDownloadUrl(cmd.Context(), revision.Id)
	if err != nil {
		return 0, "", err
	}
	return revision.Revision, urlData.DownloadAppRevision.DownloadUrl, nil
}

type snapMeta struct {
	Name          string   `yaml:"name"`
	Version       string   `yaml:"version"`
	Type          string   `yaml:"type"`
	Architectures []string `yaml:"architectures"`
}

func readSnapMeta(snapYamlPath string) (snapMeta, error) {
	data, err := os.ReadFile(snapYamlPath)
	if err != nil {
		return snapMeta{}, fmt.Errorf("reading %s: %w", snapYamlPath, err)
	}
	var meta snapMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return snapMeta{}, fmt.Errorf("parsing snap.yaml: %w", err)
	}
	if meta.Type == "" {
		meta.Type = "app" // snapd's default when type is unset
	}
	return meta, nil
}

// analyzeKernel extracts the kernel snap's initrd and returns the identity
// field line(s) baked into the initrd's snap-bootstrap binary.
func analyzeKernel(workDir, squashDir string) ([]string, error) {
	initrd := filepath.Join(squashDir, "initrd.img")
	if _, err := os.Stat(initrd); err != nil {
		return nil, fmt.Errorf("kernel snap has no initrd.img at %s: %w", initrd, err)
	}

	initrdDir := filepath.Join(workDir, "initrd")
	if err := runUnmkinitramfs(initrd, initrdDir); err != nil {
		return nil, err
	}

	bootstrap, err := findFile(initrdDir, "snap-bootstrap")
	if err != nil {
		return nil, fmt.Errorf("snap-bootstrap not found in extracted initrd: %w", err)
	}

	return stringsGrep(bootstrap, snapBootstrapFields)
}

// analyzeSnapd returns the identity field line(s) baked into the snapd snap's
// main snapd binary (usr/lib/snapd/snapd).
func analyzeSnapd(squashDir string) ([]string, error) {
	bin := filepath.Join(squashDir, "usr", "lib", "snapd", "snapd")
	if _, err := os.Stat(bin); err != nil {
		// Fall back to locating the binary anywhere in the snap.
		found, ferr := findFile(squashDir, "snapd")
		if ferr != nil {
			return nil, fmt.Errorf("snapd binary not found in snap: %w", ferr)
		}
		bin = found
	}
	return stringsGrep(bin, snapBootstrapFields)
}

// findFile returns the path of the first file named `name` under root.
func findFile(root, name string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == name {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("no file named %q under %s", name, root)
	}
	return found, nil
}

// --- external tool wrappers (fail helpfully when a tool is missing) ---

func requireTool(tool, pkgHint string) error {
	if _, err := exec.LookPath(tool); err != nil {
		return fmt.Errorf("required tool %q not found in PATH — install it first (e.g. `sudo apt install %s`)", tool, pkgHint)
	}
	return nil
}

func runUnsquashfs(snapPath, destDir string) error {
	if err := requireTool("unsquashfs", "squashfs-tools"); err != nil {
		return err
	}
	// -no-xattrs: we only read file contents, and snaps like snapd ship
	// files with security.capability xattrs that unsquashfs cannot restore
	// without root (otherwise it exits non-zero).
	out, err := exec.Command("unsquashfs", "-n", "-no-xattrs", "-f", "-d", destDir, snapPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("unsquashfs failed: %w\n%s", err, string(out))
	}
	return nil
}

// runUnmkinitramfs unpacks an initrd image. unmkinitramfs (a C binary these
// days) ships in initramfs-tools-core and itself relies on cpio.
func runUnmkinitramfs(initrdImg, destDir string) error {
	if err := requireTool("unmkinitramfs", "initramfs-tools-core"); err != nil {
		return err
	}
	if err := requireTool("cpio", "cpio"); err != nil {
		return err
	}
	out, err := exec.Command("unmkinitramfs", initrdImg, destDir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmkinitramfs failed: %w\n%s", err, string(out))
	}
	return nil
}

// stringsGrep runs `strings` over path and returns the de-duplicated lines that
// contain any of the given needles, preserving first-seen order.
func stringsGrep(path string, needles []string) ([]string, error) {
	if err := requireTool("strings", "binutils"); err != nil {
		return nil, err
	}
	out, err := exec.Command("strings", path).Output()
	if err != nil {
		return nil, fmt.Errorf("strings %s failed: %w", path, err)
	}
	seen := map[string]bool{}
	var matches []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		for _, needle := range needles {
			if strings.Contains(line, needle) {
				if !seen[line] {
					seen[line] = true
					matches = append(matches, line)
				}
				break
			}
		}
	}
	sort.Strings(matches)
	return matches, nil
}

func customAnalyzeFormatter(res SnapAnalyzeResult) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Snap:     %s\n", res.SnapName)
	if res.Version != "" {
		if res.Revision > 0 {
			fmt.Fprintf(&b, "Version:  %s (revision %d)\n", res.Version, res.Revision)
		} else {
			fmt.Fprintf(&b, "Version:  %s\n", res.Version)
		}
	}
	if res.Arch != "" {
		fmt.Fprintf(&b, "Arch:     %s\n", res.Arch)
	}
	fmt.Fprintf(&b, "Type:     %s\n", res.SnapType)
	if res.SnapType != "kernel" && res.SnapType != "snapd" {
		fmt.Fprintf(&b, "\nNo binary identity fields to extract for type %q — stopping.\n", res.SnapType)
		return b.String(), nil
	}
	b.WriteString("\nidentity fields:\n")
	if len(res.Findings) == 0 {
		b.WriteString("  (none found)\n")
	} else {
		for _, line := range res.Findings {
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}
	return b.String(), nil
}
