package tenant

import (
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all tenants",
	RunE:  runListCmd,
}

func init() {
	tenantCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// UserCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// UserCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type TenantListResult struct {
	TotalCount int          `json:"totalCount"`
	Tenants    []gql.Tenant `json:"tenants"`
}

func runListCmd(cmd *cobra.Command, args []string) error {
	// Note, this client is not responsible to hide other tenants, a given user should not see.
	queryString := `query($take: Int, $skip: Int){
result: tenants(take: $take, skip: $skip, order: {id:ASC}){
	items{
		id
		tenantName
		alias
    }
    pageInfo{ hasNextPage hasPreviousPage }  
    totalCount
}}`

	client, req := gql.PrepareClientAndRequest(cmd.Context(), queryString)
	tenants, err := gql.RunPaginated[gql.Tenant](cmd.Context(), client, req)
	if err != nil {
		return err
	}

	var result = TenantListResult{
		TotalCount: len(tenants),
		Tenants:    tenants,
	}

	return format.PrintFormattedOutput(cmd, result, customTenantListFormatter)
}

func customTenantListFormatter(res TenantListResult) (string, error) {
	table := format.NewTable(map[string]string{
		"1id":    "Id",
		"2alias": "Alias",
		"3name":  "Alias",
	})
	for _, tenant := range res.Tenants {
		table.AddRow(map[string]string{
			"1id":    string(tenant.Id),
			"2alias": tenant.Alias,
			"3name":  strings.TrimSpace(tools.ShortenRight(tenant.TenantName, 70)),
		})
	}

	return table.String(), nil
}
