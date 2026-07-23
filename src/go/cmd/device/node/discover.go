package node

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools/console"
)

var discoverCmd = &cobra.Command{
	Use:   "discover <deviceNameOrSerial> <app>",
	Short: "Discover the RPCs an app exposes on a device, with their parameters",
	Long: `Request the RPC documentation live from a device (via the read-only NodeDocumentation
RPC) and list the available RPCs of an app/node together with their parameters.

RPC documentation is per app/node, so an app is required. Use
'm2cp device node list <device>' to see which apps/nodes exist, or pass --all to
document every node on the device (in which case no app is given). The device must be
online.`,
	Args: validateDiscoverArgs,
	RunE: runDiscoverCmd,
}

func init() {
	nodeCmd.AddCommand(discoverCmd)
	discoverCmd.Flags().BoolP("all", "a", false, "document every app/node on the device (excludes sensor.* nodes)")
	discoverCmd.Flags().Bool("include-sensors", false, "with --all, also include sensor.* nodes")
}

// validateDiscoverArgs requires an explicit app unless --all is set (which documents
// every node and therefore takes no app).
func validateDiscoverArgs(cmd *cobra.Command, args []string) error {
	all, _ := cmd.Flags().GetBool("all")
	if all {
		if len(args) != 1 {
			return fmt.Errorf("with --all, pass only the device (no app)")
		}
		return nil
	}
	if len(args) != 2 {
		return fmt.Errorf("an app is required: 'device node discover <device> <app>' (or use --all); run 'device node list <device>' to see the apps")
	}
	return nil
}

type rpcParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

type rpcReturnField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	Description string `json:"description,omitempty"`
}

type discoveredRpc struct {
	Command            string           `json:"command"`
	Description        string           `json:"description,omitempty"`
	DescriptionReturns string           `json:"descriptionReturns,omitempty"`
	Parameters         []rpcParameter   `json:"parameters"`
	Returns            []rpcReturnField `json:"returns"`
}

type discoveredNode struct {
	Node string          `json:"node"`
	Rpcs []discoveredRpc `json:"rpcs"`
	// Note is set when the node exposes no RPCs or its documentation could not be
	// fetched, so the node still appears in the listing.
	Note string `json:"note,omitempty"`
}

type discoverOutput struct {
	Nodes []discoveredNode `json:"nodes"`
}

// raw shapes of the device's NodeDocumentation payload (parameters/returns are maps keyed by name)
type rawDocParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    *bool  `json:"required"`
	Optional    *bool  `json:"optional"`
	Description string `json:"description"`
}

type rawDocReturn struct {
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	Description string `json:"description"`
}

type rawDocEntry struct {
	Name               string                  `json:"name"`
	Description        string                  `json:"description"`
	DescriptionReturns string                  `json:"descriptionReturns"`
	Parameters         map[string]rawDocParam  `json:"parameters"`
	Returns            map[string]rawDocReturn `json:"returns"`
}

func runDiscoverCmd(cmd *cobra.Command, args []string) error {
	deviceId, err := gql.DeviceIdByNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	device, err := gql.DeviceByDeviceId(cmd.Context(), deviceId)
	if err != nil {
		return err
	}
	serial := device.DeviceSerial

	all, _ := cmd.Flags().GetBool("all")
	includeSensors, _ := cmd.Flags().GetBool("include-sensors")

	var nodes []string
	if all {
		nodes, err = enumerateDeviceNodes(cmd.Context(), serial, includeSensors)
		if err != nil {
			return err
		}
		if len(nodes) == 0 {
			return fmt.Errorf("no nodes found on device")
		}
	} else {
		nodes = []string{args[1]}
	}

	out := discoverOutput{Nodes: make([]discoveredNode, 0, len(nodes))}
	for _, node := range nodes {
		prefix := nodeAddressPrefix(node, serial)
		rpcs, err := fetchNodeDocumentation(cmd.Context(), serial, prefix)
		if err != nil {
			// Always list the node, marked as unavailable, and warn with the reason
			// rather than aborting — so the requested app (or every node in --all)
			// always appears in the output.
			cmd.PrintErrln(fmt.Sprintf("warning: could not document node '%s': %s", prefix, err))
			out.Nodes = append(out.Nodes, discoveredNode{Node: prefix, Rpcs: []discoveredRpc{}, Note: "documentation unavailable"})
			continue
		}
		out.Nodes = append(out.Nodes, discoveredNode{Node: prefix, Rpcs: rpcs})
	}

	return format.PrintFormattedOutput(cmd, out, customDiscoverFormatter)
}

// nodeAddressPrefix turns a node identifier into its serial-free RPC address prefix.
// A bare app name (no dot) lives under the default rpc. router; a name already carrying
// a router (a dot) is used verbatim; a trailing ".<serial>" (as DeviceInfo may return)
// is stripped.
func nodeAddressPrefix(node, serial string) string {
	node = strings.TrimSuffix(node, "."+serial)
	if !strings.Contains(node, ".") {
		return "rpc." + node
	}
	return node
}

func enumerateDeviceNodes(ctx context.Context, serial string, includeSensors bool) ([]string, error) {
	result, err := gql.ExecuteRpc(ctx, &gql.ExecuteRpcInput{
		Command: "DeviceInfo",
		Address: fmt.Sprintf("rpc.m2cp-gateway.%s", serial),
	})
	if err != nil {
		return nil, fmt.Errorf("could not enumerate nodes: %w", err)
	}
	nodes, _ := nodesFromResult(result)
	filtered := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n == "" {
			continue
		}
		if !includeSensors && strings.HasPrefix(n, "sensor.") {
			continue
		}
		filtered = append(filtered, n)
	}
	return filtered, nil
}

