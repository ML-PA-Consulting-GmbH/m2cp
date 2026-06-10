package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"m2cpcli/env"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	RequiredDockerVersionMinimum = 25
)

type ContainerInfo struct {
	Command      string `json:"Command"`
	CreatedAt    string `json:"CreatedAt"`
	ID           string `json:"ID"`
	Image        string `json:"DockerImage"`
	Labels       string `json:"Labels"`
	LocalVolumes string `json:"LocalVolumes"`
	Mounts       string `json:"Mounts"`
	Name         string `json:"Names"`
	Networks     string `json:"Networks"`
	Ports        string `json:"Ports"`
	PortsList    []ContainerInfoPortExpose
	RunningFor   string `json:"RunningFor"`
	Size         string `json:"Size"`
	State        string `json:"State"`
	Status       string `json:"Status"`
}

type ContainerInfoPortExpose struct {
	HostPort      int
	ContainerPort int
	Protocol      string
}

type InspectPortBinding struct {
	HostIp   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type InspectInfo struct {
	Id      string    `json:"Id"`
	Created time.Time `json:"Created"`
	//Path    string    `json:"Path"`
	//Args    []string  `json:"Args"`
	//State   struct {
	Status  string `json:"Status"`
	Running bool   `json:"Running"`
	//	Paused     bool      `json:"Paused"`
	//	Restarting bool      `json:"Restarting"`
	//	OOMKilled  bool      `json:"OOMKilled"`
	//	Dead       bool      `json:"Dead"`
	//	Pid        int       `json:"Pid"`
	//	ExitCode   int       `json:"ExitCode"`
	//	Error      string    `json:"Error"`
	//	StartedAt  time.Time `json:"StartedAt"`
	//	FinishedAt time.Time `json:"FinishedAt"`
	//} `json:"State"`
	Image string `json:"Image"`
	//ResolvConfPath  string      `json:"ResolvConfPath"`
	//HostnamePath    string      `json:"HostnamePath"`
	//HostsPath       string      `json:"HostsPath"`
	//LogPath         string      `json:"LogPath"`
	Name string `json:"Name"`
	//RestartCount    int         `json:"RestartCount"`
	//Driver          string      `json:"Driver"`
	//Platform        string      `json:"Platform"`
	//MountLabel      string      `json:"MountLabel"`
	//ProcessLabel    string      `json:"ProcessLabel"`
	//AppArmorProfile string      `json:"AppArmorProfile"`
	//ExecIDs         interface{} `json:"ExecIDs"`
	HostConfig struct {
		//	Binds           []string `json:"Binds"`
		//	ContainerIDFile string   `json:"ContainerIDFile"`
		//	LogConfig       struct {
		//		Type   string `json:"Type"`
		//		Config struct {
		//		} `json:"Config"`
		//	} `json:"LogConfig"`
		//	NetworkMode  string `json:"NetworkMode"`
		PortBindings map[string][]InspectPortBinding `json:"PortBindings"`
		//	RestartPolicy struct {
		//		Name              string `json:"Name"`
		//		MaximumRetryCount int    `json:"MaximumRetryCount"`
		//	} `json:"RestartPolicy"`
		//	AutoRemove           bool          `json:"AutoRemove"`
		//	VolumeDriver         string        `json:"VolumeDriver"`
		//	VolumesFrom          interface{}   `json:"VolumesFrom"`
		//	ConsoleSize          []int         `json:"ConsoleSize"`
		//	CapAdd               interface{}   `json:"CapAdd"`
		//	CapDrop              interface{}   `json:"CapDrop"`
		//	CgroupnsMode         string        `json:"CgroupnsMode"`
		//	Dns                  []interface{} `json:"Dns"`
		//	DnsOptions           []interface{} `json:"DnsOptions"`
		//	DnsSearch            []interface{} `json:"DnsSearch"`
		//	ExtraHosts           interface{}   `json:"ExtraHosts"`
		//	GroupAdd             interface{}   `json:"GroupAdd"`
		//	IpcMode              string        `json:"IpcMode"`
		//	Cgroup               string        `json:"Cgroup"`
		//	Links                interface{}   `json:"Links"`
		//	OomScoreAdj          int           `json:"OomScoreAdj"`
		//	PidMode              string        `json:"PidMode"`
		//	Privileged           bool          `json:"Privileged"`
		//	PublishAllPorts      bool          `json:"PublishAllPorts"`
		//	ReadonlyRootfs       bool          `json:"ReadonlyRootfs"`
		//	SecurityOpt          []string      `json:"SecurityOpt"`
		//	UTSMode              string        `json:"UTSMode"`
		//	UsernsMode           string        `json:"UsernsMode"`
		//	ShmSize              int           `json:"ShmSize"`
		//	Runtime              string        `json:"Runtime"`
		//	Isolation            string        `json:"Isolation"`
		//	CpuShares            int           `json:"CpuShares"`
		//	Memory               int           `json:"Memory"`
		//	NanoCpus             int           `json:"NanoCpus"`
		//	CgroupParent         string        `json:"CgroupParent"`
		//	BlkioWeight          int           `json:"BlkioWeight"`
		//	BlkioWeightDevice    []interface{} `json:"BlkioWeightDevice"`
		//	BlkioDeviceReadBps   []interface{} `json:"BlkioDeviceReadBps"`
		//	BlkioDeviceWriteBps  []interface{} `json:"BlkioDeviceWriteBps"`
		//	BlkioDeviceReadIOps  []interface{} `json:"BlkioDeviceReadIOps"`
		//	BlkioDeviceWriteIOps []interface{} `json:"BlkioDeviceWriteIOps"`
		//	CpuPeriod            int           `json:"CpuPeriod"`
		//	CpuQuota             int           `json:"CpuQuota"`
		//	CpuRealtimePeriod    int           `json:"CpuRealtimePeriod"`
		//	CpuRealtimeRuntime   int           `json:"CpuRealtimeRuntime"`
		//	CpusetCpus           string        `json:"CpusetCpus"`
		//	CpusetMems           string        `json:"CpusetMems"`
		//	Devices              []interface{} `json:"Devices"`
		//	DeviceCgroupRules    interface{}   `json:"DeviceCgroupRules"`
		//	DeviceRequests       interface{}   `json:"DeviceRequests"`
		//	MemoryReservation    int           `json:"MemoryReservation"`
		//	MemorySwap           int           `json:"MemorySwap"`
		//	MemorySwappiness     interface{}   `json:"MemorySwappiness"`
		//	OomKillDisable       interface{}   `json:"OomKillDisable"`
		//	PidsLimit            interface{}   `json:"PidsLimit"`
		//	Ulimits              []interface{} `json:"Ulimits"`
		//	CpuCount             int           `json:"CpuCount"`
		//	CpuPercent           int           `json:"CpuPercent"`
		//	IOMaximumIOps        int           `json:"IOMaximumIOps"`
		//	IOMaximumBandwidth   int           `json:"IOMaximumBandwidth"`
		//	MaskedPaths          interface{}   `json:"MaskedPaths"`
		//	ReadonlyPaths        interface{}   `json:"ReadonlyPaths"`
	} `json:"HostConfig"`
	//GraphDriver struct {
	//	Data struct {
	//		LowerDir  string `json:"LowerDir"`
	//		MergedDir string `json:"MergedDir"`
	//		UpperDir  string `json:"UpperDir"`
	//		WorkDir   string `json:"WorkDir"`
	//	} `json:"Data"`
	//	Name string `json:"Name"`
	//} `json:"GraphDriver"`
	//Mounts []struct {
	//	Type        string `json:"Type"`
	//	Source      string `json:"Source"`
	//	Destination string `json:"Destination"`
	//	Mode        string `json:"Mode"`
	//	RW          bool   `json:"RW"`
	//	Propagation string `json:"Propagation"`
	//} `json:"Mounts"`
	Config struct {
		//Hostname     string `json:"Hostname"`
		//Domainname   string `json:"Domainname"`
		//User         string `json:"User"`
		//AttachStdin  bool   `json:"AttachStdin"`
		//AttachStdout bool   `json:"AttachStdout"`
		//AttachStderr bool   `json:"AttachStderr"`
		ExposedPorts map[string]struct{} `json:"ExposedPorts"`
		//Tty        bool        `json:"Tty"`
		//OpenStdin  bool        `json:"OpenStdin"`
		//StdinOnce  bool        `json:"StdinOnce"`
		//Env        []string    `json:"Env"`
		//Cmd        []string    `json:"Cmd"`
		//Image      string      `json:"Image"`
		//Volumes    interface{} `json:"Volumes"`
		//WorkingDir string      `json:"WorkingDir"`
		//Entrypoint []string    `json:"Entrypoint"`
		//OnBuild    interface{} `json:"OnBuild"`
		//Labels     struct {
		//	OrgOpencontainersImageRefName string `json:"org.opencontainers.image.ref.name"`
		//	OrgOpencontainersImageVersion string `json:"org.opencontainers.image.version"`
		//} `json:"Labels"`
		//StopSignal string `json:"StopSignal"`
	} `json:"Config"`
	//NetworkSettings struct {
	//	Bridge     string `json:"Bridge"`
	//	SandboxID  string `json:"SandboxID"`
	//	SandboxKey string `json:"SandboxKey"`
	//	Ports      struct {
	//	} `json:"Ports"`
	//	HairpinMode            bool        `json:"HairpinMode"`
	//	LinkLocalIPv6Address   string      `json:"LinkLocalIPv6Address"`
	//	LinkLocalIPv6PrefixLen int         `json:"LinkLocalIPv6PrefixLen"`
	//	SecondaryIPAddresses   interface{} `json:"SecondaryIPAddresses"`
	//	SecondaryIPv6Addresses interface{} `json:"SecondaryIPv6Addresses"`
	//	EndpointID             string      `json:"EndpointID"`
	//	Gateway                string      `json:"Gateway"`
	//	GlobalIPv6Address      string      `json:"GlobalIPv6Address"`
	//	GlobalIPv6PrefixLen    int         `json:"GlobalIPv6PrefixLen"`
	//	IPAddress              string      `json:"IPAddress"`
	//	IPPrefixLen            int         `json:"IPPrefixLen"`
	//	IPv6Gateway            string      `json:"IPv6Gateway"`
	//	MacAddress             string      `json:"MacAddress"`
	//	Networks               struct {
	//		Bridge struct {
	//			IPAMConfig          interface{} `json:"IPAMConfig"`
	//			Links               interface{} `json:"Links"`
	//			Aliases             interface{} `json:"Aliases"`
	//			MacAddress          string      `json:"MacAddress"`
	//			DriverOpts          interface{} `json:"DriverOpts"`
	//			NetworkID           string      `json:"NetworkID"`
	//			EndpointID          string      `json:"EndpointID"`
	//			Gateway             string      `json:"Gateway"`
	//			IPAddress           string      `json:"IPAddress"`
	//			IPPrefixLen         int         `json:"IPPrefixLen"`
	//			IPv6Gateway         string      `json:"IPv6Gateway"`
	//			GlobalIPv6Address   string      `json:"GlobalIPv6Address"`
	//			GlobalIPv6PrefixLen int         `json:"GlobalIPv6PrefixLen"`
	//			DNSNames            interface{} `json:"DNSNames"`
	//		} `json:"bridge"`
	//	} `json:"Networks"`
	//} `json:"NetworkSettings"`
}

func IsDockerReady() error {
	if IsDockerInstalled() == false {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("docker is not installed but required for virtual devices. Please install 'Docker Desktop'")
		} else {
			return fmt.Errorf("docker is not installed but required for virtual devices. Please install docker")
		}
	}
	if IsDockerRunning() == false {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("docker is installed but not running, please start 'Docker Desktop'")
		} else {
			return fmt.Errorf("docker is installed but not running, please start docker with `systemctl start docker`")
		}
	}
	version, err := GetDockerVersion()
	if err != nil {
		return fmt.Errorf("could not get docker version: %w", err)
	}
	installedSemver, err := env.NewSemanticVersion(version)
	if err != nil {
		return fmt.Errorf("could not parse docker version: %w", err)
	}
	if installedSemver.Major < RequiredDockerVersionMinimum {
		return fmt.Errorf("docker version is less than required: %s < %d.0.0", installedSemver, RequiredDockerVersionMinimum)
	}
	return nil
}

