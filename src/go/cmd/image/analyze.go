package image

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"m2cpcli/format"
	"m2cpcli/tools"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/cheggaaa/pb/v3"
	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/partition/gpt"
	"github.com/spf13/cobra"
	"github.com/ulikunitz/xz"
	"gopkg.in/yaml.v2"
)

// analyzeCmd reads the per-system build.yaml out of a Ubuntu-Core-based
// edge-OS image. The image may be a raw .img or xz-compressed .img.xz;
// both are decoded without root, without mounting, and without
// external tools (pure Go via go-diskfs + xz). Output honours the
// project-wide --json flag.
var analyzeCmd = &cobra.Command{
	Use:   "analyze <image.img(.xz)>",
	Short: "Print build info from an edge-OS image",
	Long: `Open an edge-OS image (raw or xz-compressed) and extract the
contents of /var/lib/snapd/seed/systems/*/build.yaml -- the per-build
identifier file that records appstore URL, project, board, version, etc.

Works on Ubuntu Core based lines (UC20+ seed layout). Other lines
will be added as separate probes.`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	ImageCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().Bool("keep-uncompressed", false,
		"do not delete the decompressed .img after analysis; "+
			"re-run the tool with that path to skip the xz step")
	analyzeCmd.Flags().Bool("assertions", false,
		"report the seed's account/account-key assertions (and any snap "+
			"publisher referenced without an account assertion) instead of build info")
}

// BuildInfo is whatever build.yaml contained -- the structured
// payload printed by `m2cp image analyze`. A typed mirror of the
// known fields would be lossier (extra keys vanish) and a maintenance
// drag every time the build pipeline grows a column. Image-level
// metadata (path, partition name, system label) is logged via the
// stderr progress lines rather than mixed into this payload.
type BuildInfo map[string]any

// suppressBuildYAMLKeys are legacy build.yaml keys that are dropped
// from the output. They duplicate or misrepresent information that
// the model assertion / appstore-url provide authoritatively:
//
//   - "board"    : older builders used this for the model name. Now
//                  superseded by "model-name" from the assertion.
//   - "project"  : human-readable bucket that doesn't survive to
//                  runtime and can disagree with the brand-id; not
//                  useful for identifying an image.
//   - "appstore" : just the operator-chosen nickname for the store.
//                  Can be set wrongly; "appstore-url" is the
//                  authoritative pointer.
//
// Anything else in build.yaml passes through unchanged.
var suppressBuildYAMLKeys = []string{"board", "project", "appstore"}

