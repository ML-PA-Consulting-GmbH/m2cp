package snap

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"strings"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available snaps in the store",
	Args:  cobra.ExactArgs(0),
	RunE:  runListCmd,
}

func init() {
	SnapCmd.AddCommand(listCmd)

	// Todo: Below flags are not implemented yet
	listCmd.Flags().StringP("arch", "a", "", "filter by architecture (amd64, arm64)")
	//listCmd.Flags().String("sort", "snapId", "sort by (snapName, snapBase, snapDeviceArchitecture, snapId)")

	//listCmd.Flags().String("long", "", "long format with more details")
	//listCmd.Flags().StringP("mlpa", "d", "", "list snaps provided by mlpa (operating system and essential tools)") // TODO: Filter by tenant?
	//listCmd.Flags().StringP("name", "n", "", "filter by name, supports * as wildcard") // TODO: Can the backend support this kind of filter?
}

func buildQueryString(cmd *cobra.Command) (string, error) {
	queryTemplate := `query ListSnapDeclarations($take: Int, $skip: Int){
result: snapDeclarations(
take: $take
skip: $skip
{{ if ne .SortKey "" -}}
order: { {{- .SortKey -}}:ASC}
{{- else -}}
order: {id:ASC}
{{- end -}}
where: {
	{{- with .Architecture -}}
		snapDeviceArchitecture: {eq: {{- printf "%q" . -}} }
	{{- end }}
}
)
{
    items{
		id
        snapId
		snapBase
		snapDeviceArchitecture
		snapName
		snapDescription
		tenantId
    }
    pageInfo{ hasNextPage hasPreviousPage }  
    totalCount  
  }
}`

	var err error
	type Interval struct {
		Begin string
		End   string
	}
	type Parameters struct {
		Architecture string
		SortKey      string
	}
	params := Parameters{}

	if cmd.Flags().Changed("arch") {
		params.Architecture, err = cmd.Flags().GetString("arch")
		if err != nil {
			return "", err
		}
	}

	if cmd.Flags().Changed("sort") {
		params.SortKey, err = cmd.Flags().GetString("sort")
		if err != nil {
			return "", err
		}
	}

	return gql.RenderQueryOrMutation(queryTemplate, params)
}

type SnapListResult struct {
	TotalCount int                   `json:"totalCount"`
	Snaps      []gql.SnapDeclaration `json:"snapDeclarations"`
	Tenants    map[gql.UUID]string   `json:"tenants"`
}

func runListCmd(cmd *cobra.Command, args []string) error {
	queryString, err := buildQueryString(cmd)
	if err != nil {
		return err
	}

	tenants, err := gql.Tenants(cmd.Context())
	if err != nil {
		return err
	}
	if len(*tenants) == 0 {
		return fmt.Errorf("no tenants found")
	}

	client, req := gql.PrepareClientAndRequest(cmd.Context(), queryString)
	snapDeclarations, err := gql.RunPaginated[gql.SnapDeclaration](cmd.Context(), client, req)
	if err != nil {
		return err
	}

	var snaps = SnapListResult{
		TotalCount: len(snapDeclarations),
		Snaps:      snapDeclarations,
		Tenants:    tenantsAsMap(*tenants),
	}

	return format.PrintFormattedOutput(cmd, snaps, customSnapListFormatter)
}

func customSnapListFormatter(res SnapListResult) (string, error) {
	table := format.NewTable(map[string]string{
		"1id":     "Snap Id",
		"2arch":   "Arch",
		"3name":   "Alias",
		"4tenant": "Tenant",
		"5desc":   "Description",
	})
	for _, snap := range res.Snaps {
		table.AddRow(map[string]string{
			"1id":     snap.SnapId,
			"2arch":   snap.SnapDeviceArchitecture,
			"3name":   strings.TrimSpace(tools.ShortenRight(snap.SnapName, 25)),
			"4tenant": res.Tenants[snap.TenantId],
			"5desc":   strings.TrimSpace(tools.ShortenRight(snap.SnapDescription, 45)), // TODO: replace newlines and tabs?
		})
	}

	return table.String(), nil
}

func tenantsAsMap(tenants []gql.Tenant) map[gql.UUID]string {
	tenantMap := make(map[gql.UUID]string)
	for _, tenant := range tenants {
		tenantMap[tenant.Id] = tenant.Alias
	}
	return tenantMap
}
