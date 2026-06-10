package legacy

import (
	"encoding/json"
	"m2cpcli/structs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceQueries(t *testing.T) {
	value := `{"n":30,"mh":{"up":true},"gw":{"fs-free-mb":953,"cpu-pct":4,"up-s":328,"mem-mb":25,"sent-kb":333,"q-n":2,"q-bad-kb":7,"q-kb":11,"q-age-s":-1}}`
	lastUplinkSignalContent := &structs.DeviceUplinkSignalContent{}
	err := json.Unmarshal([]byte(value), lastUplinkSignalContent)
	assert.NoError(t, err)
	assert.Equal(t, 30, lastUplinkSignalContent.N)
	assert.Equal(t, true, *lastUplinkSignalContent.MessageHub.Up)
	assert.Equal(t, 953, *lastUplinkSignalContent.Gateway.FileSystemFreeMb)
	assert.Equal(t, 4, *lastUplinkSignalContent.Gateway.SystemCpuUsage)
	assert.Equal(t, 328, *lastUplinkSignalContent.Gateway.UptimeSeconds)
	assert.Equal(t, 25, *lastUplinkSignalContent.Gateway.MemGatewayMB)
	assert.Equal(t, 333, *lastUplinkSignalContent.Gateway.SentKB)
	assert.Equal(t, 2, *lastUplinkSignalContent.Gateway.BufferCount)
	assert.Equal(t, 7, *lastUplinkSignalContent.Gateway.BufferBadKB)
	assert.Equal(t, 11, *lastUplinkSignalContent.Gateway.BufferSizeKB)
	assert.Equal(t, -1, *lastUplinkSignalContent.Gateway.BufferAgeSeconds)

}
