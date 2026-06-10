package store

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"sort"
	"strconv"
	"time"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get logs from store",
	Args:  validateLogsCmdArgs,
	RunE:  runLogsCmd,
}

const filterTimeFormat = "2006-01-02T15:04"
const timeFormatLogEntry = "2006-01-02 15:04:05.000"

func init() {
	StoreCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringP("task", "t", "", "show only logs with a specific task id")
	logsCmd.Flags().StringP("search", "s", "", "search term to filter logs by multiple fields")
	logsCmd.Flags().StringP("action", "a", "", "search term for a specific action")
	logsCmd.Flags().IntP("limit", "l", 100, "limit the number of results to the n newest")
	logsCmd.Flags().String("after", "", "show only logs newer than this UTC date (format: "+filterTimeFormat+")")
	logsCmd.Flags().String("before", "", "show only logs older than this UTC date (format: "+filterTimeFormat+")")
	logsCmd.Flags().StringP("newer-than", "n", "", "show only logs newer than this duration (format: 1m, 1h)")
	logsCmd.Flags().Bool("include-dependencies", false, "include dependencies in the logs")
	logsCmd.Flags().BoolP("get-all-task-logs", "g", false, "retrieve all logs of task ids when not filtering for a specific task id"+
		"\n(please use with caution, possibly lots of backend requests)")
}

func validateLogsCmdArgs(cmd *cobra.Command, args []string) error {
	err := cobra.ExactArgs(0)(cmd, args)
	if err != nil {
		return err
	}

	// Validate that task is not set when get-all-task-logs is used
	if cmd.Flag("get-all-task-logs").Value.String() == "true" && cmd.Flag("task").Value.String() != "" {
		return fmt.Errorf("task cannot be set when get-all-task-logs is used")
	}

	// Validate that newer-than and after+before are not set at the same time
	if cmd.Flag("newer-than").Value.String() != "" && cmd.Flag("after").Value.String() != "" {
		return fmt.Errorf("newer-than and after cannot be set at the same time")
	}
	if cmd.Flag("newer-than").Value.String() != "" && cmd.Flag("before").Value.String() != "" {
		return fmt.Errorf("newer-than and before cannot be set at the same time")
	}
	if cmd.Flag("newer-than").Value.String() != "" {
		_, err := time.ParseDuration(cmd.Flag("newer-than").Value.String())
		if err != nil {
			return fmt.Errorf("invalid duration format for newer-than: %s", err)
		}
	}

	return nil
}

func runLogsCmd(cmd *cobra.Command, args []string) error {
	filterParameter, err := buildFilterParameter(cmd)
	if err != nil {
		return err
	}

	logs, err := gql.Logs(cmd.Context(), *filterParameter)
	if err != nil {
		return err
	}

	// If search and deep search is set, get full logs for each task id
	if cmd.Flag("get-all-task-logs").Value.String() == "true" && cmd.Flag("search").Value.String() != "" {
		logs, err = getAllTaskLogs(cmd, filterParameter, logs)
		if err != nil {
			return err
		}
	}

	groupedAndSortedLogs := groupLogsByTaskIdSortedByTime(logs)
	return format.PrintFormattedOutput(cmd, groupedAndSortedLogs, customLogsFormatter)
}

func getAllTaskLogs(cmd *cobra.Command, filterParameter *gql.LogRequestFilterParameter, logs *[]gql.LogQueryOutput) (*[]gql.LogQueryOutput, error) {
	var logsToReturn []gql.LogQueryOutput

	// Get all unique task ids
	taskIds := make(map[string]bool)
	for _, log := range *logs {
		taskIds[log.TaskId] = true
	}

	// Get all logs for each task id
	for taskId := range taskIds {
		filterParameter.TaskId = taskId
		filterParameter.SearchString = ""
		filterParameter.Action = ""

		taskLogs, err := gql.Logs(cmd.Context(), *filterParameter)
		if err != nil {
			return nil, fmt.Errorf("Error getting logs for task id %s: %s\n", taskId, err)
		}
		logsToReturn = append(logsToReturn, *taskLogs...)
	}
	return &logsToReturn, nil
}

