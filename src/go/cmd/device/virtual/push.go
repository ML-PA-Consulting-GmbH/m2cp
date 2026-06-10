package virtual

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/docker"
	"m2cpcli/format"
	"m2cpcli/virtual"
	"os"
	"path/filepath"
	"strings"
)

var pushCmd = &cobra.Command{
	Use:   "push [filePath]",
	Short: "install a snap directly on a virtual device",
	Args:  validateSnapPushArgs,
	RunE:  runSnapPushCmd,
}

func init() {
	VirtualDeviceCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringP("containerName", "c", "", "container name to push snap to")
}

func validateSnapPushArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("push requires a snap file path")
	}

	filePath := args[0]
	err := isSnapFilePath(filePath)
	if err != nil {
		return err
	}
	return nil
}

func isSnapFilePath(filepath string) error {
	if _, err := os.Stat(filepath); err != nil {
		return fmt.Errorf("file '%s' does not exist", filepath)
	}
	if !strings.HasSuffix(filepath, ".snap") {
		return fmt.Errorf("file '%s' should end with '.snap'", filepath)
	}
	return nil
}

func runSnapPushCmd(cmd *cobra.Command, args []string) error {
	var err error

	err = docker.IsDockerReady()
	if err != nil {
		return err
	}

	containerName, err := cmd.Flags().GetString("containerName")
	if containerName == "" {
		containerName = virtual.GetContainerNameInteractively("Which virtual device do you want to push to?")
	}

	snapFilePath := args[0]
	err = isSnapFilePath(snapFilePath)
	if err != nil {
		return err
	}

	snapFileName := filepath.Base(snapFilePath)
	destinationFilePath := fmt.Sprintf("/tmp/%s", snapFileName)

	// copy snap into docker
	_, err = docker.ExecuteDockerCommand(fmt.Sprintf("cp %s %s:%s",
		snapFilePath, containerName, destinationFilePath))
	if err != nil {
		return fmt.Errorf("failed to copy snap '%s' to container '%s': %s", snapFileName, containerName, err.Error())
	}

	// install snap within container
	_, err = docker.ExecuteDockerCommand(fmt.Sprintf("exec %s snap install %s --devmode",
		containerName, destinationFilePath))
	if err != nil {
		if strings.HasPrefix(err.Error(), "Warning") {
			cmd.Println(err.Error())
		} else {
			return fmt.Errorf("failed installing snap '%s': %s", snapFileName, err)
		}
	}

	// remove the snap file from the docker container
	_, err = docker.ExecuteDockerCommand(fmt.Sprintf("exec %s rm %s", containerName, destinationFilePath))
	if err != nil {
		return fmt.Errorf("failed remove '%s' in container '%s': %s", snapFileName, containerName, err.Error())
	}

	var result = VirtualDeviceMessageResult{
		Message: fmt.Sprintf("installed snap '%s' on virtual device '%s'", snapFileName, containerName),
	}
	return format.PrintFormattedOutput(cmd, result, customVirtualMessageFormatter)
}
