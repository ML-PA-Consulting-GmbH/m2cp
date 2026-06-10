package virtual

import (
	"bufio"
	"fmt"
	"m2cpcli/config"
	"m2cpcli/docker"
	gql "m2cpcli/graphql"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func GetVirtualDeviceContainers() ([]docker.ContainerInfo, error) {
	containers, err := docker.GetContainers()
	if err != nil {
		return nil, err
	}

	pattern := "^M2CP_VIRTUAL_DEVICE_[^_]+_STORE_[^_]+$"
	regex := regexp.MustCompile(pattern)

	var virtualDevices []docker.ContainerInfo
	for _, container := range containers {
		// Only match containers named like virtual devices
		if !regex.MatchString(container.Name) {
			continue
		}
		virtualDevices = append(virtualDevices, container)
	}
	return virtualDevices, nil
}

// TODO: when attaching, list only running devices?
// TODO: when there is only on, use that?
func GetContainerNameInteractively(question string) string {
	virtualDeviceContainers, err := GetVirtualDeviceContainers()
	if err != nil {
		return ""
	}
	if len(virtualDeviceContainers) == 0 {
		fmt.Println("There are no virtual devices. Either there are none or they are already running")
		return ""
	}

	result := ""
	// Print the list of options to the user
	fmt.Println(question)
	for i, container := range virtualDeviceContainers {
		fmt.Printf("[%d] %s\n", i+1, container.Name)
	}

	// Prompt the user to enter a choice
	fmt.Print("Enter the number of the device: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return result
	}
	if strings.Contains(input, "\r") {
		input = input[:len(input)-1]
	}

	// Convert the input to an integer and validate the range
	choice, err := strconv.Atoi(input[:len(input)-1]) // remove newline character
	if err != nil || choice < 1 || choice > len(virtualDeviceContainers) {
		return result
	}
	return virtualDeviceContainers[choice-1].Name
}

func RetrieveModelNameOfDevice(containerName string) (string, error) {
	cmdResult, err := docker.ExecuteCommandInContainer(containerName, "snap known model | grep model:")
	if err != nil {
		return "", err
	}
	cmdResult = strings.ReplaceAll(cmdResult, "model: ", "")
	cmdResult = strings.ReplaceAll(cmdResult, "\n", "")
	return cmdResult, nil
}

func RetrieveModelRevisionOfDevice(containerName string) (int, error) {
	cmdResult, err := docker.ExecuteCommandInContainer(containerName, "snap known model | grep revision:")
	if err != nil {
		return -1, err
	}
	cmdResult = strings.ReplaceAll(cmdResult, "revision: ", "")
	cmdResult = strings.ReplaceAll(cmdResult, "\n", "")
	cmdResult = strings.Trim(cmdResult, " ")
	if cmdResult == "" {
		return 1, nil
	}
	revision, err := strconv.Atoi(cmdResult)
	return revision, err
}

func RetrieveSerialOfDevice(container docker.ContainerInfo) string {
	cmdResult, err := docker.ExecuteCommandInContainer(container.Name, "snap known serial | grep serial:")
	if err != nil {
		return cmdResult
	}
	cmdResult = strings.ReplaceAll(cmdResult, "serial: ", "")
	cmdResult = strings.ReplaceAll(cmdResult, "\n", "")
	return cmdResult
}

func GetDeviceSerialFromContainer(containerName string) (string, error) {
	var serial = ""
	var err error

	containers, err := GetVirtualDeviceContainers()
	if err != nil {
		return serial, err
	}

	for _, container := range containers {
		if container.Name == containerName {
			if strings.Contains(container.Status, "Up") {
				serial = RetrieveSerialOfDevice(container)
				return serial, nil
			}
		}
	}

	return serial, fmt.Errorf("could not find container '%s'", containerName)
}

func RetrieveDebuggingPorts(container docker.ContainerInfo) string {
	var result string

	output, err := docker.ExecuteDockerCommand("port " + container.Name)
	if err != nil {
		return result
	}

	// The ports in the container are always the same
	// We assume, that IPv4 will always be available, in contrast to IPv6, which is not available in  Windows Subsystem for Linux (WSL)
	pattern := fmt.Sprintf("(%s|%s|%s)/tcp -> 0.0.0.0[\\[\\]0-9a-fA-F.:]*:(\\d+)",
		config.SnapdExpectedPort, config.RabbitMqExpectedPort, config.SshExpectedPort)
	regex := regexp.MustCompile(pattern)
	match := regex.FindAllStringSubmatch(output, -1)
	if match == nil || len(match) < 3 {
		return result
	}

	rabbitmqPort := ""
	snapdPort := ""
	sshPort := ""
	for _, m := range match {
		if m[1] == config.SnapdExpectedPort {
			snapdPort = m[2]
		}
		if m[1] == config.RabbitMqExpectedPort {
			rabbitmqPort = m[2]
		}
		if m[1] == config.SshExpectedPort {
			sshPort = m[2]
		}
	}

	ports := fmt.Sprintf("snapd=%s, rabbitmq=%s", snapdPort, rabbitmqPort)
	if sshPort != "" {
		ports += fmt.Sprintf(", ssh=%s", sshPort)
	}
	return ports
}

func AddUserToContainer(containerName string, user gql.User) error {
	// Username is mail address of user but only everything before the @
	username := strings.Split(user.Email, "@")[0]

	// Add user with home directory
	_, err := docker.ExecuteCommandInContainer(containerName, fmt.Sprintf("useradd -m %s -s /usr/bin/bash", username))
	if err != nil {
		return err
	}

	// Create .ssh directory and authorized_keys file
	_, err = docker.ExecuteCommandInContainer(containerName, fmt.Sprintf("mkdir -p /home/%s/.ssh", username))
	if err != nil {
		return err
	}

	// Add public key to authorized_keys
	_, err = docker.ExecuteCommandInContainer(containerName, fmt.Sprintf("echo %s >> /home/%s/.ssh/authorized_keys", user.SshPublicKey, username))
	return err
}

func GetRunningContainersRequiringCoapPort(containers []docker.ContainerInfo) ([]docker.ContainerInfo, error) {
	var coapContainers []docker.ContainerInfo
	for _, container := range containers {
		if strings.Contains(container.State, "running") && strings.Contains(container.Ports, fmt.Sprintf("%s->%s/udp", config.CoapExpectedPort, config.CoapExpectedPort)) {
			coapContainers = append(coapContainers, container)
		}
	}
	return coapContainers, nil
}

func ContainerInfoListToContainerNameList(containers []docker.ContainerInfo) []string {
	var containerNames []string
	for _, container := range containers {
		containerNames = append(containerNames, container.Name)
	}
	return containerNames
}

func IsCoapPortInUseByAnotherVirtualDevice(containersRequiringCoap []docker.ContainerInfo, containerName string) (bool, []string) {
	var runningContainersUsingCoap []string
	for _, container := range containersRequiringCoap {
		if container.Name == containerName {
			continue
		}
		runningContainersUsingCoap = append(runningContainersUsingCoap, container.Name)
	}
	return len(runningContainersUsingCoap) > 0, runningContainersUsingCoap
}

func IsContainerRequiringCoapPort(containersRequiringCoap []docker.ContainerInfo, containerName string) bool {
	for _, container := range containersRequiringCoap {
		if container.Name == containerName {
			return true
		}
	}
	return false
}

func IsContainerRunning(containers []docker.ContainerInfo, containerName string) bool {
	for _, container := range containers {
		if container.Name == containerName {
			return strings.Contains(container.State, "running")
		}
	}
	return false
}
