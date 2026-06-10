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

	"github.com/cavaliergopher/grab/v3"
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

	dataCmd.Flags().StringP("outpath", "o", "", "if set: a directory to write data points as csv files. Existing files will be appended to.")
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
	BlobsFound int `json:"blobs-found"`
	DataPoints int `json:"data-points-parsed"`
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

	for i, blob := range blobs.DataMessagesDownload {
		res.BlobsFound++

		dataPointsParsed, err := downloadAndParseBlob(blob.DownloadUri, outPath, callback, i+1, len(blobs.DataMessagesDownload))
		if err != nil {
			return res, err
		}
		res.DataPoints += dataPointsParsed
	}

	return res, nil
}

func downloadAndParseBlob(downloadUri, outPath string, callback ParserCallback, fileNum, totalFiles int) (int, error) {
	tmpFile, err := os.CreateTemp("", "m2cp-download-")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close() // Close immediately so grab can write to it

	// Use grab for robust downloading with automatic resume support
	client := grab.NewClient()
	req, err := grab.NewRequest(tmpFile.Name(), downloadUri)
	if err != nil {
		return 0, fmt.Errorf("failed to create download request: %w", err)
	}

	fmt.Printf("[%d/%d] Starting download...\n", fileNum, totalFiles)

	resp := client.Do(req)

	// Track download progress
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if resp.Size() > 0 {
				fmt.Printf("\r[%d/%d] Downloading... %.2f%% (%.2f / %.2f MB) @ %.2f MB/s",
					fileNum, totalFiles,
					100*resp.Progress(),
					float64(resp.BytesComplete())/1024/1024,
					float64(resp.Size())/1024/1024,
					resp.BytesPerSecond()/1024/1024)
			} else {
				fmt.Printf("\r[%d/%d] Downloading... %.2f MB @ %.2f MB/s",
					fileNum, totalFiles,
					float64(resp.BytesComplete())/1024/1024,
					resp.BytesPerSecond()/1024/1024)
			}
		case <-resp.Done:
			if err := resp.Err(); err != nil {
				fmt.Println() // New line after progress
				return 0, fmt.Errorf("failed to download blob: %w", err)
			}
			// Clear the line by overwriting with spaces, then print completion message
			fmt.Printf("\r%-80s\r[%d/%d] Download complete: %.2f MB\n", "", fileNum, totalFiles, float64(resp.BytesComplete())/1024/1024)
			goto downloadComplete
		}
	}

downloadComplete:

	dataPointsParsed, err := parseFile(tmpFile.Name(), outPath, callback)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file: %w", err)
	}

	return dataPointsParsed, nil
}