func runAnalyze(cmd *cobra.Command, args []string) error {
	keep, _ := cmd.Flags().GetBool("keep-uncompressed")
	imgPath, cleanup, err := materialize(args[0], keep)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	progress("=> opening %s", imgPath)
	d, err := diskfs.Open(imgPath)
	if err != nil {
		return fmt.Errorf("opening image: %w", err)
	}

	progress("=> reading partition table")
	table, err := d.GetPartitionTable()
	if err != nil {
		return fmt.Errorf("reading partition table: %w", err)
	}
	gptTable, ok := table.(*gpt.Table)
	if !ok {
		return errors.New("image is not GPT-partitioned")
	}
	for i, p := range gptTable.Partitions {
		if p == nil || p.Size == 0 {
			continue
		}
		progress("   partition %d: name=%q type=%s size=%s",
			i+1, p.Name, p.Type, tools.FormatBytes(int64(p.Size)))
	}

	progress("=> probing partitions for /systems/")
	fs, partLabel, err := findSeedFilesystem(d, gptTable)
	if err != nil {
		return err
	}
	progress("   found /systems/ on partition %q", partLabel)

	entries, err := fs.ReadDir("systems")
	if err != nil {
		return fmt.Errorf("reading /systems on partition %q: %w", partLabel, err)
	}

	found := map[string]BuildInfo{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "." || name == ".." {
			continue
		}
		info := BuildInfo{}

		// build.yaml is optional (only present on builds that emit one).
		buildPath := path.Join("systems", name, "build.yaml")
		progress("=> reading %s", buildPath)
		if body, err := readFile(fs, buildPath); err == nil {
			progress("   %s (%d bytes)", buildPath, len(body))
			if err := yaml.Unmarshal(body, &info); err != nil {
				return fmt.Errorf("parsing /systems/%s/build.yaml: %w", name, err)
			}
			progress("   parsed %d keys", len(info))
			for _, k := range suppressBuildYAMLKeys {
				if _, ok := info[k]; ok {
					progress("   suppressing legacy key %q", k)
					delete(info, k)
				}
			}
		} else {
			progress("   not present: %v", err)
		}

		// The model assertion is always present in a UC seed; the
		// cryptographically-signed model name + revision are the
		// canonical answer to "what is this image", so they take
		// precedence over anything build.yaml claimed.
		modelPath := path.Join("systems", name, "model")
		progress("=> reading %s", modelPath)
		if body, err := readFile(fs, modelPath); err == nil {
			progress("   %s (%d bytes)", modelPath, len(body))
			h, err := parseAssertionHeaders(body)
			if err != nil {
				return fmt.Errorf("parsing /systems/%s/model: %w", name, err)
			}
			if v, ok := h["model"]; ok {
				info["model-name"] = v
			}
			if v, ok := h["revision"]; ok {
				info["model-revision"] = v
			}
			if v, ok := h["brand-id"]; ok {
				info["model-brand-id"] = v
			}
			progress("   model=%s revision=%s brand-id=%s",
				h["model"], h["revision"], h["brand-id"])
		} else {
			progress("   not present: %v", err)
		}

		if len(info) == 0 {
			progress("   /systems/%s has neither build.yaml nor model; skipping", name)
			continue
		}
		found[name] = info
	}
	switch len(found) {
	case 0:
		return fmt.Errorf("no build.yaml or model found in any /systems/* directory of partition %q", partLabel)
	case 1:
		// Pull the single entry out so the result is the build info
		// itself, not wrapped in a one-element map.
	default:
		labels := make([]string, 0, len(found))
		for l := range found {
			labels = append(labels, l)
		}
		sort.Strings(labels)
		return fmt.Errorf("image has %d systems (%s); flat output assumes exactly one",
			len(found), strings.Join(labels, ", "))
	}

	var label string
	var only BuildInfo
	for l, info := range found {
		label, only = l, info
	}

	if assertions, _ := cmd.Flags().GetBool("assertions"); assertions {
		report, err := readSeedAssertions(fs, label)
		if err != nil {
			return err
		}
		return format.PrintFormattedOutput(cmd, report, formatAssertionReport)
	}

	return format.PrintFormattedOutput(cmd, only, formatBuildInfo)
}

// parseAssertionHeaders reads the header section of a snap-asserts
// text assertion (everything up to the first blank line) and returns
// the top-level "key: value" pairs. Indented continuation lines
// (e.g. inside "snaps:" blocks) and headerless lines are skipped.
// This is intentionally a tiny text scanner rather than a snapd
// dependency: we only need primary headers (model, revision,
// brand-id), and the assertion file format is stable.
func parseAssertionHeaders(body []byte) (map[string]string, error) {
	out := map[string]string{}
	s := bufio.NewScanner(bytes.NewReader(body))
	s.Buffer(make([]byte, 1024*64), 1024*1024)
	for s.Scan() {
		line := s.Text()
		if line == "" {
			break // end of headers
		}
		if line[0] == ' ' || line[0] == '\t' {
			continue // continuation / nested block content
		}
		colon := strings.IndexByte(line, ':')
		if colon < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])
		if val == "" {
			continue // nested-block opener like "snaps:"
		}
		out[key] = val
	}
	return out, s.Err()
}

// progress writes a status line to stderr unless --json is in effect.
// Image inspection touches multi-GB blobs; silence would look frozen.
func progress(f string, args ...any) {
	if format.IsJsonMode() {
		return
	}
	fmt.Fprintf(os.Stderr, f+"\n", args...)
}

