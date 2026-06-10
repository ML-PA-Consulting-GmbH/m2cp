package ssh

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var openCmd = &cobra.Command{
	Use:   "open <deviceName|deviceSerial>",
	Short: "Establish reverse ssh tunnel to enable ssh connection to device",
	RunE:  runOpenCmd,
	Args:  cobra.ExactArgs(1),
}

func init() {
	sshCmd.AddCommand(openCmd)
	openCmd.Flags().BoolP("reset", "r", false, "Force a reset of the ssh connection")
}

type SshCommands struct {
	Open string `json:"open"`
}

type DeviceSshOpenResult struct {
	Message string      `json:"message"`
	Command SshCommands `json:"command"`
	TaskId  string      `json:"task-id"`
}

func runOpenCmd(cmd *cobra.Command, args []string) error {
	storeId := viper.GetString("store-id")
	if storeId == "" {
		return fmt.Errorf("no store id found in config")
	}
	sshKeyFilePath := viper.GetString(storeId + ".ssh-key")
	if sshKeyFilePath == "" {
		return fmt.Errorf("no ssh key found in config of store id %s", storeId)
	}

	device, err := getDevice(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	containerStatus, err := backend.GetEdgeDeviceSshContainerStatus(cmd.Context(), device.DeviceSerial)
	if err != nil {
		return err
	}

	if containerStatus.EdgeDeviceSshStatus != nil && !containerStatus.EdgeDeviceSshStatus.IsReady {
		fmt.Println("The SSH container is not ready. The connection setup may take a few moments. Please wait...")
	}

	// Pausing execution for 1 second due to reverse SSH relay server's policy of 1 request per IP per second
	time.Sleep(1 * time.Second)

	resetSshConnection := cmd.Flags().Changed("reset")
	var sshOpenResponse *gql.EdgeDeviceSshOpenOutput

	for i := 1; i < 120; i++ {
		sshOpenResponse, err = gql.EdgeDeviceSshOpen(cmd.Context(), device.DeviceSerial, resetSshConnection)
		if err != nil {
			if strings.Contains(err.Error(), "Resource manager is currently working on creating") {
				fmt.Printf(".")
				time.Sleep(2 * time.Second)
				continue
			}
			fmt.Printf("SSH connecting not ready, yet... (%s)\n", err)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return fmt.Errorf("too many failed attempts - giving up")
	}

	sshConfig := makeSshConfig(sshKeyFilePath, device.DeviceSerial, sshOpenResponse)
	sshConfigDir, err := getSshConfigDir()
	if err != nil {
		return err
	}

	err = createDirIfNeededAndWriteFile(sshConfigDir, device.DeviceSerial, sshConfig)
	if err != nil {
		return err
	}

	msg := DeviceSshOpenResult{
		Message: "Reverse SSH tunnel opened successfully",
		Command: SshCommands{
			Open: fmt.Sprintf("ssh -F %s %s", filepath.Join(sshConfigDir, device.DeviceSerial), device.DeviceSerial),
		},
		TaskId: gql.TaskId,
	}
	return format.PrintFormattedOutput(cmd, msg, customSshOpenFormatter)
}

func customSshOpenFormatter(res DeviceSshOpenResult) (string, error) {
	// TODO: What about a flag to connect directly like 'virtual-device attach'?
	msg := ""
	msg += "command to connect to device:\n\n"
	msg += fmt.Sprintf("  %s\n\n", res.Command.Open)

	//list := format.NewList()
	//list.Add("round trip time", fmt.Sprintf("%s", res.RoundTrip.Round(time.Millisecond)))
	//list.Add("uptime", fmt.Sprintf("%s", res.Uptime.Round(time.Second))) // TODO: more human-readable format?
	//list.Add("task ID", fmt.Sprintf("%s", res.TaskId))

	return msg, nil
}

func createDirIfNeededAndWriteFile(configDirPath, filename, contents string) error {
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		err2 := os.Mkdir(configDirPath, os.ModePerm)
		if err2 != nil {
			return err2
		}
	}

	filePath := path.Join(configDirPath, filename)
	err := os.WriteFile(filePath, []byte(contents), 0644)
	if err != nil {
		return fmt.Errorf("failed to write \"%s\": %s", filePath, err)
	}

	return nil
}

func getSshConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(home, ".m2cp", "ssh"), nil
}

func makeSshConfig(privateKeyFilePath string, deviceIdentifier string, openOutput *gql.EdgeDeviceSshOpenOutput) string {
	sshConfigTemplate := `Host m2cp_jump
	Hostname %s
	Port %s
	User %s
	IdentityFile %s
	StrictHostKeyChecking no
	UserKnownHostsFile /dev/null
	ServerAliveInterval 30
	ServerAliveCountMax 3

Host %s
	Hostname localhost
	Port %s
	User %s
	IdentityFile %s
	ProxyJump m2cp_jump
	StrictHostKeyChecking no
	UserKnownHostsFile /dev/null
	ServerAliveInterval 30
	ServerAliveCountMax 3
`
	return fmt.Sprintf(sshConfigTemplate, openOutput.Host, openOutput.RemoteSshPort, openOutput.UsernameServer,
		privateKeyFilePath, deviceIdentifier, openOutput.ReverseSshPort, openOutput.UsernameDevice, privateKeyFilePath)
}

func getDevice(ctx context.Context, nameOrSerial string) (*gql.EdgeDevice, error) {
	var deviceId gql.UUID
	var err error

	if tools.IsValidUuid(nameOrSerial) {
		deviceId, err = gql.DeviceIdBySerial(ctx, nameOrSerial)
		if err != nil {
			return nil, err
		}
	} else {
		deviceId, err = gql.DeviceIdByName(ctx, nameOrSerial)
		if err != nil {
			return nil, err
		}
	}

	device, err := gql.EdgeDeviceByDeviceId(ctx, deviceId)
	if err != nil {
		return nil, err
	}
	return device, nil
}