func fetchNodeDocumentation(ctx context.Context, serial, prefix string) ([]discoveredRpc, error) {
	address := prefix + "." + serial
	result, err := gql.ExecuteRpc(ctx, &gql.ExecuteRpcInput{Command: "NodeDocumentation", Address: address})
	if err != nil {
		return nil, fmt.Errorf("could not execute NodeDocumentation on %s: %w", address, err)
	}
	if len(result.Responses) == 0 {
		return nil, fmt.Errorf("no response from %s — the device is likely offline or unreachable", address)
	}
	resp := result.Responses[0]
	if resp.Error != 0 {
		if resp.Message != "" {
			return nil, fmt.Errorf("NodeDocumentation on %s failed: %s", address, resp.Message)
		}
		return nil, fmt.Errorf("NodeDocumentation on %s failed with error %d — the device is likely offline", address, resp.Error)
	}
	value := resultValue(resp.Results, "documentation", "", "data")
	if value == "" {
		return []discoveredRpc{}, nil
	}
	return normalizeDocumentation(decodeDocumentationPayload(value)), nil
}

func resultValue(results []gql.ExecuteRpcCommandResult, keys ...string) string {
	for _, key := range keys {
		for _, r := range results {
			if r.Key == key {
				return r.Value
			}
		}
	}
	return ""
}

// decodeDocumentationPayload base64-decodes the NodeDocumentation value, falling back to
// treating it as raw JSON if it is not valid base64.
func decodeDocumentationPayload(value string) []byte {
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded
	}
	return []byte(value)
}

func normalizeDocumentation(payload []byte) []discoveredRpc {
	var raw []rawDocEntry
	if err := json.Unmarshal(payload, &raw); err != nil {
		return []discoveredRpc{}
	}
	out := make([]discoveredRpc, 0, len(raw))
	for _, e := range raw {
		if e.Name == "" {
			continue
		}
		params := make([]rpcParameter, 0, len(e.Parameters))
		for key, p := range e.Parameters {
			name := p.Name
			if name == "" {
				name = key
			}
			required := true
			if p.Required != nil {
				required = *p.Required
			} else if p.Optional != nil {
				required = !*p.Optional
			}
			params = append(params, rpcParameter{
				Name:        name,
				Type:        orDefaultString(p.Type, "string"),
				Required:    required,
				Description: p.Description,
			})
		}
		sort.Slice(params, func(i, j int) bool { return params[i].Name < params[j].Name })

		returns := make([]rpcReturnField, 0, len(e.Returns))
		for key, r := range e.Returns {
			returns = append(returns, rpcReturnField{
				Name:        key,
				Type:        orDefaultString(r.Type, "string"),
				Optional:    r.Optional,
				Description: r.Description,
			})
		}
		sort.Slice(returns, func(i, j int) bool { return returns[i].Name < returns[j].Name })

		out = append(out, discoveredRpc{
			Command:            e.Name,
			Description:        e.Description,
			DescriptionReturns: e.DescriptionReturns,
			Parameters:         params,
			Returns:            returns,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Command < out[j].Command })
	return out
}

func orDefaultString(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// customDiscoverFormatter lays out the catalog so the hierarchy is unambiguous:
// the node is the top-level heading, a "Commands:" label introduces the RPCs, each
// command is a heading indented beneath it, and its description and parameters nest one
// level deeper. Parameters carry an inline "(type, required|optional)" so a parameter row
// can never be mistaken for an RPC command. Nodes with no RPCs state "No RPC available".
func customDiscoverFormatter(out discoverOutput) (string, error) {
	if len(out.Nodes) == 0 {
		return "No RPC documentation found", nil
	}
	var b strings.Builder
	for ni, node := range out.Nodes {
		if ni > 0 {
			b.WriteString("\n")
		}
		b.WriteString(console.Colorize(console.Green, "Node: "+node.Node) + "\n")
		if len(node.Rpcs) == 0 {
			msg := "No RPC available"
			if node.Note != "" {
				msg = node.Note
			}
			b.WriteString("  " + msg + "\n")
			continue
		}
		b.WriteString("  Commands:\n")
		for i, rpc := range node.Rpcs {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString("    " + rpc.Command + "\n")
			if rpc.Description != "" {
				b.WriteString("      " + rpc.Description + "\n")
			}
			if len(rpc.Parameters) == 0 {
				b.WriteString("      Parameters: none\n")
				continue
			}
			b.WriteString("      Parameters:\n")
			var pb bytes.Buffer
			w := tabwriter.NewWriter(&pb, 0, 0, 2, ' ', 0)
			for _, p := range rpc.Parameters {
				requiredness := "optional"
				if p.Required {
					requiredness = "required"
				}
				fmt.Fprintf(w, "%s\t(%s, %s)\t%s\n", p.Name, p.Type, requiredness, strings.TrimSpace(p.Description))
			}
			_ = w.Flush()
			b.WriteString(indentLines(pb.String(), "        "))
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n", nil
}

func indentLines(s, prefix string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = prefix + l
		}
	}
	return strings.Join(lines, "\n") + "\n"
}
