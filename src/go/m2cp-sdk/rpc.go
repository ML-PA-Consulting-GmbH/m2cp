package m2cp

import (
	"time"
)

type RpcCommand interface {
	GetName() string
	GetDocumentation() RpcDocumentation
	Execute(ctp ContextPlus, params map[string]string) RpcResult
}

type RpcHandler interface {
	Process(ctp ContextPlus, commandMessage CommandMessage) (ResponseMessage, error)
	GetDocumentation() []RpcDocumentation
}

type RpcDefinitions interface {
}

type RpcDocumentation interface {
	GetName() string
	ToJson() []byte
}

type paramsReadable interface {
	Has(field string) bool
	GetInt(field string, defaultValue int) int
	GetLong(field string, defaultValue int64) int64
	GetIntArray(field string) []int
	GetDouble(field string, defaultValue float64) float64
	GetDoubleArray(field string) []float64
	GetBool(field string, defaultValue bool) bool
	GetString(field string, defaultValue string) string
	GetBinary(field string) []byte
	GetDateTime(s string, defaultValue time.Time) time.Time
	GetJson(field string, outVar interface{}) error
	GetRaw(field string, defaultValue string) string
	GetAllRaw() map[string]string
	Validate(field string, t Type, optional bool) error
	ValidateAll() error
}

type paramsWritable interface {
	SetInt(field string, value int)
	SetLong(field string, value int64)
	SetIntArray(field string, value []int)
	SetDouble(field string, value float64)
	SetDoubleArray(field string, value []float64)
	SetBool(field string, value bool)
	SetString(field string, value string)
	SetBinary(field string, value []byte)
	SetDateTime(s string, value time.Time)
	SetJson(field string, value interface{}) error
}

type resultReadable interface {
	GetMessage() string
	IsSuccessful() bool
	GetErrorCode() RpcErrorCode
}

type resultWritable interface {
	SetMessage(message string)
	SetSuccess(success bool)
	SetErrorCode(code RpcErrorCode)
}

type RpcParameters interface {
	paramsReadable
}

type RpcResult interface {
	paramsReadable
	paramsWritable
	resultReadable
	resultWritable
	GetTraces() []MessageTrace
}

type RpcResultReadonly interface {
	paramsReadable
	resultReadable
	GetTraces() []MessageTrace
}

type RpcParameter interface {
	GetName() string
	GetType() Type
	IsOptional() bool
	GetDescription() string
}

type RpcErrorCode int

const (
	RpcErrorCodeSuccess                        RpcErrorCode = 0
	RpcErrorCodeFailed                         RpcErrorCode = 1
	RpcErrorCodeParametersMismatchDefinition   RpcErrorCode = 400
	RpcErrorCodeUnauthorized                   RpcErrorCode = 401
	RpcErrorCodeCommandNotFound                RpcErrorCode = 404
	RpcErrorCodeRequestTimeout                 RpcErrorCode = 408
	RpcErrorCodeConflict                       RpcErrorCode = 409
	RpcErrorCodeReturnValuesMismatchDefinition RpcErrorCode = 500
	RpcErrorCodeNotImplemented                 RpcErrorCode = 501
	RpcErrorCodeGatewayTimeout                 RpcErrorCode = 504
	RpcErrorCodeInsufficientStorage            RpcErrorCode = 507
	RpcErrorCodeLoopDetected                   RpcErrorCode = 508
	RpcErrorCodeNetworkAuthenticationRequired  RpcErrorCode = 511
)
