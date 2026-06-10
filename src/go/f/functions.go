package f

import (
	"sort"

	"gopkg.in/yaml.v3"
)

func First[T any, U any](first T, _ U) T {
	return first
}

func Ptr[T any](v T) *T {
	var zero T
	if any(v) == any(zero) {
		return nil
	}
	return &v
}

func SliceContainsAny[T comparable](haystack []T, needles []T) bool {
	if len(haystack) == 0 || len(needles) == 0 {
		return false
	}
	for _, n := range needles {
		for _, h := range haystack {
			if h == n {
				return true
			}
		}
	}
	return false
}

// Unique returns a sorted unique copy of the slice for comparable types.
func Unique[T comparable](in []T) []T {
	if len(in) == 0 {
		return nil
	}
	m := make(map[T]struct{}, len(in))
	for _, v := range in {
		m[v] = struct{}{}
	}
	out := make([]T, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	// Best-effort stable-ish order: convert to strings if T is string; otherwise leave unsorted
	switch any(out).(type) {
	case []string:
		ss := any(out).([]string)
		sort.Strings(ss)
		return any(ss).([]T)
	}
	return out
}

func ToMap[T any](in T) (map[string]any, error) {
	// Marshal the input to YAML, then unmarshal into a generic map.
	// This preserves yaml tags and nested structures without manual reflection.
	b, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func ToMaybeMap[T any](in T) map[string]any {
	m, err := ToMap(in)
	if err != nil {
		return nil
	}
	return m
}

// MergeMaps recursively merges multiple maps, with later maps taking precedence over earlier ones
// Returns empty map if no maps provided
func MergeMaps(maps ...map[string]any) map[string]any {
	if len(maps) == 0 {
		return make(map[string]any)
	}

	result := make(map[string]any)

	// Merge each map in order
	for _, m := range maps {
		if m == nil {
			continue
		}

		for k, v := range m {
			// Skip nil overlay values so they don't clobber existing data
			if v == nil {
				if _, exists := result[k]; exists {
					continue
				}
			}
			if baseVal, exists := result[k]; exists {
				// If both values are maps, merge them recursively
				if baseMap, baseIsMap := baseVal.(map[string]any); baseIsMap {
					if overlayMap, overlayIsMap := v.(map[string]any); overlayIsMap {
						result[k] = MergeMaps(baseMap, overlayMap)
						continue
					}
				}
				// If both values are arrays, concatenate them
				if baseArray, baseIsArray := baseVal.([]any); baseIsArray {
					if overlayArray, overlayIsArray := v.([]any); overlayIsArray {
						result[k] = append(baseArray, overlayArray...)
						continue
					}
				}
			}
			// Otherwise, new value takes precedence
			result[k] = v
		}
	}

	return result
}