func IsDockerInstalled() bool {
	dockerPath, err := getDockerBinaryPath()
	if err != nil {
		return false
	}

	if _, err := os.Stat(dockerPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func GetDockerVersion() (string, error) {
	// Return the value of field `Version`
	output, err := ExecuteDockerCommand("version --format '{{.Client.Version}}'")
	if err != nil {
		return "", err
	}
	output = strings.ReplaceAll(output, "\n", "")
	output = strings.ReplaceAll(output, "'", "")
	return output, nil
}

func IsDockerRunning() bool {
	dockerPath, err := getDockerBinaryPath()
	if err != nil {
		return false
	}

	cmd := exec.Command(dockerPath, "info")
	err = cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func GetContainers() ([]ContainerInfo, error) {
	output, err := ExecuteDockerCommand("ps -a --format json")
	if err != nil {
		return nil, err
	}
	containersJson := strings.Split(output, "\n")
	var containers []ContainerInfo
	for _, s := range containersJson {
		if len(s) == 0 {
			continue
		}
		var containerInfo ContainerInfo
		err := json.Unmarshal([]byte(s), &containerInfo)
		if err != nil {
			return nil, err
		}

		// We do not get Port information, when the container is not running. We have to get it from docker inspect.
		// We use the output format of `docker ps` for exited containers
		if containerInfo.State == "exited" {
			inspectOutput, err := ExecuteDockerCommand(fmt.Sprintf("inspect %s", containerInfo.ID))
			if err != nil {
				return nil, err
			}
			var inspectInfo []InspectInfo
			err = json.Unmarshal([]byte(inspectOutput), &inspectInfo)
			if err != nil {
				return nil, err
			}
			if len(inspectInfo) != 1 {
				return nil, fmt.Errorf("expected exactly one container inspect info, got %d", len(inspectInfo))
			}

			for port, bindings := range inspectInfo[0].HostConfig.PortBindings {
				for _, binding := range bindings {
					if containerInfo.Ports != "" {
						containerInfo.Ports += ", "
					}
					// HostIp is empty, when the container is not running. We simply use localhost  here as we usually see 0.0.0.0
					containerInfo.Ports += fmt.Sprintf("0.0.0.0:%s->%s, :::%s->%s",
						binding.HostPort, port, binding.HostPort, port)
				}
			}
		}

		// Parse the ports into a more usable format
		for _, port := range strings.Split(containerInfo.Ports, ", ") {
			containerInfo.PortsList = append(containerInfo.PortsList, parsePortStringToContainerInfoPortExpose(port)...)
		}

		containers = append(containers, containerInfo)
	}
	return containers, err
}

func parsePortStringToContainerInfoPortExpose(ports string) []ContainerInfoPortExpose {
	var portsDetailed []ContainerInfoPortExpose
	regex := regexp.MustCompile(`:::(\d+)->(\d+)/(tcp|udp)`)
	matches := regex.FindAllStringSubmatch(ports, -1)

	for _, match := range matches {
		hostPort, _ := strconv.Atoi(match[1])
		containerPort, _ := strconv.Atoi(match[2])
		protocol := match[3]

		portsDetailed = append(portsDetailed, ContainerInfoPortExpose{
			HostPort:      hostPort,
			ContainerPort: containerPort,
			Protocol:      protocol,
		})
	}

	return portsDetailed
}

func ExecuteCommandInContainer(container string, cmd string) (string, error) {
	dockerPath, err := getDockerBinaryPath()
	if err != nil {
		return "", err
	}

	proc := exec.Command(dockerPath, "exec", container, "/bin/bash", "-c", cmd)
	if proc.Err != nil {
		return "", proc.Err
	}
	outputByte, err := proc.Output()
	if err != nil {
		return "", err
	}
	return string(outputByte), err
}

func CreateDockerContainer(name string, hostname string, image string, volumes string, ports []string, noPull bool, extraVars string) error {
	// I was simply not able to start a container with 'docker run'. There was another post on stackoverflow from other
	// people having the same issues. I saw two approaches here: try the docker call via bash as intermediate shell,
	// or use the docker client library for go. First approach was easier to achieve and worked, so I went with that.
	var proc *exec.Cmd
	argName := fmt.Sprintf("--hostname=%s --name=%s", hostname, name)

	argVolumes := ""
	if isOsDebianBased() {
		// On debian based machines we mount the host certificates, as our customers could use custom certificates in their networks
		argVolumes =
			fmt.Sprintf("-v /etc/ssl/certs:/etc/ssl/certs:ro -v /usr/local/share/ca-certificates:/usr/local/share/ca-certificates:ro")
	}
	if volumes != "" {
		// Todo: Add support for multiple volumes from user input
		argVolumes = fmt.Sprintf("%s -v %s", argVolumes, volumes)
	}
	argPull := "--pull always"
	if noPull {
		argPull = "--pull never"
	}

	createCommand := []string{"run", "-d", "--restart=unless-stopped", "--privileged"}
	createCommand = append(createCommand, strings.Split(argName, " ")...)
	for _, port := range ports {
		createCommand = append(createCommand, "-p", port)
	}
	if argVolumes != "" {
		createCommand = append(createCommand, strings.Split(argVolumes, " ")...)
	}
	createCommand = append(createCommand, strings.Split(argPull, " ")...)
	if extraVars != "" {
		createCommand = append(createCommand, strings.Split(extraVars, " ")...)
	}
	createCommand = append(createCommand, image)

	dockerPath, err := getDockerBinaryPath()
	if err != nil {
		return err
	}
	proc = exec.Command(dockerPath, createCommand...)
	output, err := proc.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed docker command: %s %s\n", dockerPath, strings.Join(createCommand, " "))
		return fmt.Errorf("could not get combined output when creating docker container: %s", output)
	}
	return nil
}

func ExecuteDockerCommand(cmd string) (string, error) {
	//fmt.Println(fmt.Sprintf("DEBUG docker %s", cmd))
	dockerPath, err := getDockerBinaryPath()
	if err != nil {
		return "", err
	}

	cmdSplit := strings.Split(cmd, " ")
	proc := exec.Command(dockerPath, cmdSplit...)
	if proc.Err != nil {
		return "", proc.Err
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	proc.Stdout = &stdout
	proc.Stderr = &stderr

	err = proc.Run()
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

func AttachToContainer(container string) {
	dockerPath, err := getDockerBinaryPath()
	if err != nil {
		fmt.Println(err)
	}
	cmd := exec.Command(dockerPath, "exec", "-it", container, "bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Ignoring error here
	_ = cmd.Run()
}

func isOsDebianBased() bool {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		// Any distribution with systemd should have /etc/os-release, if we can´t find/open it -> very likely something exotic
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ID_LIKE=debian") || strings.Contains(line, "ID=ubuntu") {
			return true
		}
	}
	return false
}

func IsValidDockerName(name string) (bool, error) {
	match, err := regexp.MatchString(`^[a-zA-Z0-9-]+$`, name)
	if err != nil {
		return false, err
	}
	if !match {
		return false, fmt.Errorf("name must match the pattern [a-zA-Z0-9-]+")
	}
	return true, nil
}

func getDockerBinaryPath() (string, error) {
	// Differentiate between linux and windows
	switch runtime.GOOS {
	case "windows":
		return "C:\\Program Files\\Docker\\Docker\\resources\\bin\\docker.exe", nil
	case "linux":
		return "/usr/bin/docker", nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
