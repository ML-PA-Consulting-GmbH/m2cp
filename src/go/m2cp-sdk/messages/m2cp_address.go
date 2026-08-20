package messages

import (
	"fmt"
	"m2cp"
	"m2cp/device"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

type m2CPAddress struct {
	NodeName     string
	AppName      string
	DeviceName   string
	Subtopic     string
	CoapDelivery bool
	CoapIp       string
	CoapPort     string
}

func NewAddress(ctp m2cp.ContextPlus, address string) (m2cp.Address, error) {
	return NewAddressWithSubtopic(ctp, address, "")
}

func NewAddressWithSubtopic(ctp m2cp.ContextPlus, address, subtopic string) (m2cp.Address, error) {
	if address[len(address)-1] == '.' {
		return nil, fmt.Errorf("address cannot end with a dot: %s", address)
	}
	if address[0] == '.' {
		return nil, fmt.Errorf("address cannot start with a dot: %s", address)
	}

	tokens := splitAddress(address)

	if len(tokens) < 2 {
		return nil, fmt.Errorf("invalid address format: %s", address)
	}

	for i, token := range tokens {
		if i == len(tokens)-1 {
			break
		}
		if err := validateName(token); err != nil {
			return nil, fmt.Errorf("invalid address format for '%s'. token '%s' is not valid: %w", address, token, err)
		}
	}

	m2cpAddress := &m2CPAddress{
		Subtopic:   subtopic,
		NodeName:   strings.Join(tokens[0:len(tokens)-2], "."),
		AppName:    tokens[len(tokens)-2],
		DeviceName: tokens[len(tokens)-1],
	}

	if isCoapAddressIp6(m2cpAddress.DeviceName) || isCoapAddressIp4(m2cpAddress.DeviceName) {
		// normalize the address
		m2cpAddress.CoapIp, m2cpAddress.CoapPort, _ = extractIpAndPort(m2cpAddress.DeviceName)
		if m2cpAddress.CoapPort == "" {
			m2cpAddress.CoapPort = "5683"
		}
		m2cpAddress.DeviceName = fmt.Sprintf("[%s]:%s", m2cpAddress.CoapIp, m2cpAddress.CoapPort)
		m2cpAddress.CoapDelivery = true
	} else if isValidDeviceName(m2cpAddress.DeviceName) {
		m2cpAddress.CoapDelivery = false
	} else {
		return nil, fmt.Errorf("can't construct m2cp-address with invalid device name: %s", m2cpAddress.DeviceName)
	}

	if m2cpAddress.DeviceName == "local" {
		var err error
		m2cpAddress.DeviceName, err = device.GetName(ctp)
		if err != nil {
			return nil, fmt.Errorf("can't create .local address if device name is not available: %w", err)
		}
	}

	return m2cpAddress, nil
}

func (m *m2CPAddress) GetAddress() string {
	if m.NodeName == "" {
		// this is an app-address
		return fmt.Sprintf("%s.%s", m.AppName, m.DeviceName)
	}
	// this is a node-address
	return fmt.Sprintf("%s.%s.%s", m.NodeName, m.AppName, m.DeviceName)
}

func (m *m2CPAddress) GetSubtopic() string {
	return m.Subtopic
}

func (m *m2CPAddress) GetDeviceName() string {
	return m.DeviceName
}

func (m *m2CPAddress) GetDeviceIp() string {
	return m.CoapIp
}

func (m *m2CPAddress) GetDevicePort() string {
	return m.CoapPort
}

func (m *m2CPAddress) GetAppName() string {
	return m.AppName
}

func (m *m2CPAddress) GetNodeName() string {
	return m.NodeName
}

func (m *m2CPAddress) IsIpRouted() bool {
	return m.CoapDelivery
}

func GetDeviceName(address string) string {
	parts := strings.Split(address, ".")
	return parts[len(parts)-1]
}

var reSplitAddress = regexp.MustCompile(`(?:\[[^\]]+\]:\d+|[^.]+)`)

func splitAddress(address string) []string {
	return reSplitAddress.FindAllString(address, -1)
}

var reCoapAddressIp6 = regexp.MustCompile(`^\[([a-fA-F0-9:]+(?:%[a-zA-Z0-9]+)?)\](?::(\d+))?$`)

// reCoapAddressIp4 is a regex matching IPv4 addresses in the format [x.x.x.x]:port, where x is a number between 0 and 255 and port is an optional port number.
//
// Valid examples:
//   - [192.168.1.1]
//   - [10.0.0.1]
//   - [192.168.1.1]:8080
//   - [10.0.0.1]:5683
//
// Shortened IPs like [10.1] are not valid with this regex.
var reCoapAddressIp4 = regexp.MustCompile(`^\[([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})\](?::(\d+))?$`)

func isCoapAddressIp6(address string) bool {
	return reCoapAddressIp6.MatchString(address)
}

func isCoapAddressIp4(address string) bool {
	return reCoapAddressIp4.MatchString(address)
}

var reDeviceName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isValidDeviceName(name string) bool {
	if _, err := uuid.Parse(name); err == nil {
		return true
	}
	switch name {
	case "local":
		return true
	case "cloud":
		return true
	}

	// this is a less strict check, which we should not allow in the future
	if reDeviceName.MatchString(name) {
		return true
	}

	return false
}

var reExtractIpAndPort = regexp.MustCompile(`^\[([0-9a-fA-F:.%]+)(%[a-zA-Z0-9]+)?\]((:)(\d+))?$`)

//var reExtractIp = regexp.MustCompile(`^\[([0-9a-fA-F:.%]+)\]$`)

func extractIpAndPort(address string) (ip, port string, success bool) {
	matches := reExtractIpAndPort.FindStringSubmatch(address)
	if len(matches) == 6 {
		return fmt.Sprintf("%s%s", matches[1], matches[2]), matches[5], true
	}
	//matches = reExtractIp.FindStringSubmatch(address)
	//if len(matches) == 2 {
	//	return matches[1], "", true
	//}
	return "", "", false
}

// validateName checks if the name contains only lowercase alphanumerics and hyphens
func validateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '-' && char != '_' {
			return fmt.Errorf("name must contain only alphanumeric characters and hyphens and underscores")
		}
	}
	return nil
}
