package message

import (
	"fmt"
	"m2cp"
	"m2cp/m2cp_new"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var dataCmd = &cobra.Command{
	Use:   "data <device> <date>",
	Short: "Fetch data points sent by <device> on <date:YYYY-MM-DD> from cloud storage account.",
	Long:  "Fetch data points sent by <device> on <date:YYYY-MM-DD> from cloud storage account.",
	Args:  cobra.RangeArgs(2, 2),
	RunE:  runDataCmd,
}

func init() {
	messageCmd.AddCommand(dataCmd)

	dataCmd.Flags().StringP("outpath", "o", "", "if set: a directory to write the unaltered JSON blobs (all parts) to. Files from an earlier download are replaced.")
	dataCmd.Flags().String("topic", "*", "select data points containing <topic>")
	dataCmd.Flags().String("origin", "*", "select data points containing <origin>")
	dataCmd.Flags().String("print", "", "print ("+printData+"|"+listTopic+"|"+listOrigin+"|"+listCombined+") found in the data points - don't print the data itself")
	dataCmd.Flags().IntP("number", "n", 0, "print only the last <number> data points")
}

const (
	printData    = "data"
	listTopic    = "topics"
	listOrigin   = "origins"
	listCombined = "combined"
)

type DataResult struct {
	BlobsFound int      `json:"blobs-found"`
	DataPoints int      `json:"data-points-parsed"`
	Files      []string `json:"files,omitempty"`
}

type dataPoint struct {
	Topic    string
	Origin   string
	DateTime string
	Data     map[string]string
}

func runDataCmd(cmd *cobra.Command, args []string) error {
	var err error
	ctp := m2cp_new.ContextPlusFromContext(cmd.Context())

	parsedDate, err := time.Parse("2006-01-02", args[1])
	if err != nil {
		return fmt.Errorf("failed parsing date. '%s' is not in YYYY-MM-DD format: %s", args[1], err)
	}

	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(ctp, args[0])
	if err != nil {
		return err
	}

	printMode := cmd.Flags().Lookup("print").Value.String()
	if printMode != "" && printMode != listTopic && printMode != listOrigin && printMode != listCombined && printMode != printData {
		return fmt.Errorf("invalid value for --print: %s", printMode)
	}

	outPath := cmd.Flags().Lookup("outpath").Value.String()
	var callback ParserCallback = nil
	if outPath != "" {
		if outPath, err = tools.Abspath(outPath); err != nil {
			return fmt.Errorf("could not make output path absolute: %s", err)
		}
	}

	topicSearch := cmd.Flag("topic").Value.String()
	var filterTopic func(string) bool
	if topicSearch == "*" || topicSearch == "" {
		filterTopic = func(topic string) bool {
			return true
		}
	} else {
		pattern := "^" + regexp.QuoteMeta(topicSearch) + "$"
		pattern = strings.ReplaceAll(pattern, "\\*", ".*")
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern for --topic: %s", topicSearch)
		}
		filterTopic = func(topic string) bool {
			return regex.MatchString(topic)
		}
	}

	originSearch := cmd.Flag("origin").Value.String()
	var filterOrigin func(string) bool
	if originSearch == "*" || originSearch == "" {
		filterOrigin = func(origin string) bool {
			return true
		}
	} else {
		pattern := "^" + regexp.QuoteMeta(originSearch) + "$"
		pattern = strings.ReplaceAll(pattern, "\\*", ".*")
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern for --origin: %s", originSearch)
		}
		filterOrigin = func(origin string) bool {
			return regex.MatchString(origin)
		}
	}

	listResult := make(map[string]int)
	dataPointsResult := make([]dataPoint, 0)
	if printMode == listTopic {
		callback = func(topic, origin, datetime string, data map[string]string) {
			if filterTopic(topic) && filterOrigin(origin) {
				listResult[topic]++
			}
		}
	} else if printMode == listOrigin {
		callback = func(topic, origin, datetime string, data map[string]string) {
			if filterTopic(topic) && filterOrigin(origin) {
				listResult[origin]++
			}
		}
	} else if printMode == listCombined {
		callback = func(topic, origin, datetime string, data map[string]string) {
			if filterTopic(topic) && filterOrigin(origin) {
				listResult[topic+" "+origin]++
			}
		}
	} else if printMode == printData {
		callback = func(topic, origin, datetime string, data map[string]string) {
			if filterTopic(topic) && filterOrigin(origin) {
				dataPointsResult = append(dataPointsResult, dataPoint{
					Topic:    topic,
					Origin:   origin,
					DateTime: datetime,
					Data:     data,
				})
			}
		}
	}

	res, err := ProcessDataMessages(ctp, deviceId, parsedDate, outPath, callback)
	if err != nil {
		return err
	}

	if printMode == printData {
		if number := cmd.Flag("number").Value.String(); number != "" {
			n, err := strconv.Atoi(number)
			if err != nil {
				return fmt.Errorf("invalid value for --number: %s", number)
			}
			if n < 0 {
				return fmt.Errorf("invalid value for --number: %s", number)
			}
			if n > 0 && n < len(dataPointsResult) {
				dataPointsResult = dataPointsResult[len(dataPointsResult)-n:]
			}
		}
		return format.PrintFormattedOutput(cmd, dataPointsResult, nil)
	} else if printMode != "" {
		return format.PrintFormattedOutput(cmd, listResult, nil)
	}
	return format.PrintFormattedOutput(cmd, res, nil)
}

