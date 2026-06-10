package model

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"m2cpcli/backend"
	gql "m2cpcli/graphql"

	"gopkg.in/yaml.v3"
)

// modelCompare captures exactly the fields that make up the model
// declaration we push. Everything else the appstore stores in the
// signed assertion (authority-id, brand-id, per-snap ids, revision,
// classic, timestamp, sign-key, signature) is either a placeholder
// on our side, server-assigned, or derivable -- so it can't matter
// for "is this the same model we'd push".
//
// Snaps are compared by (name, type): snap names are bijective with
// snap ids (and our pushed ids are PLACEHOLDER), but a snap's role in
// a model -- gadget, kernel, base, snapd, app -- is part of the model
// declaration. Two models that share names but disagree on a role
// are not the same model.
//
// The dual json/yaml tags let the same struct decode both our
// to-push JSON (encoding/json) and a stored assertion header (yaml).
type modelCompare struct {
	Architecture  string `json:"architecture" yaml:"architecture"`
	Base          string `json:"base" yaml:"base"`
	Grade         string `json:"grade" yaml:"grade"`
	StorageSafety string `json:"storage-safety" yaml:"storage-safety"`
	Model         string `json:"model" yaml:"model"`
	Snaps         []struct {
		Name string `json:"name" yaml:"name"`
		Type string `json:"type" yaml:"type"`
	} `json:"snaps" yaml:"snaps"`
}

// snapTypeMap returns a map of snap name -> declared type. Two models
// that share snap names but disagree on a snap's role (e.g. one says
// m2cp-coap is a base, the other says it's an app) are NOT the same
// model: snapd's seedwriter cross-checks the model's per-snap type
// against the downloaded snap and refuses on mismatch ("base \"x\"
// has unexpected type: app"). So the fingerprint compares (name,
// type) tuples, not names alone.
func (m modelCompare) snapTypeMap() map[string]string {
	s := make(map[string]string, len(m.Snaps))
	for _, sn := range m.Snaps {
		s[sn.Name] = sn.Type
	}
	return s
}

// parseToPushModel parses the model.json we're about to push.
func parseToPushModel(modelJson string) (modelCompare, error) {
	var mc modelCompare
	if err := json.Unmarshal([]byte(modelJson), &mc); err != nil {
		return mc, fmt.Errorf("parsing model json: %w", err)
	}
	if mc.Model == "" {
		return mc, fmt.Errorf("model json has no \"model\" field")
	}
	return mc, nil
}

// parseAssertionModel parses a stored model assertion's header. The
// body is "<header>\n\n<signature>"; only the header is YAML.
func parseAssertionModel(body string) (modelCompare, error) {
	header := strings.SplitN(body, "\n\n", 2)[0]
	var mc modelCompare
	if err := yaml.Unmarshal([]byte(header), &mc); err != nil {
		return mc, fmt.Errorf("parsing assertion header: %w", err)
	}
	return mc, nil
}

// diffModels compares the model we want to push against an existing
// revision (have). It returns whether they're identical in the
// fields we push, and otherwise a short delta expressed as the
// change FROM the existing revision TO our model:
//
//	-snap                   a snap in the revision we no longer declare
//	+snap                   a snap we declare that the revision lacks
//	snap:type:old->new      a snap that's in both but with a different role
//	field:old->new          a scalar field we changed
func diffModels(want, have modelCompare) (identical bool, summary string) {
	var parts []string

	for _, f := range []struct{ name, have, want string }{
		{"arch", have.Architecture, want.Architecture},
		{"base", have.Base, want.Base},
		{"grade", have.Grade, want.Grade},
		{"storage-safety", have.StorageSafety, want.StorageSafety},
	} {
		if f.have != f.want {
			parts = append(parts, fmt.Sprintf("%s:%s->%s", f.name, orNone(f.have), orNone(f.want)))
		}
	}

	wmap, hmap := want.snapTypeMap(), have.snapTypeMap()
	var removed, added, retyped []string
	for n, ht := range hmap {
		if _, ok := wmap[n]; !ok {
			removed = append(removed, n)
		} else if wt := wmap[n]; wt != ht {
			retyped = append(retyped, fmt.Sprintf("%s:type:%s->%s", n, orNone(ht), orNone(wt)))
		}
	}
	for n := range wmap {
		if _, ok := hmap[n]; !ok {
			added = append(added, n)
		}
	}
	sort.Strings(removed)
	sort.Strings(added)
	sort.Strings(retyped)
	parts = append(parts, retyped...)
	for _, n := range removed {
		parts = append(parts, "-"+n)
	}
	for _, n := range added {
		parts = append(parts, "+"+n)
	}

	if len(parts) == 0 {
		return true, ""
	}
	return false, strings.Join(parts, " ")
}

