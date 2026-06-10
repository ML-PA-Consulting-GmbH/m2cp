package env

import (
	"fmt"
	"regexp"
	"strconv"
)

type SemanticVersion struct {
	Major         uint64 `json:"major" yaml:"major"`
	Minor         uint64 `json:"minor" yaml:"minor"`
	Patch         uint64 `json:"patch" yaml:"patch"`
	PreRelease    string `json:"pre-release" yaml:"pre-release"`
	BuildMetadata string `json:"build-metadata" yaml:"build-metadata"`
}

func (s *SemanticVersion) String() string {
	str := fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
	if s.PreRelease != "" {
		str += fmt.Sprintf("-%s", s.PreRelease)
	}
	if s.BuildMetadata != "" {
		str += fmt.Sprintf("+%s", s.BuildMetadata)
	}
	return str
}

func NewSemanticVersion(value string) (*SemanticVersion, error) {
	// Handle "unknown" version (when not set at build time)
	if value == "unknown" {
		return &SemanticVersion{
			Major:         0,
			Minor:         0,
			Patch:         0,
			PreRelease:    "unknown",
			BuildMetadata: "",
		}, nil
	}

	// source: https://regex101.com/r/Ly7O1x/196
	semVerExpr := "(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"
	matcher := regexp.MustCompile(semVerExpr)
	if !matcher.MatchString(value) {
		return nil, fmt.Errorf("could not match regular expression")
	}

	match := matcher.FindStringSubmatch(value)
	major, err := strconv.ParseUint(match[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("could not parse major version")
	}
	minor, err := strconv.ParseUint(match[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("could not parse minor version")
	}
	patch, err := strconv.ParseUint(match[3], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("could not parse patch version")
	}
	s := SemanticVersion{
		major,
		minor,
		patch,
		match[4],
		match[5],
	}
	return &s, nil
}

func (s *SemanticVersion) IsValid() bool {
	str := s.String()
	_, err := NewSemanticVersion(str)
	if err != nil {
		return false
	}
	return true
}
