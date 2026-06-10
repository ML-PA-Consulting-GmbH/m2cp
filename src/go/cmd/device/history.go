package device

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/backend"
	"m2cpcli/format"
	"strconv"
	"time"
)

var historyCmd = &cobra.Command{
	Use:     "history <device-serial or device-name>",
	Short:   "Get installation history of the device",
	Long:    ``,
	Example: ``,
	Args:    cobra.ExactArgs(1),
	RunE:    runHistoryCmd,
}

type historyCmdOutput struct {
	Changes    *backend.GetEdgeDeviceInstallHistoriesResponse
	Snapshots  *backend.GetEdgeDeviceInstallSnapshotsResponse
	CoreAppIds []string
}

func init() {
	DeviceCmd.AddCommand(historyCmd)
	historyCmd.Flags().Bool("changes", false, "Show changes only")
	historyCmd.Flags().Bool("snapshots", false, "Show snapshots only")
	historyCmd.Flags().IntP("limit", "l", 100, "limit the number of results to the n newest")
	historyCmd.Flags().StringP("newer-than", "n", "", "only print logs newer than given duration (format: 1m, 1h)")
	historyCmd.Flags().StringP("app", "a", "", "show only logs with a specific app")
	historyCmd.Flags().StringP("after", "", "", "show only logs newer than this UTC date (format: 2006-01-02T15:04)")
	historyCmd.Flags().StringP("before", "", "", "show only logs older than this UTC date (format: 2006-01-02T15:04")
}

func runHistoryCmd(cmd *cobra.Command, args []string) error {
	deviceNameOrSerial := args[0]

	changes, _ := cmd.Flags().GetBool("changes")
	snapshots, _ := cmd.Flags().GetBool("snapshots")
	newerThan, _ := cmd.Flags().GetString("newer-than")
	after, _ := cmd.Flags().GetString("after")

	// Step 1: Validate command flags

	if changes && snapshots {
		return fmt.Errorf("both changes and snapshots cannot be provided together")
	}

	if !changes && !snapshots {
		return fmt.Errorf("either changes or snapshots must be provided")
	}

	if newerThan != "" && after != "" {
		return fmt.Errorf("newer-than and after cannot be set at the same time")
	}

	if newerThan != "" {
		_, err := time.ParseDuration(newerThan)
		if err != nil {
			return err
		}
	}

	// Step 2: Get device ID

	if deviceNameOrSerial == "" {
		return fmt.Errorf("device name or serial is required")
	}

	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), deviceNameOrSerial)
	if err != nil {
		return err
	}

	if changes {
		err = getDeviceInstallChanges(cmd, deviceId)
		if err != nil {
			return err
		}
	} else {
		err = getDeviceInstallSnapshots(cmd, deviceId)
		if err != nil {
			return err
		}
	}

	return nil
}

func getDeviceInstallChanges(cmd *cobra.Command, id string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	app, _ := cmd.Flags().GetString("app")
	after, _ := cmd.Flags().GetString("after")
	before, _ := cmd.Flags().GetString("before")
	newerThan, _ := cmd.Flags().GetString("newer-than")

	// Create a filter for the backend query

	filter := backend.EdgeDeviceInstallStateHistoryFilterInput{}
	conditions := []*backend.EdgeDeviceInstallStateHistoryFilterInput{
		{EdgeDeviceId: &backend.ComparableGuidOperationFilterInput{Eq: &id}},
	}

	if app != "" {
		conditions = append(conditions, &backend.EdgeDeviceInstallStateHistoryFilterInput{
			SnapRevision: &backend.SnapRevisionFilterInput{
				SnapDeclaration: &backend.SnapDeclarationFilterInput{
					SnapName: &backend.StringOperationFilterInput{
						Eq: &app,
					},
				},
			},
		})
	}

	if newerThan != "" {
		duration, _ := time.ParseDuration(newerThan)
		after = time.Now().Add(-duration).Format(time.RFC3339)
	}

	if after != "" {
		conditions = append(conditions, &backend.EdgeDeviceInstallStateHistoryFilterInput{
			DateTime: &backend.ComparableDateTimeOperationFilterInput{
				Gte: &after,
			},
		})
	}

	if before != "" {
		conditions = append(conditions, &backend.EdgeDeviceInstallStateHistoryFilterInput{
			DateTime: &backend.ComparableDateTimeOperationFilterInput{
				Lte: &before,
			},
		})
	}

	filter.And = conditions
	changes, err := backend.GetEdgeDeviceInstallHistories(cmd.Context(), &filter, &limit)

	if err != nil {
		return err
	}

	coreAppIds, err := backend.GetCoreAppIdsByDeviceId(cmd.Context(), id)
	if err != nil {
		return err
	}

	output := historyCmdOutput{
		Changes:    changes,
		CoreAppIds: coreAppIds,
	}

	return format.PrintFormattedOutput(cmd, output, historyOutputFormatter)
}

