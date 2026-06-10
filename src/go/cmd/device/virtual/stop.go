package virtual

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/docker"
	"m2cpcli/format"
	"m2cpcli/virtual"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop a virtual device",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runStopCmd,
}

func init() {
	VirtualDeviceCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// userCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//stopCmd.Flags().StringP("container", "c", "", "container name to stop")
}

func runStopCmd(cmd *cobra.Command, args []string) error {
	var err error
	err = docker.IsDockerReady()
	if err != nil {
		return err
	}

	var containerName string
	if len(args) > 0 {
		containerName = args[0]
	} else {
		containerName = virtual.GetContainerNameInteractively("Which virtual device do you want to stop?")
	}

	_, err = docker.ExecuteDockerCommand(fmt.Sprintf("stop %s", containerName))
	if err != nil {
		return fmt.Errorf("failed to stop virtual device container '%s': %s", containerName, err)
	}

	var result = VirtualDeviceMessageResult{
		Message: fmt.Sprintf("stopped virtual device '%s'", containerName),
	}

	return format.PrintFormattedOutput(cmd, result, customVirtualMessageFormatter)
}
