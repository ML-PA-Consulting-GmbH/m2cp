package message

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"
	"time"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <device-name or device-serial> <date:YYYY-MM-DD>",
	Short: "List existing data message blobs for a given date",
	Long:  `Output a list of existing data message blobs.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runListCmd,
}

type listOutput struct {
	BlobName            string
	CreatedAt           string
	ModifiedAt          string
	Length              int64
	CommittedBlockCount int
	DownloadUri         string
}

func init() {
	messageCmd.AddCommand(listCmd)
}

func runListCmd(cmd *cobra.Command, args []string) error {
	deviceNameOrSerial := args[0]
	parsedDate, err := time.Parse("2006-01-02", args[1])

	if err != nil {
		return fmt.Errorf("failed parsing date. '%s' is not in YYYY-MM-DD format: %s", args[1], err)
	}

	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), deviceNameOrSerial)
	if err != nil {
		return err
	}

	resp, err := backend.GetDeviceDataMessages(cmd.Context(), deviceId, parsedDate.Format("2006/01/02"))
	if err != nil {
		return err
	}

	if resp.DataMessagesDownload == nil || len(resp.DataMessagesDownload) == 0 {
		return fmt.Errorf("no data messages found")
	}

	output := []listOutput{}
	for _, msg := range resp.DataMessagesDownload {
		output = append(output, listOutput{
			BlobName:            msg.BlobName,
			CreatedAt:           *msg.CreatedAt,
			ModifiedAt:          *msg.ModifiedAt,
			Length:              msg.Length,
			CommittedBlockCount: msg.CommittedBlockCount,
			DownloadUri:         msg.DownloadUri,
		})
	}

	return format.PrintFormattedOutput(cmd, output, listOutputFormatter)
}

func listOutputFormatter(output []listOutput) (string, error) {
	outputStr := ""

	table := format.NewTable(map[string]string{
		"name":     "Blob Name",
		"created":  "Created At",
		"modified": "Modified At",
		"size":     "Size",
		"blocks":   "Block Count",
	})

	for _, item := range output {
		parsedCreatedAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
		formattedCreatedAt := parsedCreatedAt.Format("2006-01-02 15:04:05")

		parsedModifiedAt, _ := time.Parse(time.RFC3339, item.ModifiedAt)
		formattedModifiedAt := parsedModifiedAt.Format("2006-01-02 15:04:05")

		table.AddRow(map[string]string{
			"name":     item.BlobName,
			"created":  formattedCreatedAt,
			"modified": formattedModifiedAt,
			"size":     tools.FormatBytes(item.Length),
			"blocks":   fmt.Sprintf("%d", item.CommittedBlockCount),
		})
	}

	outputStr += table.Sort("name").StringSelect([]string{"name", "created", "modified", "size", "blocks"})
	return outputStr, nil
}
