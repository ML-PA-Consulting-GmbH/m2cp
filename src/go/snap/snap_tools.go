//go:build !windows

package snap

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)
import "gopkg.in/yaml.v2"

func UnmarshalDeclaration(raw []byte) (*Yaml, error) {
	var parsed Yaml
	if err := yaml.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	// If a base field is not specified, the snap will use the default base snap, which is currently core20.
	if parsed.Base == "" {
		parsed.Base = "core20"
	}
	if parsed.Type == "" {
		parsed.Type = "app"
	}
	return &parsed, nil
}

func MarshalDeclaration(declaration *Yaml) ([]byte, error) {
	return yaml.Marshal(declaration)
}

func VersionInc(version string) (string, error) {
	var err error
	pattern := regexp.MustCompile(`[0-9]+(\.[0-9]+)*(-[a-zA-Z0-9]+){0,1}`)
	if !pattern.MatchString(version) {
		return "", fmt.Errorf("illegal version number format. must match [0-9]+(\\.[0-9]+)*(-[a-zA-Z0-9]+){0,1}")
	}
	parts := strings.Split(version, "-")
	nums := strings.Split(parts[0], ".")
	for k, v := range nums {
		var n int
		n, err = strconv.Atoi(v)
		if err != nil {
			return "", fmt.Errorf("illegal version number format. must match [0-9]+(\\.[0-9]+)*(-[a-zA-Z0-9]+){0,1}")
		}
		if k == len(nums)-1 {
			n = n + 1
		}
		nums[k] = fmt.Sprintf("%d", n)
	}
	parts[0] = strings.Join(nums, ".")
	return strings.Join(parts, "-"), nil
}

type Yaml struct {
	Name            string                 `yaml:"name"`
	Version         string                 `yaml:"version"`
	Type            string                 `yaml:"type"`
	Architectures   []string               `yaml:"architectures,omitempty"`
	Assumes         []string               `yaml:"assumes,omitempty"`
	Title           string                 `yaml:"title,omitempty"`
	Description     string                 `yaml:"description"`
	Summary         string                 `yaml:"summary"`
	Provenance      string                 `yaml:"provenance,omitempty"`
	License         string                 `yaml:"license,omitempty"`
	Epoch           string                 `yaml:"epoch,omitempty"`
	Base            string                 `yaml:"base,omitempty"`
	Confinement     string                 `yaml:"confinement,omitempty"`
	Environment     OrderedMap             `yaml:"environment,omitempty"`
	Plugs           map[string]*PlugJSON   `yaml:"plugs,omitempty"`
	Slots           map[string]*SlotJSON   `yaml:"slots,omitempty"`
	Apps            map[string]AppYaml     `yaml:"apps,omitempty"`
	Hooks           map[string]HookYaml    `yaml:"hooks,omitempty"`
	Layout          map[string]LayoutYaml  `yaml:"layout,omitempty"`
	SystemUsernames map[string]interface{} `yaml:"system-usernames,omitempty"`
	Links           map[string][]string    `yaml:"links,omitempty"`
	// TypoLayouts is used to detect the use of the incorrect plural form of "layout"
	TypoLayouts TypoDetector `yaml:"layouts,omitempty"`
}

// PlugJSON aids in marshaling snap.PlugInfo into JSON.
type PlugJSON struct {
	Snap                string                 `json:"snap,omitempty"`
	Name                string                 `json:"plug,omitempty"`
	Interface           string                 `json:"interface,omitempty"`
	Attrs               map[string]interface{} `json:"attrs,omitempty"`
	Apps                []string               `json:"apps,omitempty"`
	Label               string                 `json:"label,omitempty"`
	AllowAutoConnection string                 `json:"allow-auto-connection,omitempty"`
	AllowInstall        string                 `json:"allow-installation,omitempty"`
	// Connections are synthesized, they are not on the original type.
	Connections []SlotRef `json:"connections,omitempty"`
}

type SlotRef struct {
	Snap string `json:"snap"`
	Name string `json:"slot"`
}

// SlotJSON aids in marshaling snap.SlotInfo into JSON.
type SlotJSON struct {
	Snap            string                 `json:"snap,omitempty"`
	Name            string                 `json:"slot,omitempty"`
	Interface       string                 `json:"interface,omitempty"`
	Attrs           map[string]interface{} `json:"attrs,omitempty"`
	Apps            []string               `json:"apps,omitempty"`
	Label           string                 `json:"label,omitempty"`
	AllowConnection string                 `json:"allow-connection,omitempty"`
	//AllowInstall        string                 `json:"allow-install,omitempty"` // used to be this, changed because of problems with lxd
	AllowInstall string `json:"allow-installation,omitempty"`
	// Connections are synthesized, they are not on the original type.
	Connections []PlugRef `json:"connections,omitempty"`
}

type PlugRef struct {
	Snap string `json:"snap"`
	Name string `json:"plug"`
}

