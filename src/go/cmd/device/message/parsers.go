package message

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Parse JSON
type messageType struct {
	Payload struct {
		Header struct {
			Topic  string `json:"Topic"`
			Origin string `json:"Origin"`
		}
		Body struct {
			FormatId string `json:"FormatId"`
			Columns  []struct {
				Name string `json:"Name"`
				Type string `json:"Type"`
			} `json:"Columns"`
			Data []struct {
				T string   `json:"T"`
				D []string `json:"D"`
			} `json:"Data"`
		} `json:"Body"`
	} `json:"Payload"`
}

type eventMessageType struct {
	Event    string `json:"event"`
	Time     string `json:"time"`
	Entity   string `json:"entity"`
	EntityId string `json:"entityId"`
	Action   string `json:"action"`
	Data     struct {
		Device      string `json:"device"`
		Topic       string `json:"topic"`
		Origin      string `json:"origin"`
		Format      string `json:"format"`
		ReceiveTime string `json:"receiveTime"`
		Fields      map[string]struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"fields"`
	} `json:"data"`
}

type field struct {
	Key   string
	Value struct {
		Value string `json:"value"`
		Type  string `json:"type"`
	}
}

func getSortedFields(fields map[string]struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}) []field {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make([]field, 0, len(fields))
	for _, k := range keys {
		sorted = append(sorted, field{
			Key:   k,
			Value: fields[k],
		})
	}

	return sorted
}

func parseFile(jsonPath, outputPath string, callback ParserCallback) (int, error) {
	parsed := 0

	// Open JSON file
	file, err := os.Open(jsonPath)
	if err != nil {
		return parsed, fmt.Errorf("could not open JSON file: %v", err)
	}
	defer file.Close()

	// Read file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if outputPath != "" {
			if err := parseLineToCsv(line, outputPath); err != nil {
				return parsed, err
			}
		}
		if callback != nil {
			if err := parseLineToCallback(line, callback); err != nil {
				return parsed, err
			}
		}
		parsed++
	}
	if err := scanner.Err(); err != nil {
		return parsed, fmt.Errorf("error reading JSON file: %v", err)
	}

	return parsed, nil
}

type ParserCallback func(topic, origin, datetime string, data map[string]string)

func parseLineToCallback(line []byte, callback ParserCallback) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil {
		return fmt.Errorf("could not parse JSON: %v", err)
	}

	if _, ok := raw["event"]; ok {

		// eventMessageType

		var message eventMessageType
		if err := json.Unmarshal(line, &message); err != nil {
			return fmt.Errorf("failed to parse eventMessageType: %v", err)
		}

		message.Data.Topic, _ = strings.CutPrefix(message.Data.Topic, "data/")

		sortedFields := getSortedFields(message.Data.Fields)

		data := make(map[string]string, len(sortedFields))
		for _, field := range sortedFields {
			data[field.Key+":"+field.Value.Type] = field.Value.Value
		}
		messageTime, _ := time.Parse(time.RFC3339, message.Time)
		timestamp := messageTime.Format("2006-01-02 15:04:05")
		callback(message.Data.Topic, message.Data.Origin, timestamp, data)

	} else {

		// messageType

		var message messageType
		if err := json.Unmarshal(line, &message); err != nil {
			return fmt.Errorf("could not parse JSON: %v", err)
		}

		message.Payload.Header.Topic, _ = strings.CutPrefix(message.Payload.Header.Topic, "data/")

		for _, row := range message.Payload.Body.Data {
			timestamp, err := convertTimestamp(row.T)
			if err != nil {
				return err
			}
			data := make(map[string]string, len(message.Payload.Body.Columns))
			for i, col := range message.Payload.Body.Columns {
				data[col.Name+":"+col.Type] = row.D[i]
			}
			callback(message.Payload.Header.Topic, message.Payload.Header.Origin, timestamp, data)
		}
	}

	return nil
}

func parseLineToCsv(line []byte, outputPath string) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil {
		return fmt.Errorf("could not parse JSON: %v", err)
	}

	if _, ok := raw["event"]; ok {

		// eventMessageType

		var message eventMessageType
		if err := json.Unmarshal(line, &message); err != nil {
			return fmt.Errorf("failed to parse eventMessageType: %v", err)
		}

		message.Data.Topic, _ = strings.CutPrefix(message.Data.Topic, "data/")

		sortedFields := getSortedFields(message.Data.Fields)

		data := make(map[string]string, len(sortedFields))
		for _, field := range sortedFields {
			data[field.Key+":"+field.Value.Type] = field.Value.Value
		}
		messageTime, _ := time.Parse(time.RFC3339, message.Time)
		timestamp := messageTime.Format("2006-01-02 15:04:05")

		if message.Data.Format == "" {
			message.Data.Format = "signal"
		}

		//Construct CSV filename
		csvFilename := filepath.Join(outputPath, message.Data.Format+".csv")

		//Check if CSV file exists
		fileExists := true
		if _, err := os.Stat(csvFilename); os.IsNotExist(err) {
			fileExists = false
		}

		file, err := os.OpenFile(csvFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("could not open CSV file: %v", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		writer.Comma = ';'
		defer writer.Flush()

		// Write column names if file does not exist
		if !fileExists {
			var columnNames = []string{"topic", "origin", "datetime"}

			for _, field := range sortedFields {
				columnNames = append(columnNames, field.Key)
			}
			if err := writer.Write(columnNames); err != nil {
				return fmt.Errorf("could not write column names: %v", err)
			}
		}

		// Write data rows
		dataRow := []string{message.Data.Topic, message.Data.Origin, timestamp}
		for _, row := range sortedFields {
			dataRow = append(dataRow, row.Value.Value)
		}

		if err := writer.Write(dataRow); err != nil {
			return fmt.Errorf("could not write data row: %v", err)
		}

	} else {

		// messageType

		var message messageType
		if err := json.Unmarshal(line, &message); err != nil {
			return fmt.Errorf("could not parse JSON: %v", err)
		}

		message.Payload.Header.Topic, _ = strings.CutPrefix(message.Payload.Header.Topic, "data/")

		//Construct CSV filename
		csvFilename := filepath.Join(outputPath, message.Payload.Body.FormatId+".csv")

		//Check if CSV file exists
		fileExists := true
		if _, err := os.Stat(csvFilename); os.IsNotExist(err) {
			fileExists = false
		}

		file, err := os.OpenFile(csvFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("could not open CSV file: %v", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		writer.Comma = ';'
		defer writer.Flush()

		// Write column names if file does not exist
		if !fileExists {
			var columnNames = []string{"topic", "origin", "datetime"}
			for _, col := range message.Payload.Body.Columns {
				columnNames = append(columnNames, col.Name)
			}
			if err := writer.Write(columnNames); err != nil {
				return fmt.Errorf("could not write column names: %v", err)
			}
		}

		// Write data rows
		for _, row := range message.Payload.Body.Data {
			timestamp, err := convertTimestamp(row.T)
			if err != nil {
				return err
			}
			data := []string{message.Payload.Header.Topic, message.Payload.Header.Origin, timestamp}
			data = append(data, row.D...)
			if err := writer.Write(data); err != nil {
				return fmt.Errorf("could not write data row: %v", err)
			}
		}

	}
	return nil
}

func convertTimestamp(timestamp string) (string, error) {
	f, err := strconv.ParseFloat(timestamp, 64)
	if err != nil {
		return "", fmt.Errorf("could not parse float64: %v", err)
	}

	t := time.Unix(int64(f), 0)
	formattedTime := t.Format("2006-01-02 15:04:05")
	return formattedTime, nil
}