// formatBuildInfo is the human-readable formatter. Emits the
// build.yaml as a flat key:value block in a stable order
// (well-known keys first, then alphabetical for anything
// unrecognised).
func formatBuildInfo(info BuildInfo) (string, error) {
	knownOrder := []string{
		"builder", "builder-version",
		"model-name", "model-revision", "model-brand-id",
		"appstore-url",
		"arch", "version", "date", "artifact", "grade",
		"context", "repo", "commit",
		"manifest-commit", "manifest-version",
	}

	var b strings.Builder
	seen := map[string]bool{}
	for _, k := range knownOrder {
		if v, ok := info[k]; ok {
			fmt.Fprintf(&b, "%s: %v\n", k, v)
			seen[k] = true
		}
	}
	extra := make([]string, 0, len(info))
	for k := range info {
		if !seen[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	for _, k := range extra {
		fmt.Fprintf(&b, "%s: %v\n", k, info[k])
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

// materialize returns a path to a raw image file usable by go-diskfs.
// If the input is xz-compressed it is streamed into a temp file with
// a progress bar driven off the compressed input size. When `keep`
// is false the returned cleanup removes the temp file; when true the
// path is preserved and printed so the user can re-run against it
// directly (skipping the multi-minute xz step).
func materialize(input string, keep bool) (string, func(), error) {
	if !strings.HasSuffix(strings.ToLower(input), ".xz") {
		progress("=> input is raw .img: %s", input)
		return input, nil, nil
	}

	f, err := os.Open(input)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return "", nil, err
	}

	tmp, err := os.CreateTemp("", "m2cp-image-*.img")
	if err != nil {
		return "", nil, err
	}
	progress("=> decompressing %s (%s) -> %s",
		input, tools.FormatBytes(fi.Size()), tmp.Name())

	var reader io.Reader = f
	var bar *pb.ProgressBar
	if !format.IsJsonMode() {
		// Progress is driven off compressed bytes consumed -- the
		// uncompressed total isn't known up front for xz. Output
		// goes to stderr to keep stdout clean for the structured
		// payload later.
		bar = pb.Full.Start64(fi.Size())
		bar.Set(pb.Bytes, true)
		bar.SetWriter(os.Stderr)
		reader = bar.NewProxyReader(f)
	}

	xzReader, err := xz.NewReader(reader)
	if err != nil {
		if bar != nil {
			bar.Finish()
		}
		tmp.Close()
		os.Remove(tmp.Name())
		return "", nil, fmt.Errorf("xz reader: %w", err)
	}

	n, err := io.Copy(tmp, xzReader)
	if bar != nil {
		bar.Finish()
	}
	if err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", nil, fmt.Errorf("decompressing: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", nil, err
	}
	progress("   decompressed: %s", tools.FormatBytes(n))

	if keep {
		progress("   kept (re-run with this path to skip xz): %s", tmp.Name())
		return tmp.Name(), nil, nil
	}
	return tmp.Name(), func() { os.Remove(tmp.Name()) }, nil
}

// findSeedFilesystem walks the GPT partitions and returns the first
// filesystem that contains a /systems/ directory -- the Ubuntu Core
// seed convention. Per-partition probe outcomes are logged so a
// "not found" failure is diagnosable.
func findSeedFilesystem(d *disk.Disk, gptTable *gpt.Table) (filesystem.FileSystem, string, error) {
	for i, p := range gptTable.Partitions {
		if p == nil || p.Size == 0 {
			continue
		}
		progress("   trying partition %d (%q)...", i+1, p.Name)
		fs, err := d.GetFilesystem(i + 1)
		if err != nil {
			progress("      not a readable filesystem: %v", err)
			continue
		}
		entries, err := fs.ReadDir("systems")
		if err != nil {
			progress("      no /systems/ here: %v", err)
			continue
		}
		if len(entries) == 0 {
			progress("      /systems/ is empty")
			continue
		}
		progress("      /systems/: %d entries", len(entries))
		return fs, p.Name, nil
	}
	return nil, "", errors.New("no FAT partition with a /systems/ directory found")
}

func readFile(fs filesystem.FileSystem, p string) ([]byte, error) {
	f, err := fs.OpenFile(p, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}

// --- seed assertion inventory (m2cp image analyze --assertions) ---

type AssertionAccount struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName,omitempty"`
}

type AssertionKey struct {
	AccountID string `json:"accountId"`
	Name      string `json:"name,omitempty"`
	Sha3_384  string `json:"sha3-384,omitempty"`
}

type AssertionReport struct {
	System           string             `json:"system"`
	Accounts         []AssertionAccount `json:"accounts"`
	AccountKeys      []AssertionKey     `json:"accountKeys"`
	SnapDeclarations int                `json:"snapDeclarations"`
	// MissingAccounts are account-ids referenced by seeded assertions (snap
	// publishers, key owners, signing authorities) that have no account
	// assertion in the seed -- the boot-blocking gap.
	MissingAccounts []string `json:"missingAccounts,omitempty"`
}

// readSeedAssertions parses every assertion under
// /systems/<system>/assertions/ and summarises the identity assertions,
// flagging referenced account-ids that lack an account assertion.
func readSeedAssertions(fs filesystem.FileSystem, system string) (AssertionReport, error) {
	dir := path.Join("systems", system, "assertions")
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return AssertionReport{}, fmt.Errorf("reading %s: %w", dir, err)
	}

	rep := AssertionReport{System: system}
	present := map[string]bool{}    // account-ids with an account assertion in the seed
	referenced := map[string]bool{} // account-ids referenced by any assertion

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		body, err := readFile(fs, path.Join(dir, e.Name()))
		if err != nil {
			return AssertionReport{}, fmt.Errorf("reading %s: %w", path.Join(dir, e.Name()), err)
		}
		for _, block := range splitAssertions(body) {
			h, err := parseAssertionHeaders(block)
			if err != nil {
				return AssertionReport{}, err
			}
			if a := h["authority-id"]; a != "" {
				referenced[a] = true
			}
			switch h["type"] {
			case "account":
				rep.Accounts = append(rep.Accounts, AssertionAccount{AccountID: h["account-id"], DisplayName: h["display-name"]})
				present[h["account-id"]] = true
			case "account-key":
				rep.AccountKeys = append(rep.AccountKeys, AssertionKey{AccountID: h["account-id"], Name: h["name"], Sha3_384: h["public-key-sha3-384"]})
				if id := h["account-id"]; id != "" {
					referenced[id] = true
				}
			case "snap-declaration":
				rep.SnapDeclarations++
				if pid := h["publisher-id"]; pid != "" {
					referenced[pid] = true
				}
			}
		}
	}

	// Exclude known trust anchors that live in the snapd system DB rather
	// than the seed: canonical (root of trust) and the model's brand-id.
	delete(referenced, "canonical")
	if body, err := readFile(fs, path.Join("systems", system, "model")); err == nil {
		if h, err := parseAssertionHeaders(body); err == nil {
			if brand := h["brand-id"]; brand != "" {
				delete(referenced, brand)
			}
		}
	}
	for id := range referenced {
		if !present[id] {
			rep.MissingAccounts = append(rep.MissingAccounts, id)
		}
	}
	sort.Strings(rep.MissingAccounts)
	sort.Slice(rep.Accounts, func(i, j int) bool { return rep.Accounts[i].AccountID < rep.Accounts[j].AccountID })
	sort.Slice(rep.AccountKeys, func(i, j int) bool {
		if rep.AccountKeys[i].AccountID != rep.AccountKeys[j].AccountID {
			return rep.AccountKeys[i].AccountID < rep.AccountKeys[j].AccountID
		}
		return rep.AccountKeys[i].Name < rep.AccountKeys[j].Name
	})
	return rep, nil
}

