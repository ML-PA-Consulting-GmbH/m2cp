package messages

import (
	"m2cp/device"
)

// TestNewLocalAddress tests automatic replacement of ".local" with the device name
func (t *TestSuite) TestAddressFormatValidation() {
	var err error
	valid := func(address string) {
		_, err = NewAddress(t.ctp, address)
		t.NoError(err, "expected '%s' to be valid, but got error: %s", address, err)
	}
	invalid := func(address string) {
		_, err = NewAddress(t.ctp, address)
		t.Error(err, "expected '%s' to be invalid, but got no error", address)
	}

	valid("a.b.c.foo.bar.local")
	valid("a.b.[::1]")
	valid("A.b.123e4567-e89b-12d3-a456-426614174000")
	valid("_._.123e4567-e89b-12d3-a456-426614174000")
	invalid("foo!.local")
	valid("FooBar.local")
	invalid("foo.bar.local.")
	invalid(".foo.bar.local")
}

// TestNewLocalAddress tests automatic replacement of ".local" with the device name
func (t *TestSuite) TestNewLocalAddressTooShort() {
	_, err := NewAddress(t.ctp, "local")
	t.Error(err)
}

// TestNewLocalAddress tests automatic replacement of ".local" with the device name
func (t *TestSuite) TestNewLocalAppAddressValid() {
	_, err := NewAddress(t.ctp, "app.local")
	t.NoError(err)
}

func (t *TestSuite) TestNewLocalNodeAddressValid() {
	a, err := NewAddress(t.ctp, "foo.bar.local")
	t.NoError(err)
	deviceName, err := device.GetName(t.ctp)
	t.NoError(err)
	t.Equal("foo.bar."+deviceName, a.GetAddress())
	t.Len(a.GetAddress(), 44)
}

func (t *TestSuite) TestNewLocalLongNodeAddressValid() {
	a, err := NewAddress(t.ctp, "a.b.c.foo.bar.local")
	t.NoError(err)
	deviceName, err := device.GetName(t.ctp)
	t.NoError(err)
	t.Equal("a.b.c.foo.bar."+deviceName, a.GetAddress())
	t.Equal("a.b.c.foo", a.GetNodeName())
	t.Equal("bar", a.GetAppName())
	t.Len(a.GetAddress(), 50)
}

func (t *TestSuite) TestNewCoapAddress() {
	a, err := NewAddress(t.ctp, "foo.bar.[::1]:5500")
	t.NoError(err)
	t.True(a.IsIpRouted())
	t.Equal("[::1]:5500", a.GetDeviceName())
	t.Equal("::1", a.GetDeviceIp())
	t.Equal("5500", a.GetDevicePort())
}

func (t *TestSuite) TestNewCoapAddressIp6WithInterface() {
	a, err := NewAddress(t.ctp, "rpc.m2cp-gateway.[fe80::df7:2023:1:2%eth0]")
	t.NoError(err)
	t.True(a.IsIpRouted())
	t.Equal("[fe80::df7:2023:1:2%eth0]:5683", a.GetDeviceName())
	t.Equal("fe80::df7:2023:1:2%eth0", a.GetDeviceIp())
	t.Equal("5683", a.GetDevicePort())
}

func (t *TestSuite) TestNewCoapAddressNoPort() {
	a, err := NewAddress(t.ctp, "foo.bar.[::1]")
	t.NoError(err)
	t.True(a.IsIpRouted())
	t.Equal("[::1]:5683", a.GetDeviceName())
	t.Equal("::1", a.GetDeviceIp())
	t.Equal("5683", a.GetDevicePort())
}

func (t *TestSuite) TestNewCoapAddressIp4() {
	a, err := NewAddress(t.ctp, "foo.bar.[192.168.10.12]:5500")
	t.NoError(err)
	t.True(a.IsIpRouted())
	t.Equal("[192.168.10.12]:5500", a.GetDeviceName())
	t.Equal("192.168.10.12", a.GetDeviceIp())
	t.Equal("5500", a.GetDevicePort())
}

func (t *TestSuite) TestNewCoapAddressIp4Long() {
	a, err := NewAddress(t.ctp, "a.b.c.foo.bar.[192.168.10.12]:5500")
	t.NoError(err)
	t.True(a.IsIpRouted())
	t.Equal("[192.168.10.12]:5500", a.GetDeviceName())
	t.Equal("192.168.10.12", a.GetDeviceIp())
	t.Equal("5500", a.GetDevicePort())
	t.Equal("a.b.c.foo", a.GetNodeName())
	t.Equal("bar", a.GetAppName())
}
