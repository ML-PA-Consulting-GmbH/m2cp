package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/tools"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/philippseith/signalr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var listenCmd = &cobra.Command{
	Use:   "listen <deviceSerial or deviceName>",
	Short: "Listen to websocket message traffic of a device",
	Args:  cobra.ExactArgs(1),
	RunE:  runListenCmd,
}

func init() {
	listenCmd.Flags().String("format", "", "output format (jsonl)")
	messageCmd.AddCommand(listenCmd)
}

func runListenCmd(cmd *cobra.Command, args []string) error {

	if args[0] == "" {
		return fmt.Errorf("device id, name or serial is required")
	}

	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	deviceInfo, err := backend.GetDeviceInfoById(cmd.Context(), deviceId)
	if err != nil {
		return err
	}

	if deviceInfo.UplinkMode == nil {
		return fmt.Errorf("device %s has unknown uplink mode", deviceId)
	}
	if *deviceInfo.UplinkMode != "SIGNALR_HUB" {
		return fmt.Errorf("device %s is currently not using websocket, but '%s' - live listening not possible",
			deviceId, tools.MaybeStringToString(deviceInfo.UplinkMode, "n/a"))
	}

	if deviceInfo.LastHubEndpoint == nil {
		return fmt.Errorf("device %s has unknown last hub endpoint", deviceId)
	}

	outputFormat, _ := cmd.Flags().GetString("format")
	if outputFormat != "" && outputFormat != "jsonl" {
		return fmt.Errorf("unsupported format %q, supported: jsonl", outputFormat)
	}

	jwtString := viper.GetString("jwt")

	return runWebsocketListener(cmd.Context(), deviceInfo.DeviceSerial, *deviceInfo.LastHubEndpoint, jwtString, outputFormat)
}

type messageReceiver struct {
	signalr.Hub
	deviceSerial string
	outputFormat string
}

func (r *messageReceiver) ReceiveEventMessage(message string) {
	if r.outputFormat == "jsonl" {
		// Compact the JSON and print as a single line
		var buf bytes.Buffer
		if err := json.Compact(&buf, []byte(message)); err != nil {
			// If compaction fails, print as-is (it's already a string)
			fmt.Println(message)
		} else {
			fmt.Println(buf.String())
		}
		return
	}

	// Parse the incoming message
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(message), &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Failed to parse message: %v\n", r.deviceSerial, err)
		fmt.Printf("Raw message: %s\n", message)
		return
	}

	// Check if this is a signalMessage or dataMessage
	entity, ok := parsed["entity"].(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "[%s] Unknown message format - no entity field\n", r.deviceSerial)
		prettyJSON, _ := json.MarshalIndent(parsed, "", "  ")
		fmt.Printf("Raw:\n%s\n", string(prettyJSON))
		return
	}

	timestamp, _ := parsed["time"].(string)

	switch entity {
	case "signalMessage":
		r.formatSignalMessage(parsed, timestamp)
	case "dataMessage":
		r.formatDataMessage(parsed, timestamp)
	default:
		fmt.Fprintf(os.Stderr, "[%s] Unknown entity type: %s\n", r.deviceSerial, entity)
		prettyJSON, _ := json.MarshalIndent(parsed, "", "  ")
		fmt.Printf("Raw:\n%s\n", string(prettyJSON))
	}
}

func printHeader(mType, origin, topic, timestamp string) {
	fmt.Printf("**[%s %s (%s) @ %s]**\n", mType, origin, topic, timestamp)
}

func (r *messageReceiver) formatSignalMessage(parsed map[string]interface{}, timestamp string) {
	data, ok := parsed["data"].(map[string]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "[%s] signalMessage missing data field\n", r.deviceSerial)
		return
	}

	origin, _ := data["origin"].(string)
	topic, _ := data["topic"].(string)

	printHeader("SIGNAL", origin, topic, timestamp)

	// Extract specific signal fields (name and content)
	fields := r.extractFields(data)
	if name, exists := fields["name"]; exists {
		fmt.Printf("name: %s\n", name.Value)
	}
	if content, exists := fields["content"]; exists {
		fmt.Print(r.formatContent(content.Value))
	}
	fmt.Println()

}

func (r *messageReceiver) formatDataMessage(parsed map[string]interface{}, timestamp string) {
	data, ok := parsed["data"].(map[string]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "[%s] dataMessage missing data field\n", r.deviceSerial)
		return
	}

	origin, _ := data["origin"].(string)
	topic, _ := data["topic"].(string)

	// Format header
	printHeader("DATA", origin, topic, timestamp)

	// Show all fields
	fields := r.extractFields(data)
	if len(fields) > 0 {
		for fieldName, fieldData := range fields {
			fmt.Printf("%s:%s = %s\n", fieldName, fieldData.Type, fieldData.Value)
		}
	}

	fmt.Println("")
}

type FieldData struct {
	Type  string
	Value string
}