// splitAssertions splits a concatenated assertion stream into the
// header blocks of each assertion. The wire format separates an
// assertion's headers from its signature -- and one assertion from the
// next -- with a blank line, so splitting on "\n\n" yields alternating
// header/signature chunks; only the header chunks start with "type:".
func splitAssertions(body []byte) [][]byte {
	var out [][]byte
	for _, p := range bytes.Split(body, []byte("\n\n")) {
		if bytes.HasPrefix(p, []byte("type:")) {
			out = append(out, p)
		}
	}
	return out
}

func formatAssertionReport(r AssertionReport) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "system: %s\n\n", r.System)

	fmt.Fprintf(&b, "accounts in seed (%d):\n", len(r.Accounts))
	for _, a := range r.Accounts {
		fmt.Fprintf(&b, "  %-22s %s\n", a.AccountID, a.DisplayName)
	}

	fmt.Fprintf(&b, "\naccount-keys in seed (%d):\n", len(r.AccountKeys))
	for _, k := range r.AccountKeys {
		name := k.Name
		if name == "" {
			name = "-"
		}
		fmt.Fprintf(&b, "  %-22s %-22s %s\n", k.AccountID, name, k.Sha3_384)
	}

	fmt.Fprintf(&b, "\nsnap-declarations: %d\n", r.SnapDeclarations)

	if len(r.MissingAccounts) > 0 {
		b.WriteString("\nMISSING account assertions (referenced in seed, not present):\n")
		for _, id := range r.MissingAccounts {
			fmt.Fprintf(&b, "  %s\n", id)
		}
		b.WriteString("\nnote: brand trust anchors injected into the snapd system DB may show\n")
		b.WriteString("here, but a snap publisher missing its account assertion boot-loops with\n")
		b.WriteString("\"cannot resolve prerequisite assertion: account (<id>)\".\n")
	} else {
		b.WriteString("\nno missing account assertions detected.\n")
	}
	return strings.TrimRight(b.String(), "\n"), nil
}