func getDeviceInstallSnapshots(cmd *cobra.Command, id string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	app, _ := cmd.Flags().GetString("app")
	after, _ := cmd.Flags().GetString("after")
	before, _ := cmd.Flags().GetString("before")
	newerThan, _ := cmd.Flags().GetString("newer-than")

	// Create a filter for the backend query

	filter := backend.EdgeDeviceInstallStateSnapshotFilterInput{}
	conditions := []*backend.EdgeDeviceInstallStateSnapshotFilterInput{
		{EdgeDeviceId: &backend.ComparableGuidOperationFilterInput{Eq: &id}},
	}

	if app != "" {
		conditions = append(conditions, &backend.EdgeDeviceInstallStateSnapshotFilterInput{
			SnapRevision: &backend.SnapRevisionFilterInput{
				SnapDeclaration: &backend.SnapDeclarationFilterInput{
					SnapName: &backend.StringOperationFilterInput{
						Eq: &app,
					},
				},
			},
		})
	}

	if newerThan != "" {
		duration, _ := time.ParseDuration(newerThan)
		after = time.Now().Add(-duration).Format(time.RFC3339)
	}

	if after != "" {
		conditions = append(conditions, &backend.EdgeDeviceInstallStateSnapshotFilterInput{
			DateTime: &backend.ComparableDateTimeOperationFilterInput{
				Gte: &after,
			},
		})
	}

	if before != "" {
		conditions = append(conditions, &backend.EdgeDeviceInstallStateSnapshotFilterInput{
			DateTime: &backend.ComparableDateTimeOperationFilterInput{
				Lte: &before,
			},
		})
	}

	filter.And = conditions
	snapshots, err := backend.GetEdgeDeviceInstallSnapshots(cmd.Context(), &filter, &limit)
	if err != nil {
		return err
	}

	coreAppIds, err := backend.GetCoreAppIdsByDeviceId(cmd.Context(), id)
	if err != nil {
		return err
	}

	output := historyCmdOutput{
		Snapshots:  snapshots,
		CoreAppIds: coreAppIds,
	}

	return format.PrintFormattedOutput(cmd, output, historyOutputFormatter)
}

func historyOutputFormatter(res historyCmdOutput) (string, error) {
	outputStr := ""

	if res.Changes != nil {
		changesTable := format.NewTable(map[string]string{
			"datetime": "DateTime",
			"core":     "Core",
			"action":   "Action",
			"name":     "Name",
			"version":  "Version",
			"revision": "Revision",
		})

		for _, change := range res.Changes.EdgeDeviceInstallStateHistories.Items {
			core := "-"
			for _, coreAppId := range res.CoreAppIds {
				if coreAppId == change.SnapRevision.SnapDeclarationId {
					core = "yes"
					break
				}
			}
			parsedTime, _ := time.Parse(time.RFC3339, change.DateTime)
			formattedTime := parsedTime.Format("2006-01-02 15:04:05")

			changesTable.AddRow(map[string]string{
				"datetime": formattedTime,
				"core":     core,
				"action":   change.Action,
				"name":     change.SnapRevision.SnapDeclaration.SnapName,
				"version":  change.SnapRevision.SnapVersion,
				"revision": strconv.Itoa(change.SnapRevision.Revision),
			})
		}

		outputStr += changesTable.Sorts([]string{"datetime", "asc", "core", "asc", "name", "asc", "action", "desc"}).StringSelect([]string{"datetime", "core", "action", "name", "version", "revision"})

	} else {
		tmp := ""
		tableHasItems := false
		snapshotTable := format.NewTable(map[string]string{
			"core":     "Core",
			"name":     "Name",
			"version":  "Version",
			"revision": "Revision",
		})

		for _, snapshot := range res.Snapshots.EdgeDeviceInstallStateSnapshots.Items {
			core := "-"
			for _, coreAppId := range res.CoreAppIds {
				if coreAppId == snapshot.SnapRevision.SnapDeclarationId {
					core = "yes"
					break
				}
			}
			parsedTime, _ := time.Parse(time.RFC3339, snapshot.DateTime)
			formattedTime := parsedTime.Format("2006-01-02 15:04:05")

			if tmp != snapshot.ChangeSetId {
				if tableHasItems {
					outputStr += snapshotTable.Sorts([]string{"core", "asc", "name", "asc"}).StringSelect([]string{"core", "name", "version", "revision"})

					snapshotTable = format.NewTable(map[string]string{
						"core":     "Core",
						"name":     "Name",
						"version":  "Version",
						"revision": "Revision",
					})
					tableHasItems = false

				}

				outputStr += fmt.Sprintf("\n--------------- [%s] ---------------\n", formattedTime)

				tmp = snapshot.ChangeSetId
			}

			snapshotTable.AddRow(map[string]string{
				"core":     core,
				"name":     snapshot.SnapRevision.SnapDeclaration.SnapName,
				"version":  snapshot.SnapRevision.SnapVersion,
				"revision": strconv.Itoa(snapshot.SnapRevision.Revision),
			})

			tableHasItems = true
		}

		if tableHasItems {
			outputStr += snapshotTable.Sorts([]string{"core", "asc", "name", "asc"}).StringSelect([]string{"core", "name", "version", "revision"})
		}
	}

	return outputStr, nil
}
