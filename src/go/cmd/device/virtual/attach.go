package virtual

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/docker"
	"m2cpcli/virtual"
)

var attachCmd = &cobra.Command{
	Use:   "attach [containerName]", //  or deviceSerial
	Short: "attach to a virtual device as an interactive shell",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runAttachCmd,
}

func init() {
	VirtualDeviceCmd.AddCommand(attachCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// userCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runAttachCmd(cmd *cobra.Command, args []string) error {
	var err error
	err = docker.IsDockerReady()
	if err != nil {
		return err
	}

	var containerName string
	if len(args) > 0 {
		containerName = args[0]
	} else {
		containerName = virtual.GetContainerNameInteractively("Which virtual device do you want attach to?")
	}

	cmd.Println(fmt.Sprintf("Attaching to virtual device '%s' ...", containerName))
	docker.AttachToContainer(containerName)
	cmd.Println("Detached from virtual device")

	return nil
}
