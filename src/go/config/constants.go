package config

import "time"

const (
	MLPA              = "mlpa"
	SnapStoreApiLevel = 1
)

// RabbitMqExpectedPort Expected port of RabbitMq inside the container
const RabbitMqExpectedPort string = "5672"

// SnapdExpectedPort Expected port of Snapd inside the container
const SnapdExpectedPort string = "5555"

// SshExpectedPort Expected port of SSH inside the container
const SshExpectedPort string = "22"

// CoapExpectedPort Expected port of CoAP inside the container
const CoapExpectedPort string = "5683"

const GraphQLTimeout time.Duration = 10 * time.Second
const GraphQLMaxPageSize int = 1000
