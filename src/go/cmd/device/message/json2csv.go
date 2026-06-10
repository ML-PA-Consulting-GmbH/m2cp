package message

import (
	"fmt"
	"m2cpcli/format"
	"m2cpcli/tools"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var json2csvCmd = &cobra.Command{
	Use:   "json2csv <in-path> <out-path>",
	Short: "Scan <in-path> recursively for json files containing data points and convert them to csv files in <out-path>",
	Long:  "Scan <in-path> recursively for json files containing data points (format as used by cloud in storage account) and convert them to csv files in <out-path>",
	Args:  cobra.ExactArgs(2),
	RunE:  runJson2CsvCmd,
}

func init() {
	messageCmd.AddCommand(json2csvCmd)
}

func runJson2CsvCmd(cmd *cobra.Command, args []string) error {

	var result struct {
		FilesScanned   int `json:"files-scanned"`
		FilesConverted int `json:"files-converted"`
	}

	pathIn, err := tools.Abspath(args[0])
	if err != nil {
		return fmt.Errorf("could not make input path absolute: %v", err)
	}

	pathOut, err := tools.Abspath(args[1])
	if err != nil {
		return fmt.Errorf("could not make output path absolute: %v", err)
	}

	if _, err := os.Stat(pathIn); os.IsNotExist(err) {
		return fmt.Errorf("input path does not exist: %v", pathIn)
	}

	if _, err := os.Stat(pathOut); os.IsNotExist(err) {
		return fmt.Errorf("output path does not exist: %v", pathOut)
	}

	err = filepath.Walk(pathIn, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			result.FilesScanned++
			_, err := parseFile(path, pathOut, nil)
			if err != nil {
				return err
			}
			result.FilesConverted++
		}
		return nil
	})

	return format.PrintFormattedOutput(cmd, result, nil)
}
