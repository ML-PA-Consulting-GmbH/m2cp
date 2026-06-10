package virtual

import (
	"github.com/spf13/cobra"
	"m2cpcli/docker"
	"m2cpcli/format"
	"m2cpcli/virtual"
	"regexp"
	"strconv"
	"strings"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list virtual devices",
	RunE:  runListCmd,
}

func init() {
	VirtualDeviceCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

type VirtualDevice struct {
	Name           string `json:"name"`
	Store          string `json:"store"`
	Serial         string `json:"deviceSerial"`
	DebuggingPorts string `json:"debuggingPorts"` // TODO: split values from composite string?
	Running        bool   `json:"running"`
	Container      string `json:"container"` // TODO: split values from composite string?
	//Type           string `json:"type"`
	//Model          string `json:"model"`
}

type ListResult struct {
	TotalCount     int             `json:"totalCount"`
	VirtualDevices []VirtualDevice `json:"virtualDevices"`
}

func runListCmd(cmd *cobra.Command, args []string) error {
	var err error
	err = docker.IsDockerReady()
	if err != nil {
		return err
	}

	containers, err := virtual.GetVirtualDeviceContainers()
	if err != nil {
		return err
	}

	virtualDevices := make([]VirtualDevice, len(containers))
	for idx, container := range containers {
		virtualDevices[idx], err = reformatContainer(container)
		if err != nil {
			return err
		}
	}
	// TODO: sort? virtualDevices = virtualDevices.Sort("store:desc")

	result := ListResult{
		len(virtualDevices),
		virtualDevices,
	}

	return format.PrintFormattedOutput(cmd, result, customDeviceListFormatter)
}

func customDeviceListFormatter(res ListResult) (string, error) {
	table := format.NewTable(map[string]string{
		"1name":      "Alias",
		"2store":     "Store",
		"3serial":    "OS Serial Number",
		"4ports":     "Debugging Ports",
		"5running":   "Running",
		"6container": "Container Name",
	})
	for _, device := range res.VirtualDevices {
		table.AddRow(map[string]string{
			"1name":      device.Name,
			"2store":     device.Store,
			"3serial":    device.Serial,
			"4ports":     device.DebuggingPorts,
			"5running":   strconv.FormatBool(device.Running),
			"6container": device.Container,
		})
	}
	return table.String(), nil
}

// parseContainerName retrieves name, store and tenant-alias from the container name.
func parseContainerName(containerName string) (name, store string) {
	re := regexp.MustCompile("DEVICE_(.+)_STORE_(.+)")
	match := re.FindStringSubmatch(containerName)
	//v.Type = match[1]
	name = match[1]
	store = match[2]
	return name, store
}

func reformatContainer(container docker.ContainerInfo) (VirtualDevice, error) {
	v := VirtualDevice{
		Container: container.Name,
	}
	if strings.Contains(container.Status, "Up") {
		v.Running = true
	}
	if v.Running {
		v.Serial = virtual.RetrieveSerialOfDevice(container)
		//v.Model = docker.RetrieveModelOfDevice(container)
		v.DebuggingPorts = virtual.RetrieveDebuggingPorts(container)
	}
	v.Name, v.Store = parseContainerName(container.Name)
	return v, nil
}
