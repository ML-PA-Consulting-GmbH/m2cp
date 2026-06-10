package virtual

import (
	"bufio"
	"context"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/config"
	"m2cpcli/docker"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"m2cpcli/virtual"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

const (
	VirtualDeviceDefaultFleetDescription = "Automatically created fleet for virtual device"
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "create a virtual device",
	Long:  "create a new virtual device",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runCreateCmd,
}

func init() {
	VirtualDeviceCmd.AddCommand(createCmd)

	createCmd.Flags().Bool("coap", false, "expose the CoAP port of the virtual device. If multiple virtual devices have this option enabled, only one can be started at a time")
	createCmd.Flags().Bool("yes", false, "auto-approve")
	createCmd.Flags().String("volumes", "", "volumes to mount (syntax of docker volumes expected, e.g. /host/path:/container/path)")
	createCmd.Flags().Bool("fleet", true, "create a default fleet for the virtual device with its pre-installed apps")
	createCmd.Flags().StringP("auto-update-mode", "m", "stable", "auto update mode for the default fleet (off, stable, edge)")

	createCmd.Flags().Bool("no-pull", false, "do not pull docker image of virtual device (for development of virtual devices)")
}

func runCreateCmd(cmd *cobra.Command, args []string) error {
	var err error

	err = docker.IsDockerReady()
	if err != nil {
		return err
	}

	containerRegistryCredentials, err := gql.VirtualDeviceContainerRegistryCredentials(cmd.Context())
	if err != nil {
		return err
	}

	containerRegistryUri := strings.TrimPrefix(containerRegistryCredentials.ContainerRegistryUri, "https://")
	containerRegistryUsername := containerRegistryCredentials.Username
	containerRegistryPassword := containerRegistryCredentials.Password
	err = loginDockerRegistry(containerRegistryUri, containerRegistryUsername, containerRegistryPassword)
	if err != nil {
		return err
	}

	storeUrl := viper.GetString("store")

	storeAlias := config.StoreToAlias(storeUrl)
	if storeAlias == "unknown" {
		return fmt.Errorf("unknown store: %s", storeUrl)
	}

	store := config.Store{
		Alias:       config.StoreToAlias(storeUrl),
		URL:         storeUrl,
		DockerImage: config.DockerImage(containerRegistryUri, storeAlias),
	}

	var autoApprove bool
	autoApprove, err = cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	var inputDeviceName string
	var volumes string
	var createFleet bool
	fleetAutoUpdateMode := "stable"

	var exposeCoapPort bool
	if len(args) > 0 {
		inputDeviceName = args[0]
	}

	// Set auto update mode if provided via flag
	if cmd.Flags().Changed("auto-update-mode") {
		changedAutoUpdateMode, err := cmd.Flags().GetString("auto-update-mode")
		if err != nil {
			return err
		}
		// Validate that the auto update mode exists
		if _, found := backend.GetAutoUpdateModeIdByName(changedAutoUpdateMode); !found {
			return fmt.Errorf("unknown auto update mode '%s'", changedAutoUpdateMode)
		}
		fleetAutoUpdateMode = changedAutoUpdateMode
	}

	// Set volumes if provided via flag
	if cmd.Flags().Changed("volumes") {
		volumes, err = cmd.Flags().GetString("volumes")
		if err != nil {
			return err
		}
	}

	if autoApprove {
		if inputDeviceName == "" {
			return fmt.Errorf("name is required when auto-approving")
		}
		if cmd.Flags().Changed("coap") {
			exposeCoapPort = true
		}
		if cmd.Flags().Changed("fleet") {
			createFleet = true
		}
	} else {
		inputDeviceName, err = interactivelyPromptForName()
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("coap") {
			exposeCoapPort = true
		} else {
			exposeCoapPort = interactivelyPromptForCoapExpose()
		}
		if cmd.Flags().Changed("fleet") {
			createFleet = true
		} else {
			createFleet = interactivelyPromptForFleet()
		}

		err = interactivelyConfirm(inputDeviceName, volumes, exposeCoapPort, createFleet, fleetAutoUpdateMode, store)
		if err != nil {
			return err
		}
	}
	_, err = docker.IsValidDockerName(inputDeviceName)
	if err != nil {
		return err
	}

	deviceNameCanBeUsed, err := deviceNameCanBeUsed(cmd, inputDeviceName)
	if err != nil {
		return err
	}
	if !deviceNameCanBeUsed {
		return fmt.Errorf("device name '%s' is in use", inputDeviceName)
	}

	if createFleet {
		_, err := isFleetNameUsable(cmd.Context(), inputDeviceName)
		if err != nil {
			return fmt.Errorf("cannot create default fleet, as a a fleet with the name '%s' already exists. Either delete the fleet or choose another device name", inputDeviceName)
		}
	}

	dockerContainers, err := docker.GetContainers()
	if err != nil {
		return err
	}
	portResult, err := getFreeServicePorts(dockerContainers)
	if err != nil {
		return fmt.Errorf("failed to get free ports")
	}

	if exposeCoapPort {
		portResult.CoapPort = new(int)
		*portResult.CoapPort = 5683

		runningContainersRequiringCoap, err := virtual.GetRunningContainersRequiringCoapPort(dockerContainers)
		if err != nil {
			return fmt.Errorf("failed to get containers requiring CoAP port: %s", err)
		}
		if len(runningContainersRequiringCoap) > 0 {
			return fmt.Errorf("cannot create virtual device '%s' because the CoAP port %s is in use by the following running container: %s",
				inputDeviceName, config.CoapExpectedPort, strings.Join(virtual.ContainerInfoListToContainerNameList(runningContainersRequiringCoap), ", "))
		}
	}

	var containerName string
	noPull := cmd.Flags().Changed("no-pull")
	containerName, err = createVirtualDeviceContainer(inputDeviceName, volumes, noPull, store, portResult)
	if err != nil {
		return err
	}

	err = pollUntilReady(containerName)
	if err != nil {
		return err
	}

	//user, err := gql.UserInfo(cmd.Context())
	//if err != nil {
	//	return err
	//}

	// Add current user to the container to make SSH access possible
	//err = virtual.AddUserToContainer(containerName, *user)
	//if err != nil {
	//	return fmt.Errorf("could not add user to container: %s", err)
	//}

	// Update device name in backend
	deviceSerial, err := virtual.GetDeviceSerialFromContainer(containerName)
	if err != nil {
		return err
	}
	deviceId, err := gql.DeviceIdByNameOrSerial(cmd.Context(), deviceSerial)
	if err != nil {
		return err
	}
	_, err = gql.DeviceModify(cmd.Context(), deviceId, inputDeviceName, "", nil)
	if err != nil {
		return err
	}
	deviceModel, err := virtual.RetrieveModelNameOfDevice(containerName)
	if err != nil {
		return err
	}
	deviceModelRevision, err := virtual.RetrieveModelRevisionOfDevice(containerName)
	if err != nil {
		return err
	}

	// Create fleet if requested
	err = createDefaultFleet(cmd, deviceModel, deviceModelRevision, createFleet, fleetAutoUpdateMode, containerName, inputDeviceName, deviceId)
	if err != nil {
		return err
	}

	return nil
}

func createDefaultFleet(cmd *cobra.Command, deviceModel string, deviceModelRevision int, createFleet bool, autoUpdateMode string,
	containerName string, inputDeviceName string, deviceId gql.UUID) error {
	if !createFleet {
		return nil
	}

	installedSnaps, err := getInstalledSnaps(containerName)
	if err != nil {
		return fmt.Errorf("could not determine snaps installed in the virtual device: %s", err)
	}

	createdFleet, err := backend.CreateFleetWithAutoUpdate(cmd.Context(), inputDeviceName, VirtualDeviceDefaultFleetDescription,
		false, "m2cp", deviceModel, deviceModelRevision, autoUpdateMode)
	if err != nil {
		return fmt.Errorf("could not create fleet for virtual device: %s", err)
	}
	fleetInfo, err := backend.GetFleetInfoById(cmd.Context(), createdFleet.Id)
	if err != nil {
		return fmt.Errorf("could not get fleet info: %s", err)
	}

	// Check, if any snap is added by default to the fleet
	var preAddedSnapsInFleet []string
	for _, fbs := range fleetInfo.Fleet.FleetBridgeSnapRevisions {
		preAddedSnapsInFleet = append(preAddedSnapsInFleet, fbs.SnapRevision.SnapDeclaration.SnapName)
	}

	for snapName, snapRevision := range installedSnaps {
		if slices.Contains(preAddedSnapsInFleet, snapName) {
			continue
		}
		snapRevision, err := backend.GetSnapRevisionByNameArchRevision(cmd.Context(), snapName, "amd64", snapRevision)
		if err != nil {
			return fmt.Errorf("could not get snap revision of snap '%s' for '%s': %s", snapName, "amd64", err)
		}

		_, err = backend.CreateFleetBridgeSnapRevisions(cmd.Context(), createdFleet.Id, snapRevision.Id)
		if err != nil {
			return fmt.Errorf("could not add snap '%s' to fleet: %s", snapName, err)
		}
	}

	// Now modify the automatically added snaps of fleet to have the correct revision.
	// But only, if auto update mode is off, as otherwise the fleet bridge snap revisions are set to the most recent stable/edge version
	if autoUpdateMode == "off" {
		for snapName, snapRevision := range installedSnaps {
			if !slices.Contains(preAddedSnapsInFleet, snapName) {
				continue
			}
			// Find the fleet bridge snap revision id we have to modify
			var fleetBridgeSnapRevisionId string
			for _, fbs := range fleetInfo.Fleet.FleetBridgeSnapRevisions {
				if fbs.SnapRevision.SnapDeclaration.SnapName == snapName {
					fleetBridgeSnapRevisionId = fbs.Id
					break
				}
			}

			// Get the installed snap revision and update the fleet bridge snap revision
			installedSnapRevision, err := backend.GetSnapRevisionByNameArchRevision(cmd.Context(), snapName, "amd64", snapRevision)
			if err != nil {
				return fmt.Errorf("could not get snap revision of snap '%s' for '%s': %s", snapName, "amd64", err)
			}
			_, err = backend.UpdateFleetBridgeSnapRevision(cmd.Context(), fleetBridgeSnapRevisionId, installedSnapRevision.Id)
			if err != nil {
				return err
			}
		}
	}

	// Add virtual device to the fleet
	_, err = backend.SetDeviceFleet(cmd.Context(), string(deviceId), &createdFleet.Id)
	if err != nil {
		return fmt.Errorf("could not add virtual device to fleet: %s", err)
	}
	return nil
}

func loginDockerRegistry(containerRegistryUri string, user string, passwd string) error {
	output, err := docker.ExecuteDockerCommand(fmt.Sprintf("login --username %s --password %s %s",
		user, passwd, containerRegistryUri))
	if err != nil {
		return fmt.Errorf("failed to login user '%s' to container registry '%s'':\n%s\n%s",
			user, containerRegistryUri, output, err)
	}
	return nil
}

func interactivelyPromptForName() (name string, err error) {
	fmt.Print("Give the device a name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name = scanner.Text()
	_, err = docker.IsValidDockerName(name)
	if err != nil {
		return "", err
	}
	return name, nil
}

func interactivelyPromptForVolumes() (volumes string, err error) {
	fmt.Print("Give volumes to mount in the device (syntax of docker volumes expected, e.g. /host/path:/container/path)" +
		" Leave blank if no volumes required: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	volumes = scanner.Text()
	return volumes, nil
}

func interactivelyPromptForFleet() (createFleet bool) {
	fmt.Print("Create a default fleet for the virtual device with its pre-installed apps? (Y/n): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := scanner.Text()
	if response == "n" {
		return false
	} else {
		return true
	}
}

func interactivelyPromptForCoapExpose() (coapExpose bool) {
	fmt.Print("Expose the CoAP port of the virtual device? If multiple virtual devices have this option enabled, only one can be started at a time.\nLeave blank to not expose CoAP (y/N): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := scanner.Text()
	if response == "y" {
		return true
	} else {
		return false
	}
}

func interactivelyConfirm(name, volumes string, exposeCoap, createDefaultFleet bool,
	autoUpdateMode string, store config.Store) error {
	var volumesMessage string
	if volumes == "" {
		volumesMessage = "none"
	} else {
		volumesMessage = volumes
	}

	fmt.Println("You are about to create the following virtual device:")
	fmt.Println(fmt.Sprintf("	Alias: %s", name))
	fmt.Println(fmt.Sprintf("	Store: %s (%s)", store.URL, store.Alias))
	fmt.Println(fmt.Sprintf("	Expose CoAP port: %t", exposeCoap))
	fmt.Println(fmt.Sprintf("	Volumes: %s", volumesMessage))
	fmt.Println(fmt.Sprintf("	Create fleet: %t", createDefaultFleet))
	if createDefaultFleet {
		fmt.Println(fmt.Sprintf("	Fleet auto update mode: %s", autoUpdateMode))
	}
	fmt.Println("Are you sure you want to continue? (y/N)")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return err
	}
	if (response != "y") && (response != "Y") {
		return nil
	}
	return nil
}

// deviceNameCanBeUsed checks, if a device of the same name already exists
func deviceNameCanBeUsed(cmd *cobra.Command, name string) (bool, error) {
	_, err := gql.DeviceIdByName(cmd.Context(), name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

type PortResult struct {
	SnapdPort    int
	RabbitMqPort int
	SshPort      int
	CoapPort     *int // Optional
}

func getFreeServicePorts(dockerContainers []docker.ContainerInfo) (*PortResult, error) {
	const MinBasePort = 20000
	const MaxBasePort = 65532
	usedPorts, err := tools.GetUsedPortsOs()
	if err != nil {
		return nil, fmt.Errorf("failed to get used ports of operating system")
	}
	// Get used ports of containers
	for _, container := range dockerContainers {
		for _, port := range container.PortsList {
			usedPorts = append(usedPorts, port.HostPort)
		}
	}

	for i := MinBasePort; i < MaxBasePort; i += 4 {
		if !slices.Contains(usedPorts, i) && !slices.Contains(usedPorts, i+1) && !slices.Contains(usedPorts, i+2) {
			res := PortResult{
				SnapdPort:    i,
				RabbitMqPort: i + 1,
				SshPort:      i + 2,
			}
			return &res, nil
		}
	}
	return nil, fmt.Errorf("no free ports found")
}

func createVirtualDeviceContainer(name string, volumes string, noPull bool, store config.Store, portResult *PortResult) (string, error) {
	containerName := fmt.Sprintf("M2CP_VIRTUAL_DEVICE_%s_STORE_%s", name, store.Alias)

	containers, err := docker.GetContainers()
	if err != nil {
		return containerName, err
	}

	for _, container := range containers {
		if container.Name == containerName {
			return containerName, fmt.Errorf("a virtual device container for device name '%s' already exists for store '%s'' with name '%s'", name, store.Alias, containerName)
		}
	}

	// Start container
	extraVars := ""
	if store.Alias == "local" {
		// We allow access to the host from the container for local deployments as we assume in a local deployment
		// all backend services to be running on the host (for development and debugging purposes)
		extraVars = "--add-host=host.docker.internal:host-gateway"
	}

	ports := []string{
		fmt.Sprintf("%d:%s", portResult.SnapdPort, config.SnapdExpectedPort),
		fmt.Sprintf("%d:%s", portResult.RabbitMqPort, config.RabbitMqExpectedPort),
		fmt.Sprintf("%d:%s", portResult.SshPort, config.SshExpectedPort),
	}
	if portResult.CoapPort != nil {
		ports = append(ports, fmt.Sprintf("%d:%s/udp", *portResult.CoapPort, config.CoapExpectedPort))
	}

	fmt.Printf("Creating device '%s' as container '%s' \n", name, containerName)
	err = docker.CreateDockerContainer(containerName, name, store.DockerImage, volumes, ports, noPull, extraVars)
	if err != nil {
		return containerName, err
	}

	return containerName, nil
}

func parseStates(stdout string, states map[string]bool) {
	lines := strings.Split(stdout, "\n")
	for _, line := range lines[1:] {
		columns := strings.Fields(line)
		if len(columns) > 2 {
			service := columns[0]
			state := columns[2]
			if state == "active" {
				states[service] = true
			}
		}
	}
}

func pollUntilReady(containerName string) error {
	// Idea: Attach to the container, and look at the "Current" column of `snap services`.
	// m2cp-gateway and m2cp-message-hub must be `active`.
	fmt.Print("Waiting for virtual device to be ready.")
	defer fmt.Println("")

	emptyResponseRetries := 0
	states := make(map[string]bool)
	for {
		stdout, err := docker.ExecuteCommandInContainer(containerName, "snap services")
		if err != nil {
			// Snap is not always directly ready, and it can take some time depending on the speed of the system. So we give
			// it some time to start up.
			if emptyResponseRetries < 5 {
				time.Sleep(2 * time.Second)
				emptyResponseRetries++
				continue
			}
			return err
		}

		parseStates(stdout, states)
		// states["m2cp-gateway.gateway-daemon"] is still used by local virtual devices as they still use C# gateway. Can be removed when local virtual devices get updated
		if (states["m2cp-gateway.gateway-daemon"] || states["m2cp-gateway.m2cp-gateway"] || states["m2cp-gateway.uplink-connection"]) && (states["m2cp-message-hub.rabbitmq-server"] || states["m2cp-message-hub.lavinmq"]) {
			return nil
		}

		time.Sleep(1 * time.Second)
		fmt.Print(".")
	}
}

func getInstalledSnaps(containerName string) (map[string]int, error) {
	stdout, err := docker.ExecuteCommandInContainer(containerName, "snap list")
	if err != nil {
		return nil, err
	}

	// Get column Name and Revision
	lines := strings.Split(stdout, "\n")
	snapMap := make(map[string]int)
	for _, line := range lines[1:] {
		columns := strings.Fields(line)
		if len(columns) > 2 {
			revision, err := strconv.Atoi(columns[2])
			if err != nil {
				return nil, fmt.Errorf("could not parse revision: %s", columns[2])
			}
			snapMap[columns[0]] = revision
		}
	}
	return snapMap, nil
}

func isFleetNameUsable(ctx context.Context, fleetName string) (bool, error) {
	_, err := gql.FleetIdByName(ctx, fleetName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