func ProcessDataMessages(ctp m2cp.ContextPlus, deviceId string, date time.Time, outPath string, callback ParserCallback) (DataResult, error) {
	res := DataResult{}

	blobs, err := backend.GetDeviceDataMessages(ctp, deviceId, date.Format("2006/01/02"))
	if err != nil {
		return res, err
	}

	if len(blobs.DataMessagesDownload) == 0 {
		return res, fmt.Errorf("no data messages found")
	}

	fmt.Printf("Found %d data message blob(s) to download\n", len(blobs.DataMessagesDownload))

	if outPath != "" {
		if err := os.MkdirAll(outPath, 0755); err != nil {
			return res, fmt.Errorf("could not create output path: %w", err)
		}
	}

	for i, blob := range blobs.DataMessagesDownload {
		res.BlobsFound++

		file, dataPointsParsed, err := downloadAndParseBlob(date.Format("2006-01-02"), blob.BlobName, blob.DownloadUri, outPath, callback, i+1, len(blobs.DataMessagesDownload))
		if err != nil {
			return res, err
		}
		if file != "" {
			res.Files = append(res.Files, file)
		}
		res.DataPoints += dataPointsParsed
	}

	return res, nil
}

// downloadAndParseBlob keeps the unaltered blob in outPath (if set) and feeds callback (if set).
func downloadAndParseBlob(day, blobName, downloadUri, outPath string, callback ParserCallback, fileNum, totalFiles int) (string, int, error) {
	var f *os.File
	var err error
	finalPath := ""
	if outPath != "" {
		f, finalPath, err = createPartFile(outPath, day, blobName)
	} else if callback != nil {
		f, err = os.CreateTemp("", "m2cp-download-")
	} else {
		return "", 0, nil
	}
	if err != nil {
		return "", 0, err
	}
	partPath := f.Name()
	filePath := partPath
	defer func() {
		f.Close()
		os.Remove(partPath) // no-op once the .part file has been renamed
	}()

	label := fmt.Sprintf("[%d/%d]", fileNum, totalFiles)
	fmt.Printf("%s Starting download...\n", label)
	if err := downloadTo(f, downloadUri, label); err != nil {
		fmt.Println()
		return "", 0, err
	}
	if err := f.Close(); err != nil {
		return "", 0, fmt.Errorf("failed to write file: %w", err)
	}
	if finalPath != "" {
		if err := os.Rename(partPath, finalPath); err != nil {
			return "", 0, fmt.Errorf("failed to replace %s: %w", finalPath, err)
		}
		filePath = finalPath
	}

	dataPointsParsed, err := parseFile(filePath, "", callback)
	if err != nil {
		return filePath, 0, fmt.Errorf("failed to parse file: %w", err)
	}

	return filePath, dataPointsParsed, nil
}
