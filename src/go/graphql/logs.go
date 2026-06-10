package graphql

import (
	"context"
)

type LogRequestFilterParameter struct {
	TaskId                     string   `json:"taskId"`
	Action                     string   `json:"operationNameFilter"`
	Limit                      int      `json:"limit"`
	Message                    string   `json:"messageFilter"`
	After                      DateTime `json:"startDateTime"` // Must be set
	Before                     DateTime `json:"endDateTime"`   // Must be set
	SearchString               string   `json:"searchString"`
	IncludeTraces              bool     `json:"includeTraces"`
	IncludeExceptions          bool     `json:"includeExceptions"`
	IncludeDependencies        bool     `json:"includeDependencies"`
	ExcludeEntityFrameworkLogs bool     `json:"excludeEntityFrameworkLogs"`
	ExcludeLogsWithEmptyTaskId bool     `json:"excludeLogsWithEmptyTaskId"`
}

func Logs(ctx context.Context, filter LogRequestFilterParameter) (*[]LogQueryOutput, error) {
	queryString := `query Logs($filters: LogQueryInput!) {
  logs(filters: $filters) {
    operationName
	operationId
	operationParentId
    message
    itemType
    taskId
    applicationName
    name
    source
    exceptionType
	success
	dateTime
	severityLevel
	duration
	exceptionOuterType
	exceptionOuterMessage
	exceptionFormattedMessage
  }
}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("filters", filter)

	var result struct {
		Logs []LogQueryOutput `json:"logs"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return &result.Logs, nil
}