type AppYaml struct {
	Aliases []string `yaml:"aliases,omitempty"`

	Command      string   `yaml:"command"`
	CommandChain []string `yaml:"command-chain,omitempty"`

	Daemon      string `yaml:"daemon,omitempty"`
	DaemonScope string `yaml:"daemon-scope,omitempty"`

	StopCommand     string `yaml:"stop-command,omitempty"`
	ReloadCommand   string `yaml:"reload-command,omitempty"`
	PostStopCommand string `yaml:"post-stop-command,omitempty"`
	StopTimeout     string `yaml:"stop-timeout,omitempty"`
	StartTimeout    string `yaml:"start-timeout,omitempty"`
	WatchdogTimeout string `yaml:"watchdog-timeout,omitempty"`
	Completer       string `yaml:"completer,omitempty"`
	RefreshMode     string `yaml:"refresh-mode,omitempty"`
	StopMode        string `yaml:"stop-mode,omitempty"`
	InstallMode     string `yaml:"install-mode,omitempty"`

	RestartCond  string   `yaml:"restart-condition,omitempty"`
	RestartDelay string   `yaml:"restart-delay,omitempty"`
	SlotNames    []string `yaml:"slots"`
	PlugNames    []string `yaml:"plugs"`

	BusName     string   `yaml:"bus-name,omitempty"`
	ActivatesOn []string `yaml:"activates-on,omitempty"`
	CommonID    string   `yaml:"common-id,omitempty"`

	Environment OrderedMap `yaml:"environment,omitempty"`

	Sockets map[string]SocketsYaml `yaml:"sockets,omitempty"`

	After  []string `yaml:"after,omitempty"`
	Before []string `yaml:"before,omitempty"`

	Timer string `yaml:"timer,omitempty"`

	Autostart string `yaml:"autostart,omitempty"`
}

type HookYaml struct {
	PlugNames    []string   `yaml:"plugs,omitempty"`
	SlotNames    []string   `yaml:"slots,omitempty"`
	Environment  OrderedMap `yaml:"environment,omitempty"`
	CommandChain []string   `yaml:"command-chain,omitempty"`
}

type LayoutYaml struct {
	Bind     string `yaml:"bind,omitempty"`
	BindFile string `yaml:"bind-file,omitempty"`
	Type     string `yaml:"type,omitempty"`
	User     string `yaml:"user,omitempty"`
	Group    string `yaml:"group,omitempty"`
	Mode     string `yaml:"mode,omitempty"`
	Symlink  string `yaml:"symlink,omitempty"`
}

type SocketsYaml struct {
	ListenStream string      `yaml:"listen-stream,omitempty"`
	SocketMode   os.FileMode `yaml:"socket-mode,omitempty"`
}

type TypoDetector struct {
	Hint string
}

// OrderedMap is a map of strings to strings that preserves the
// insert order when calling "Keys()".
//
// Heavily based on the spread.Environment code (thanks for that!)
type OrderedMap struct {
	keys []string
	vals map[string]string
}

// NewOrderedMap creates a new ordered map initialized with the
// given pairs of strings.
func NewOrderedMap(pairs ...string) *OrderedMap {
	o := &OrderedMap{vals: make(map[string]string),
		keys: make([]string, len(pairs)/2),
	}
	for i := 0; i+1 < len(pairs); i += 2 {
		o.vals[pairs[i]] = pairs[i+1]
		o.keys[i/2] = pairs[i]
	}
	return o
}

// Keys returns a list of keys in the map sorted by insertion order
func (o *OrderedMap) Keys() []string {
	return append([]string(nil), o.keys...)
}

// Get returns the value for the given key
func (o *OrderedMap) Get(k string) string {
	return o.vals[k]
}

// Del removes the given key from the data structure
func (o *OrderedMap) Del(key string) {
	l := len(o.vals)
	delete(o.vals, key)
	if len(o.vals) != l {
		for i, k := range o.keys {
			if k == key {
				copy(o.keys[i:], o.keys[i+1:])
				o.keys = o.keys[:len(o.keys)-1]
			}
		}
	}
}

// Set adds the given key, value to the map. If the key already
// exists it is removed and the new value is put on the end.
func (o *OrderedMap) Set(k, v string) {
	o.Del(k)
	o.keys = append(o.keys, k)
	o.vals[k] = v
}

// Copy makes a copy of the map
func (o *OrderedMap) Copy() *OrderedMap {
	copy := &OrderedMap{}
	copy.keys = append([]string(nil), o.keys...)
	copy.vals = make(map[string]string)
	for k, v := range o.vals {
		copy.vals[k] = v
	}
	return copy
}

// UnmarshalYAML unmarshals a yaml string map and preserves the order
func (o *OrderedMap) UnmarshalYAML(u func(any) error) error {
	var vals map[string]string
	if err := u(&vals); err != nil {
		return err
	}

	var seen = make(map[string]bool)
	var keys = make([]string, len(vals))
	var order yaml.MapSlice
	if err := u(&order); err != nil {
		return err
	}
	for i, item := range order {
		k, ok := item.Key.(string)
		_, good := vals[k]
		if !ok || !good {
			return fmt.Errorf("cannot read %q", item.Key)
		}
		if seen[k] {
			return fmt.Errorf("found duplicate key %q", k)
		}
		seen[k] = true
		keys[i] = k
	}
	o.keys = keys
	o.vals = vals
	return nil
}
