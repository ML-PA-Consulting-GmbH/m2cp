package dgroup

import (
	"fmt"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Deployment Groups",
	Args:  cobra.ExactArgs(0),
	RunE:  runListCmd,
}

var sort string

func init() {
	listCmd.Flags().String("sort", "id", "sort by column (e.g. Name), sort order can be modified: Name:desc")
	listCmd.Flags().String("type", "", "filter by type")
	listCmd.Flags().String("model", "", "filter by model")
	listCmd.Flags().Int("revision", 0, "filter by revision")
	listCmd.Flags().String("arch", "", "filter by architecture")
	listCmd.Flags().String("admin", "", "filter by admin")
	listCmd.Flags().Bool("my", false, "show only fleets owned by you")
	FleetCmd.AddCommand(listCmd)
	DGroupCmd.AddCommand(listCmd)
}

func buildQueryString(cmd *cobra.Command) (string, error) {
	var err error

	queryTemplate := `query ListAllFleets($take: Int, $skip: Int){
		result: fleets(
			take: $take
			skip: $skip
			order: {fleetName:ASC}
			where: {
				{{- with .Type -}}
					edgeDeviceModel: { modelType: {eq: {{- printf "%q" . -}} } }
				{{- end }}
				{{- with .Model -}}
					edgeDeviceModel: { modelName: {eq: {{- printf "%q" . -}} } }
				{{- end }}
				{{- with .Revision -}}
					edgeDeviceModel: { modelRevision: {eq: {{- printf "%d" . -}} } }
				{{- end }}
				{{- with .Architecture -}}
					edgeDeviceModel: { modelArchitecture: {eq: {{- printf "%q" . -}} } }
				{{- end }}
				{{- with .Admin -}}
					ownerUserId: {eq: {{- printf "%q" . -}} }
				{{- end }}
			}
		)
		{
			items{
				id
				fleetName
				description
				architecture
				ownerUser{
					id: userId
					displayName
					email
				}
				edgeDeviceModel {
					id
					modelType
					modelName
					modelRevision
					modelArchitecture
				}
			}
			pageInfo{ hasNextPage hasPreviousPage }  
			totalCount  
		  }
		}`

	type Parameters struct {
		Type         string
		Model        string
		Revision     int
		Architecture string
		Admin        string
	}

	params := Parameters{}

	if cmd.Flags().Changed("type") {
		params.Type, err = cmd.Flags().GetString("type")
		if err != nil {
			return "", err
		}
	}

	if cmd.Flags().Changed("model") {
		params.Model, err = cmd.Flags().GetString("model")
		if err != nil {
			return "", err
		}
	}

	if cmd.Flags().Changed("revision") {
		params.Revision, err = cmd.Flags().GetInt("revision")
		if err != nil {
			return "", err
		}
	}

	if cmd.Flags().Changed("arch") {
		params.Architecture, err = cmd.Flags().GetString("arch")
		if err != nil {
			return "", err
		}
	}

	if cmd.Flags().Changed("my") {
		var userInfo *gql.User
		userInfo, err = gql.UserInfo(cmd.Context())
		if err != nil {
			return "", err
		}
		params.Admin = string(userInfo.Id)
	} else if cmd.Flags().Changed("admin") {
		if adminEmail, err1 := cmd.Flags().GetString("admin"); err1 != nil {
			return "", err1
		} else if user, err2 := gql.UserByEmail(cmd.Context(), adminEmail); err2 != nil {
			return "", err
		} else {
			params.Admin = string(user.Id)
		}
	}

	return gql.RenderQueryOrMutation(queryTemplate, params)
}

type FleetListResult struct {
	TotalCount int         `json:"totalCount"`
	Fleets     []gql.Fleet `json:"fleets"`
}

func runListCmd(cmd *cobra.Command, args []string) error {
	queryString, err := buildQueryString(cmd)
	if err != nil {
		return err
	}

	client, request := gql.PrepareClientAndRequest(cmd.Context(), queryString)
	fleets, err := gql.RunPaginated[gql.Fleet](cmd.Context(), client, request)
	if err != nil {
		return err
	}

	var result = FleetListResult{
		TotalCount: len(fleets),
		Fleets:     fleets,
	}

	if sort, err = cmd.Flags().GetString("sort"); err != nil {
		return fmt.Errorf("failed getting sort flag: %s", err)
	}

	return format.PrintFormattedOutput(cmd, result, customFleetListFormatter)
}

func customFleetListFormatter(res FleetListResult) (string, error) {
	table := format.NewTable(map[string]string{
		"1id":       "Id",
		"2name":     "Name",
		"3type":     "Type",
		"4model":    "Model",
		"5revision": "Rev",
		"6arch":     "Arch",
		"7admin":    "Admin",
		"8desc":     "Description",
	})
	for _, fleet := range res.Fleets {
		table.AddRow(map[string]string{
			"1id":       string(fleet.Id),
			"2name":     tools.ShortenRight(fleet.FleetName, 36), // Note, if the name was set automatically to be the device' serial, then we don't want that to be shortened
			"3type":     fleet.Model.Type,
			"4model":    fleet.Model.Name,
			"5revision": strconv.Itoa(fleet.Model.Revision),
			"6arch":     fleet.Model.Architecture,
			"7admin":    fleet.Admin.Email,
			"8desc":     strings.TrimSpace(tools.ShortenRight(fleet.Description, 45)),
		})
	}

	if sort != "" {
		table = table.Sort(sort)
	}

	return table.String(), nil
}
