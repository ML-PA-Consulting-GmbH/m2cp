package networks

/**
 * A domain is the address space of an app in a m2cp-messaing-network.
 * The app can instantiate multiple nodes in this domain.
 * Domains follow the scheme <appname>.<devicename>
 * Nodes follow the scheme <nodename>.<appname>.<devicename>
 */

import (
	"fmt"
	"m2cp"
	"m2cp/device"
	"math/rand"
	"os"
	"strings"
)

type domain struct {
	appName    string
	deviceName string
	tokens     []string
}

func newDomain(ctx m2cp.ContextPlus, appName, deviceName string) (*domain, error) {
	var err error

	d := domain{
		appName:    appName,
		deviceName: deviceName,
	}

	if appName == "" {
		d.appName = getContainerName(ctx)
	}

	if deviceName == "" {
		d.deviceName, err = device.GetName(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &d, nil
}

func newDomainCustom(name string) *domain {
	tokens := strings.Split(name, ".")
	d := &domain{
		appName:    tokens[0],
		deviceName: tokens[1],
	}
	return d
}

func newDomainTest(ctx m2cp.ContextPlus) (*domain, error) {
	deviceName, err := device.GetName(ctx)
	if err != nil {
		return nil, err
	}
	return &domain{
		appName:    "test" + getNextTestCaseNumber(),
		deviceName: deviceName,
		tokens:     []string{},
	}, nil
}

func (d *domain) GetName() string {
	return fmt.Sprintf("%s.%s", d.appName, d.deviceName)
}

func (d *domain) GetAppName() string {
	return d.appName
}

func (d *domain) GetDeviceName() string {
	return d.deviceName
}

func (d *domain) HasSameDeviceName(address string) bool {
	return strings.Split(address, ".")[len(strings.Split(address, "."))-1] == d.deviceName
}

func getContainerName(ctx m2cp.ContextPlus) string {
	containerName := os.Getenv("SNAP_NAME")
	if containerName != "" {
		return containerName
	}
	developmentName := "development-" + getNextRandomNumber()
	ctx.LogDebug("SNAP_NAME not set, created random container name: %s", developmentName)
	return developmentName
}

var testCaseCounter int

func getNextTestCaseNumber() string {
	testCaseCounter++
	return fmt.Sprintf("%d", testCaseCounter)
}

func getNextRandomNumber() string {
	return fmt.Sprintf("%d", rand.Intn(999999))
}