func orNone(s string) string {
	if s == "" {
		return "<none>"
	}
	return s
}

// revisionRef is a stored model revision we may compare against.
type revisionRef struct {
	id       string
	revision int
}

// findMatchingRevision enumerates every existing revision of the
// model (ascending) and narrates each one's delta against the model
// we're about to push. Every revision is examined and commented --
// even after a match is found -- so the operator sees the full
// picture (e.g. "rev 1 matches, but revs 12/14 carry two extra
// snaps"). It returns the earliest identical revision so the caller
// reuses it instead of pushing; found is false when none match.
//
// This is the CLI-side workaround for the appstore's order-sensitive
// duplicate check: the comparison is semantic (snap order ignored),
// so a re-push of an already-present model is recognised.
func findMatchingRevision(ctx context.Context, want modelCompare, out io.Writer) (rev int, found bool, err error) {
	refs, err := listRevisions(ctx, want.Model)
	if err != nil {
		return 0, false, err
	}
	if len(refs) == 0 {
		fmt.Fprintf(out, "  (no existing revisions of %q)\n", want.Model)
		return 0, false, nil
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].revision < refs[j].revision })

	for _, ref := range refs {
		m, err := gql.EdgeDeviceModelById(ctx, gql.UUID(ref.id),
			gql.EdgeDeviceModelQueryOptions{WithAssertion: true})
		if err != nil {
			return 0, false, fmt.Errorf("fetching revision %d assertion: %w", ref.revision, err)
		}
		if m.EdgeDeviceModelAssertion == nil {
			fmt.Fprintf(out, "  rev %d: (no assertion body; skipped)\n", ref.revision)
			continue
		}
		have, err := parseAssertionModel(m.EdgeDeviceModelAssertion.AssertionBody)
		if err != nil {
			return 0, false, fmt.Errorf("parsing revision %d: %w", ref.revision, err)
		}
		identical, summary := diffModels(want, have)
		switch {
		case identical && !found:
			fmt.Fprintf(out, "  rev %d: MATCH (identical model declaration) <- will reuse\n", ref.revision)
			rev, found = ref.revision, true
		case identical:
			fmt.Fprintf(out, "  rev %d: also matches (keeping earliest, rev %d)\n", ref.revision, rev)
		default:
			fmt.Fprintf(out, "  rev %d: differs (%s)\n", ref.revision, summary)
		}
	}
	return rev, found, nil
}

// listRevisions returns every stored revision of the named model.
func listRevisions(ctx context.Context, modelName string) ([]revisionRef, error) {
	take := 100
	skip := 0
	var refs []revisionRef
	for {
		result, err := backend.GetDeviceModelRevisionList(ctx, nil, &take, &skip)
		if err != nil {
			return nil, err
		}
		if result.DeviceModelRevisions == nil || len(result.DeviceModelRevisions.Items) == 0 {
			break
		}
		for _, item := range result.DeviceModelRevisions.Items {
			if item.DeviceModel.ModelName != modelName {
				continue
			}
			rev := 0
			if item.Revision != nil {
				rev = *item.Revision
			}
			refs = append(refs, revisionRef{id: item.Id, revision: rev})
		}
		if !result.DeviceModelRevisions.PageInfo.HasNextPage {
			break
		}
		skip += take
	}
	return refs, nil
}