func (r *messageReceiver) extractFields(data map[string]interface{}) map[string]FieldData {
	result := make(map[string]FieldData)

	fields, ok := data["fields"].(map[string]interface{})
	if !ok {
		return result
	}

	for fieldName, fieldData := range fields {
		if fieldMap, ok := fieldData.(map[string]interface{}); ok {
			fieldType, _ := fieldMap["type"].(string)
			fieldValue, _ := fieldMap["value"].(string)
			result[fieldName] = FieldData{
				Type:  fieldType,
				Value: fieldValue,
			}
		}
	}

	return result
}

func (r *messageReceiver) formatContent(content string) string {
	// Try to parse as JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(content), &jsonData); err != nil {
		// Not JSON, return as-is
		return content
	}

	// Convert to YAML with 2-character indentation
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(jsonData); err != nil {
		// Failed to convert to YAML, return original
		return content
	}
	encoder.Close()

	return buf.String()
}

// Simple logger to suppress debug noise
type quietLogger struct{}

func (l *quietLogger) Log(keyVals ...interface{}) error {
	// Only log errors, suppress debug/info
	if len(keyVals) >= 2 {
		if level, ok := keyVals[0].(string); ok && level == "level" {
			if value, ok := keyVals[1].(string); ok && value == "error" {
				fmt.Fprintf(os.Stderr, "SignalR Error: %v\n", keyVals)
			}
		}
	}
	return nil
}

func runWebsocketListener(ctx context.Context, deviceSerial string, hubUrl string, jwtString string, outputFormat string) error {
	if !strings.HasSuffix(hubUrl, "/userhub") {
		hubUrl = fmt.Sprintf("%s/userhub", hubUrl)
	}
	fmt.Fprintf(os.Stderr, "=== Starting SignalR connection ===\n")
	fmt.Fprintf(os.Stderr, "Hub endpoint: %s\n", hubUrl)
	fmt.Fprintf(os.Stderr, "Device serial: %s\n", deviceSerial)

	// Create a receiver for callbacks from the server
	fmt.Fprintln(os.Stderr, "Creating message receiver...")
	receiver := &messageReceiver{
		deviceSerial: deviceSerial,
		outputFormat: outputFormat,
	}

	// Create a Connection with Bearer token authentication
	fmt.Fprintln(os.Stderr, "Creating HTTP connection with Bearer token...")
	conn, err := signalr.NewHTTPConnection(ctx, hubUrl,
		signalr.WithHTTPHeaders(func() http.Header {
			header := http.Header{}
			header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtString))
			return header
		}))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create HTTP connection: %v\n", err)
		return err
	}
	fmt.Fprintln(os.Stderr, "HTTP connection created successfully")

	// Create the client and set a receiver for callbacks from the server
	fmt.Fprintln(os.Stderr, "Creating SignalR client...")
	client, err := signalr.NewClient(ctx,
		signalr.WithConnection(conn),
		signalr.WithReceiver(receiver),
		signalr.Logger(&quietLogger{}, false)) // Suppress debug logging
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create SignalR client: %v\n", err)
		return err
	}
	fmt.Fprintln(os.Stderr, "SignalR client created successfully")

	// Start the client loop
	fmt.Fprintln(os.Stderr, "Starting client loop...")
	client.Start()
	defer client.Stop()
	fmt.Fprintln(os.Stderr, "Client loop started")

	// Wait for connection to be established
	fmt.Fprintln(os.Stderr, "Waiting for client to connect...")
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	errCh := client.WaitForState(timeoutCtx, signalr.ClientConnected)
	select {
	case err = <-errCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
			return fmt.Errorf("failed to connect to device: %v", err)
		}
		fmt.Fprintln(os.Stderr, "Client connected successfully")
	case <-timeoutCtx.Done():
		fmt.Fprintln(os.Stderr, "Connection timeout")
		return fmt.Errorf("cancelled while waiting for connection")
	}

	// Subscribe to device messages
	fmt.Fprintf(os.Stderr, "Subscribing to device: %s\n", deviceSerial)
	ch := <-client.Invoke("Subscribe", deviceSerial)
	if ch.Error != nil {
		fmt.Fprintf(os.Stderr, "Subscribe failed: %v\n", ch.Error)
		return fmt.Errorf("failed to subscribe: %w", ch.Error)
	}
	fmt.Fprintln(os.Stderr, "Subscribed successfully")

	fmt.Fprintf(os.Stderr, "\nListening for messages from device: %s\n", deviceSerial)
	fmt.Fprintln(os.Stderr, "Press Ctrl+C to exit")

	// Wait for interrupt
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	fmt.Fprintln(os.Stderr, "\n=== Shutting down ===")

	// Unsubscribe
	fmt.Fprintln(os.Stderr, "Unsubscribing...")
	ch = <-client.Invoke("Unsubscribe", deviceSerial)
	if ch.Error != nil {
		fmt.Fprintf(os.Stderr, "Failed to unsubscribe: %v\n", ch.Error)
	} else {
		fmt.Fprintln(os.Stderr, "Unsubscribed successfully")
	}

	fmt.Fprintln(os.Stderr, "=== Connection closed ===")
	return nil
}
