package snap

import (
	"fmt"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"strings"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <snapId | snapName> <arch> [flags]",
	Short: "Get information on a snap in your store",
	Args:  validateSnapInfoArgs,
	RunE:  runInfoCmd,
}

var (
	flagFleets bool
	flagCommit bool
)

func validateSnapInfoArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
		return err
	}

	switch len(args) {
	case 1:
		snapId := args[0]
		if !tools.IsValidAppId(snapId) {
			return fmt.Errorf("invalid snapId '%s'", snapId)
		}
	case 2:
		// snapName and snapDeviceArchitecture are given
		snapName := args[0]
		snapDeviceArchitecture := args[1]
		if !tools.IsValidAppName(snapName) {
			return fmt.Errorf("invalid snapName '%s'", snapName)
		}
		if !tools.IsValidDeviceArchitecture(snapDeviceArchitecture) {
			return fmt.Errorf("invalid snapDeviceArchitecture '%s'", snapDeviceArchitecture)
		}
	default:
		return fmt.Errorf("too many arguments")
	}
	return nil
}

func init() {
	SnapCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringP("rating", "r", "", "show only revisions with a specific rating "+
		"(unrated, experimental, stable, edge, deprecated, broken, denied)")
	infoCmd.Flags().Bool("fleets", true, "print details on fleets using each revisions")
	infoCmd.Flags().Bool("commit", true, "print details on revisions commit: developers message, upload date")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// TODO: Flags
	///infoCmd.Flags().Bool("declaration", true, "print snap declaration") // TODO: Do we need that section all of the time?
	//infoCmd.Flags().Bool("revisions", true, "print table of snap revisions")

	//infoCmd.Flags().StringP("revision", "r", "", "request specific revision number or \"latest\"")
	//infoCmd.Flags().String("assertion", "", "print latest assertion by type (\"snap-declaration\", \"snap-revision\" or \"all\")")
	//infoCmd.Flags().String("sets", "", "list all set names that contain a revision of this snap")
}

func retrieveSnapDatabaseId(cmd *cobra.Command, args []string) (gql.UUID, error) {
	if len(args) == 1 {
		return gql.SnapDatabaseIdBySnapId(cmd.Context(), args[0])
	} else {
		return gql.SnapDatabaseIdByNameAndArchitecture(cmd.Context(), args[0], args[1])
	}
}

type SnapInfoResult struct {
	SnapDeclaration *gql.SnapDeclaration `json:"snapDeclaration"`
}

func runInfoCmd(cmd *cobra.Command, args []string) error {
	var id gql.UUID
	var err error
	id, err = retrieveSnapDatabaseId(cmd, args)
	if err != nil {
		return err
	}

	result := SnapInfoResult{}

	var withDeclaration bool
	var withRevisions bool

	if cmd.Flags().Changed("fleets") {
		flagFleets, err = cmd.Flags().GetBool("fleets")
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("commit") {
		flagCommit, err = cmd.Flags().GetBool("commit")
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("declaration") {
		withDeclaration, err = cmd.Flags().GetBool("declaration")
		if err != nil {
			return err
		}
	}
	if cmd.Flags().Changed("revisions") {
		withRevisions, err = cmd.Flags().GetBool("revisions")
		if err != nil {
			return err
		}
	}

	// If both flags not given, print all
	if !cmd.Flags().Changed("declaration") && !cmd.Flags().Changed("revisions") {
		withDeclaration = true
		withRevisions = true
	}

	result.SnapDeclaration, err = gql.SnapDeclarationByDatabaseId(cmd.Context(), id, withDeclaration, withRevisions)
	if err != nil {
		return err
	}

	filterForRating, _ := cmd.Flags().GetString("rating")
	if filterForRating != "" {
		ratingDb, err := gql.SnapRatingByName(cmd.Context(), filterForRating)
		if err != nil {
			return err
		}
		result = filterRating(ratingDb.Name, result)
	}

	return format.PrintFormattedOutput(cmd, result, customSnapInfoFormatter)
}

func filterRating(rating string, result SnapInfoResult) SnapInfoResult {
	filteredRevisions := []gql.SnapRevision{}
	for _, revision := range result.SnapDeclaration.SnapRevisions {
		if revision.SnapRating.Name == rating {
			filteredRevisions = append(filteredRevisions, revision)
		}
	}
	result.SnapDeclaration.SnapRevisions = filteredRevisions
	return result
}

func snapDeclarationList(snapDeclaration *gql.SnapDeclaration) string {
	list := format.NewList()
	list.Add("Name", snapDeclaration.SnapName)
	list.Add("Architecture", snapDeclaration.SnapDeviceArchitecture)
	list.Add("TenantId", string(snapDeclaration.TenantId))       // TODO: resolve tenant alias
	list.Add("TenantName", string(snapDeclaration.Tenant.Alias)) // TODO: resolve tenant alias
	list.Add("Summary", strings.TrimSpace(snapDeclaration.SnapSummary))
	list.Add("Description", strings.TrimSpace(snapDeclaration.SnapDescription))
	list.Add("Base", snapDeclaration.SnapBase)
	list.Add("SnapId", snapDeclaration.SnapId)
	list.Add("DatabaseId", string(snapDeclaration.Id))
	list.Add("Created", string(snapDeclaration.CreatedAt)) // TODO: parse and reformat time?
	list.Add("AssertionId", string(snapDeclaration.AssertionId))
	return list.String()
}

func snapRevisionsTable(snapRevisions []gql.SnapRevision) string {
	table := format.NewTable(map[string]string{"revision": "Revision", "version": "Version", "developerNote": "Commit-Message",
		"size": "Size", "uploaded": "Uploaded", "rating": "Rating", "description": "Description", "fleetsCount": "#Fleets", "fleetsNames": "Fleets"})

	for _, revision := range snapRevisions {
		fleetsNames := make([]string, len(revision.Fleets))
		for i, fleet := range revision.Fleets {
			fleetsNames[i] = fleet.FleetName
		}
		table.AddRow(map[string]string{
			"revision":      fmt.Sprintf("%d", revision.Revision),
			"version":       revision.SnapVersion,
			"developerNote": tools.ShortenRight(revision.UploadMessage, 50),
			"size":          tools.FormatBytes(revision.SnapDownloadSize),
			"uploaded":      string(revision.CreatedAt), // TODO: parse and reformat time?
			"rating":        revision.SnapRating.Name,
			"description":   tools.ShortenRight(revision.Description, 50),
			"fleetsCount":   fmt.Sprintf("%d", len(revision.Fleets)),
			"fleetsNames":   strings.Join(fleetsNames, ", "),
		})
	}

	cols := []string{
		"revision",
		"version",
		"rating",
		"description",
		"size",
		"fleetsCount",
	}
	if flagCommit {
		cols = append(cols, "developerNote", "uploaded")
	}
	if flagFleets {
		cols = append(cols, "fleetsNames")
	}

	return table.Sort("revision:nasc").StringSelect(cols)
}

func customSnapInfoFormatter(res SnapInfoResult) (string, error) {
	str := ""
	str += snapDeclarationList(res.SnapDeclaration)
	str += "\n"
	str += snapRevisionsTable(res.SnapDeclaration.SnapRevisions)
	return str, nil
}
