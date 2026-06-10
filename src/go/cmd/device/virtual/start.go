package virtual

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/config"
	"m2cpcli/docker"
	"m2cpcli/format"
	"m2cpcli/tools"
	"m2cpcli/virtual"
	"strings"
)

var startCmd = &cobra.Command{
	Use:   "start [containerName]",
	Short: "start a (non-running) virtual device",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runStartCmd,
}

func init() {
	VirtualDeviceCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// userCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//startCmd.Flags().StringP("container", "c", "", "container name to start")
}

type VirtualDeviceMessageResult struct {
	Message string `json:"message"`
}

func runStartCmd(cmd *cobra.Command, args []string) error {
	var err error
	err = docker.IsDockerReady()
	if err != nil {
		return err
	}

	var targetContainerName string
	if len(args) > 0 {
		targetContainerName = args[0]
	} else {
		targetContainerName = virtual.GetContainerNameInteractively("Which virtual device do you want to start?")
	}

	dockerContainers, err := docker.GetContainers()
	if err != nil {
		return fmt.Errorf("failed to get list of running containers: %s", err)
	}
	if virtual.IsContainerRunning(dockerContainers, targetContainerName) {
		return fmt.Errorf("virtual device '%s' is already running. To restart, stop the virtual device first", targetContainerName)
	}

	runningContainersRequiringCoap, err := virtual.GetRunningContainersRequiringCoapPort(dockerContainers)
	if err != nil {
		return fmt.Errorf("failed to get containers requiring CoAP port: %s", err)
	}
	targetContainerRequiresCoapPort := virtual.IsContainerRequiringCoapPort(dockerContainers, targetContainerName)
	_, runningContainersUsingCoap := virtual.IsCoapPortInUseByAnotherVirtualDevice(runningContainersRequiringCoap, targetContainerName)

	if targetContainerRequiresCoapPort {
		if len(runningContainersUsingCoap) > 0 {
			return fmt.Errorf("cannot start virtual device '%s' because the CoAP port %s is already in use by the following running container: %s",
				targetContainerName, config.CoapExpectedPort, strings.Join(runningContainersUsingCoap, ", "))
		}

		if tools.IsUdpPortInUse(config.CoapExpectedPort) {
			return fmt.Errorf("cannot start virtual device '%s' because the CoAP port %s is already in use by another process",
				targetContainerName, config.CoapExpectedPort)
		}
	}

	_, err = docker.ExecuteDockerCommand(fmt.Sprintf("start %s", targetContainerName))
	if err != nil {
		return fmt.Errorf("failed to start virtual device container '%s': %s", targetContainerName, err)
	}

	var result = VirtualDeviceMessageResult{
		Message: fmt.Sprintf("started virtual device '%s'", targetContainerName),
	}

	return format.PrintFormattedOutput(cmd, result, customVirtualMessageFormatter)
}

func customVirtualMessageFormatter(res VirtualDeviceMessageResult) (string, error) {
	return res.Message, nil
}
