package device

import (
	"encoding/json"
	"fmt"
	coap_server "m2cp/coap/coap-server"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeersParsing(t *testing.T) {
	list := `[{"peer":{"os_serial":"992aca91-b5a8-5f30-96cc-65b02312b1f7","address":"fe80::df7:2023:1:5%eth1","hw_serial":"11020","hw_model":"unknown","hw_vendor":"ml-pa.com","dev_model":"dft_sensor","app_name":"dft_sensor","app_version":"4.0.2","seq_no":3002008,"last_seen":"2025-08-22T07:13:26.129204867Z","keys":["t7u2cI1x+twxkxq+W3tMH+nkMcsL4iqYdwRbjttRAO8=","ugxPiX1fP1xBnN7QV6a8VuvM429wUHBtRgAz63apav0="],"update_result_time":"0001-01-01T00:00:00Z","firmware_type":-1}},{"peer":{"os_serial":"b85ef6c0-3d51-593c-80b1-5ddcdb40a82a","address":"fe80::df7:2023:1:1%eth1","hw_serial":"10678","dev_model":"dft_control","seq_no":2004002,"last_seen":"2025-08-22T07:13:26.1673599Z","uptime":43233,"update_result_time":"0001-01-01T00:00:00Z","firmware_type":-1,"hardware_revision":-1,"legacy":true}},{"peer":{"os_serial":"c37e751e-110a-54b9-81a4-a36428d5f6fc","address":"fe80::df7:2023:1:3%eth1","hw_serial":"9266","dev_model":"dft_power","seq_no":23005009,"last_seen":"2025-08-22T07:13:26.170296941Z","uptime":43235,"update_result_time":"0001-01-01T00:00:00Z","firmware_type":-1,"hardware_revision":-1,"legacy":true}},{"peer":{"os_serial":"d4a734a3-0acb-5d3d-9693-5472d099ac12","address":"fe80::df7:2023:1:4%eth1","hw_serial":"11190","dev_model":"dft_communication","seq_no":2008011,"last_seen":"2025-08-22T07:13:26.184912895Z","uptime":43230,"update_result_time":"0001-01-01T00:00:00Z","firmware_type":-1,"hardware_revision":-1,"legacy":true}}]`
	var peers []struct {
		Peer coap_server.Peer `json:"peer"`
	}
	err := json.Unmarshal([]byte(list), &peers)
	assert.NoError(t, err)
	fmt.Println(peers)
}
