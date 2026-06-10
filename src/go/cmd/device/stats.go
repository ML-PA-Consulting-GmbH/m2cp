package device

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"sort"
	"strconv"
)

var statsCmd = &cobra.Command{
	Use:   "stats [deviceSerial|deviceName]",
	Short: "Get live system statistics of a device",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatsCmd,
}

func init() {
	DeviceCmd.AddCommand(statsCmd)
	// Todo: Implement  flags
	//statsCmd.Flags().Bool("full", false, "Retrieve full statistics")
	//statsCmd.Flags().Bool("raw", false, "Disable filtering/renaming of disk/processes")
}

func runStatsCmd(cmd *cobra.Command, args []string) error {
	deviceId, err := gql.DeviceIdByNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	device, err := gql.DeviceByDeviceId(cmd.Context(), deviceId)
	if err != nil {
		return err
	}

	queryString := `query($serial: String!){
result: edgeDeviceSystemStats(input: {deviceSerial: $serial}){
	logsUsageBytes
	statistics {
      overallMemoryTotalBytes
      overallMemoryUsedBytes
      diskInformation {
        partitionName
        partitionSizeBytes
        partitionUsedBytes
      }
      processInformation {
        processCpuUsagePercentage
        processMemoryUsedBytes
        processName
      }
    }
}}`
	// overallCpuUsagePercentage
	client, req := gql.PrepareClientAndRequest(cmd.Context(), queryString)
	req.Var("serial", device.DeviceSerial)

	var result struct {
		Result gql.EdgeDeviceSystemStatsOutput `json:"result"`
	}
	err = client.Run(cmd.Context(), req, &result)
	if err != nil {
		return err
	}

	// TODO: all zero-values?
	return format.PrintFormattedOutput(cmd, result.Result, customStatisticFormatter)
}

func generalListSection(res gql.EdgeDeviceSystemStatsOutput) string {
	list := format.NewList()
	list.Add("CPU", fmt.Sprintf("%3.2f %%", float32(res.Statistics.OverallCpuUsagePercentage)))
	memPercent := float32(res.Statistics.OverallMemoryUsedBytes) / float32(res.Statistics.OverallMemoryTotalBytes)
	list.Add("MEM", fmt.Sprintf("%s / %s (%.2f %%)",
		helper.SizeWithBinaryUnit(res.Statistics.OverallMemoryUsedBytes),
		helper.SizeWithBinaryUnit(res.Statistics.OverallMemoryTotalBytes),
		memPercent*100.0))

	logPercent := float32(res.LogsUsageBytes) / float32(res.Statistics.OverallMemoryTotalBytes)
	list.Add("LOG", fmt.Sprintf("%s / %s (%3.2f %%)",
		helper.SizeWithBinaryUnit(res.LogsUsageBytes),
		helper.SizeWithBinaryUnit(res.Statistics.OverallMemoryTotalBytes),
		logPercent*100.0))
	return list.String()
}

func diskInformationSection(res gql.EdgeDeviceSystemStatsOutput) string {
	table := format.NewTable(map[string]string{
		"1name":    "Partition Name",
		"2used":    "Used",
		"3total":   "Total",
		"4percent": "Percent Used",
	})
	diskInfo := res.Statistics.DiskInformation
	sort.Slice(diskInfo[:], func(i, j int) bool {
		percentI := float32(diskInfo[i].PartitionUsedBytes) / float32(diskInfo[i].PartitionSizeBytes)
		percentJ := float32(diskInfo[j].PartitionUsedBytes) / float32(diskInfo[j].PartitionSizeBytes)
		return percentI > percentJ
	})
	for _, partition := range diskInfo {
		percentage := 100.0 * float32(partition.PartitionUsedBytes) / float32(partition.PartitionSizeBytes)
		table.AddRow(map[string]string{
			"1name":    partition.PartitionName,
			"2used":    helper.SizeWithBinaryUnit(partition.PartitionUsedBytes),
			"3total":   helper.SizeWithBinaryUnit(partition.PartitionSizeBytes),
			"4percent": fmt.Sprintf("%2.2f %%", percentage),
		})
	}

	return table.String()
}

func processInformationSection(res gql.EdgeDeviceSystemStatsOutput) string {
	table := format.NewTable(map[string]string{
		"1name": "Process Name",
		"2cpu":  "CPU (%)",
		"3mem":  "MEM",
	})
	procInfo := res.Statistics.ProcessInformation
	sort.Slice(procInfo[:], func(i, j int) bool {
		return procInfo[i].ProcessMemoryUsedBytes > procInfo[j].ProcessMemoryUsedBytes
	})

	for _, proc := range res.Statistics.ProcessInformation {
		table.AddRow(map[string]string{
			"1name": proc.ProcessName,
			"2cpu":  strconv.FormatFloat(proc.ProcessCpuUsagePercentage, 'f', 2, 32),
			"3mem":  helper.SizeWithBinaryUnit(proc.ProcessMemoryUsedBytes),
		})
	}

	return table.String()
}

func customStatisticFormatter(res gql.EdgeDeviceSystemStatsOutput) (string, error) {
	output := generalListSection(res) + "\n" + diskInformationSection(res) + "\n" + processInformationSection(res)

	return output, nil
}