func buildFilterParameter(cmd *cobra.Command) (*gql.LogRequestFilterParameter, error) {
	searchTerm := cmd.Flag("search").Value.String()
	if searchTerm != "" && len(searchTerm) < 3 {
		// Policy comes from AppInsights, less than 3 characters returns always zero results
		return nil, fmt.Errorf("search term must be at least 3 characters long")
	}

	var filterParameter gql.LogRequestFilterParameter
	filterParameter.TaskId = cmd.Flag("task").Value.String()
	filterParameter.Action = cmd.Flag("action").Value.String()
	filterParameter.SearchString = cmd.Flag("search").Value.String()
	filterParameter.IncludeDependencies = cmd.Flag("include-dependencies").Value.String() == "true"
	filterParameter.IncludeTraces = true
	filterParameter.IncludeExceptions = true
	filterParameter.ExcludeEntityFrameworkLogs = true
	filterParameter.ExcludeLogsWithEmptyTaskId = true

	before := cmd.Flag("before").Value.String()
	if before == "" {
		filterParameter.Before = gql.DateTime(time.Now().Format(time.RFC3339))
	} else {
		t, err := time.Parse(filterTimeFormat, before)
		if err != nil {
			return nil, err
		}
		filterParameter.Before = gql.DateTime(t.Format(time.RFC3339))
	}

	after := cmd.Flag("after").Value.String()
	if after == "" {
		// default to 24 hours ago (same default as backend, when parameter not given)
		filterParameter.After = gql.DateTime(time.Now().Add(-24 * time.Hour).Format(time.RFC3339))
	} else {
		t, err := time.Parse(filterTimeFormat, after)
		if err != nil {
			return nil, err
		}
		filterParameter.After = gql.DateTime(t.Format(time.RFC3339))
	}

	newerThan := cmd.Flag("newer-than").Value.String()
	if newerThan != "" {
		duration, err := time.ParseDuration(newerThan)
		if err != nil {
			return nil, err
		}
		filterParameter.After = gql.DateTime(time.Now().Add(-duration).Format(time.RFC3339))
	}

	limitStr := cmd.Flag("limit").Value.String()
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil, err
	}
	filterParameter.Limit = limit

	return &filterParameter, nil
}

type logsByTaskId struct {
	TaskId           string
	FirstLogDateTime gql.DateTime
	Logs             []gql.LogQueryOutput
}

func groupLogsByTaskIdSortedByTime(logs *[]gql.LogQueryOutput) []logsByTaskId {
	// Group logs by task id
	groupedLogs := make(map[string][]gql.LogQueryOutput)
	for _, log := range *logs {
		groupedLogs[log.TaskId] = append(groupedLogs[log.TaskId], log)
	}

	// Sort logs by time
	var logsByTaskIdSortedByTime []logsByTaskId
	for taskId, logs := range groupedLogs {
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].DateTime <= logs[j].DateTime
		})
		logsByTaskIdSortedByTime = append(logsByTaskIdSortedByTime, logsByTaskId{taskId, logs[0].DateTime, logs})
	}

	// Sort logs by time
	sort.Slice(logsByTaskIdSortedByTime, func(i, j int) bool {
		return logsByTaskIdSortedByTime[i].FirstLogDateTime <= logsByTaskIdSortedByTime[j].FirstLogDateTime
	})

	return logsByTaskIdSortedByTime
}

func customLogsFormatter(res []logsByTaskId) (string, error) {
	if len(res) == 0 {
		return "No logs found", nil
	}

	outputStr := ""
	for _, task := range res {
		if task.TaskId == "" {
			continue
		}
		taskFirstTime := tools.ReformatTime(string(task.FirstLogDateTime), timeFormatLogEntry)
		outputStr += fmt.Sprintf("\n──┬─[%s tid %s]────────────────────────────────────────\n", taskFirstTime, task.TaskId)

		for _, log := range task.Logs {
			errorFlag := ""
			if log.Success == "False" || log.ItemType == "exception" {
				errorFlag = "ERROR "
			}
			timestamp, err := time.Parse(time.RFC3339, string(log.DateTime))
			if err != nil {
				return "", err
			}

			if log.ItemType == "exception" {
				log.Message = fmt.Sprintf("%s (%s): %s", log.ExceptionFormattedMessage, log.ExceptionType, log.ExceptionOuterMessage)
			}

			outputStr += fmt.Sprintf("  ├─[%s%s %s] %s\n", errorFlag, timestamp.Format(timeFormatLogEntry), log.OperationName, log.Message)
		}
	}
	return outputStr, nil
}
