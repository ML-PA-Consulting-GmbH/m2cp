package rpctypes

import (
	"encoding/base64"
	"encoding/json"
	fmt "fmt"
	"m2cp"
	"strconv"
	"strings"
	"time"
)

type RpcObject struct {
	Dict         map[string]string
	Definitions  map[string]m2cp.RpcParameter
	Error        m2cp.RpcErrorCode
	runtimeError error
	Message      string
	Trace        []m2cp.MessageTrace
}

func NewRpcResultSuccess(message string) m2cp.RpcResult {
	return &RpcObject{
		Dict:        map[string]string{},
		Definitions: map[string]m2cp.RpcParameter{},
		Error:       m2cp.RpcErrorCodeSuccess,
		Message:     message,
		Trace:       []m2cp.MessageTrace{},
	}
}

func NewRpcResultError(message string, errorCode m2cp.RpcErrorCode) m2cp.RpcResult {
	return &RpcObject{
		Dict:        map[string]string{},
		Definitions: map[string]m2cp.RpcParameter{},
		Error:       errorCode,
		Message:     message,
		Trace:       []m2cp.MessageTrace{},
	}
}

func (o *RpcObject) GetMessage() string {
	return o.Message
}

func (o *RpcObject) GetTraces() []m2cp.MessageTrace {
	return o.Trace
}

func (o *RpcObject) SetMessage(message string) {
	o.Message = message
}

func (o *RpcObject) SetSuccess(success bool) {
	if success {
		o.Error = m2cp.RpcErrorCodeSuccess
	} else {
		o.Error = m2cp.RpcErrorCodeFailed
	}
}

func (o *RpcObject) GetErrorCode() m2cp.RpcErrorCode {
	return o.Error
}

func (o *RpcObject) SetErrorCode(code m2cp.RpcErrorCode) {
	o.Error = code
}

func (o *RpcObject) GetInt(field string, defaultValue int) int {
	if err := o.validateGet(field, m2cp.TypeInt); err != nil {
		return defaultValue
	}
	if val, ok := o.Dict[field]; ok {
		n, err := strconv.Atoi(val)
		if err != nil {
			return defaultValue
		}
		return n
	}
	return defaultValue
}

func (o *RpcObject) GetLong(field string, defaultValue int64) int64 {
	if err := o.validateGet(field, m2cp.TypeInt); err != nil {
		return defaultValue
	}
	if val, ok := o.Dict[field]; ok {
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return defaultValue
		}
		return n
	}
	return defaultValue
}

func (o *RpcObject) GetIntArray(field string) []int {
	if err := o.validateGet(field, m2cp.TypeIntArray); err != nil {
		return nil
	}
	if val, ok := o.Dict[field]; ok {
		parts := strings.Split(val, ",")
		result := make([]int, 0, len(parts))
		for _, part := range parts {
			i, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil
			}
			result = append(result, i)
		}
		return result
	}
	return []int{}
}

func (o *RpcObject) GetDouble(field string, defaultValue float64) float64 {
	if err := o.validateGet(field, m2cp.TypeDouble); err != nil {
		return defaultValue
	}
	if val, ok := o.Dict[field]; ok {
		d, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return defaultValue
		}
		return d
	}
	return defaultValue
}

func (o *RpcObject) GetDoubleArray(field string) []float64 {
	if err := o.validateGet(field, m2cp.TypeDoubleArray); err != nil {
		return nil
	}

	if val, ok := o.Dict[field]; ok {
		parts := strings.Split(val, ",")
		result := make([]float64, 0, len(parts))
		for _, part := range parts {
			f, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
			if err != nil {
				return nil
			}
			result = append(result, f)
		}
		return result
	}
	return []float64{}
}

func (o *RpcObject) GetBool(field string, defaultValue bool) bool {
	if err := o.validateGet(field, m2cp.TypeBool); err != nil {
		return defaultValue
	}
	if val, ok := o.Dict[field]; ok {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return defaultValue
		}
		return b
	}
	return defaultValue
}

func (o *RpcObject) GetString(field string, defaultValue string) string {
	if err := o.validateGet(field, m2cp.TypeString); err != nil {
		return defaultValue
	}
	if val, ok := o.Dict[field]; ok {
		return val
	}
	return defaultValue
}

func (o *RpcObject) GetBinary(field string) []byte {
	if err := o.validateGet(field, m2cp.TypeBinary); err != nil {
		return nil
	}
	if val, ok := o.Dict[field]; ok {
		bin, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return nil
		}
		return bin
	}
	return nil
}

func (o *RpcObject) GetDateTime(field string, defaultValue time.Time) time.Time {
	if err := o.validateGet(field, m2cp.TypeDateTime); err != nil {
		return defaultValue
	}
	if val, ok := o.Dict[field]; ok {
		dt, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return defaultValue
		}
		return dt
	}
	return defaultValue
}

func (o *RpcObject) GetJson(field string, outVar interface{}) error {
	if err := o.validateGet(field, m2cp.TypeJson); err != nil {
		return err
	}
	if val, ok := o.Dict[field]; ok {
		return json.Unmarshal([]byte(val), outVar)
	}
	return nil
}

func (o *RpcObject) GetRaw(field string, defaultValue string) string {
	if val, ok := o.Dict[field]; ok {
		return val
	}
	return defaultValue
}

func (o *RpcObject) GetAllRaw() map[string]string {
	return o.Dict
}

func (o *RpcObject) set(field string, value interface{}) error {

	switch v := value.(type) {
	case string:
		o.Dict[field] = v
	case bool:
		o.Dict[field] = strconv.FormatBool(v)
	case int:
		o.Dict[field] = strconv.Itoa(v)
	case int64:
		o.Dict[field] = strconv.FormatInt(v, 10)
	case float64:
		o.Dict[field] = strconv.FormatFloat(v, 'f', -1, 64)
	case []int:
		strValues := make([]string, len(v))
		for i, n := range v {
			strValues[i] = strconv.Itoa(n)
		}
		o.Dict[field] = strings.Join(strValues, ",")
	case []int64:
		strValues := make([]string, len(v))
		for i, n := range v {
			strValues[i] = strconv.FormatInt(n, 10)
		}
		o.Dict[field] = strings.Join(strValues, ",")
	case []float64:
		strValues := make([]string, len(v))
		for i, n := range v {
			strValues[i] = strconv.FormatFloat(n, 'f', -1, 64)
		}
		o.Dict[field] = strings.Join(strValues, ",")
	case []byte:
		o.Dict[field] = base64.StdEncoding.EncodeToString(v)
	default:
		return fmt.Errorf("unsupported value type")
	}
	return nil
}

func (o *RpcObject) SetInt(field string, value int) {
	_ = o.set(field, value)
}

func (o *RpcObject) SetLong(field string, value int64) {
	_ = o.set(field, value)
}

func (o *RpcObject) SetIntArray(field string, value []int) {
	_ = o.set(field, value)
}

func (o *RpcObject) SetDouble(field string, value float64) {
	_ = o.set(field, value)
}

func (o *RpcObject) SetDoubleArray(field string, value []float64) {
	_ = o.set(field, value)
}

func (o *RpcObject) SetBool(field string, value bool) {
	_ = o.set(field, value)
}

func (o *RpcObject) SetString(field string, value string) {
	_ = o.set(field, value)
}

func (o *RpcObject) SetBinary(field string, value []byte) {
	_ = o.set(field, base64.StdEncoding.EncodeToString(value))
}

func (o *RpcObject) SetDateTime(s string, value time.Time) {
	_ = o.set(s, value.UTC().Format(time.RFC3339))
}

func (o *RpcObject) SetJson(field string, value interface{}) error {
	j, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_ = o.set(field, string(j))
	return nil
}

func (o *RpcObject) Has(field string) bool {
	var ok bool
	if _, ok = o.Dict[field]; ok {
		return true
	}
	return false
}

func (o *RpcObject) Get(field string, defaultValue string) string {
	return o.GetRaw(field, defaultValue)
}

func (o *RpcObject) GetAll() map[string]string {
	return o.Dict
}

func (o *RpcObject) IsSuccessful() bool {
	return o.Error == m2cp.RpcErrorCodeSuccess
}

func (o *RpcObject) Validate(field string, t m2cp.Type, optional bool) error {
	if !o.Has(field) {
		if optional {
			return nil
		}
		return fmt.Errorf("missing required field %s", field)
	}
	switch t {
	case m2cp.TypeString:
		return nil
	case m2cp.TypeBool:
		return o.validateGet(field, m2cp.TypeBool)
	case m2cp.TypeDateTime:
		return o.validateGet(field, m2cp.TypeDateTime)
	case m2cp.TypeInt:
		return o.validateGet(field, m2cp.TypeInt)
	case m2cp.TypeIntArray:
		return o.validateGet(field, m2cp.TypeIntArray)
	case m2cp.TypeDouble:
		return o.validateGet(field, m2cp.TypeDouble)
	case m2cp.TypeDoubleArray:
		return o.validateGet(field, m2cp.TypeDoubleArray)
	case m2cp.TypeBinary:
		return o.validateGet(field, m2cp.TypeBinary)
	case m2cp.TypeJson:
		// can't validate json without knowing the target type
		return nil
	default:
		return fmt.Errorf("unknown type %s", t)
	}
}

func (o *RpcObject) ValidateAll() error {
	if o.Definitions == nil {
		return nil
	}
	for field, def := range o.Definitions {
		if err := o.Validate(field, def.GetType(), def.IsOptional()); err != nil {
			return err
		}
	}
	return nil
}

func (o *RpcObject) validateGet(field string, expectedType m2cp.Type) error {
	if o.Definitions == nil {
		return nil
	}
	def, ok := o.Definitions[field]
	if !ok {
		return fmt.Errorf("unknown field '%s'", field)
	}

	if def.GetType() != expectedType {
		return fmt.Errorf("getting %s (defined with type %s) as invalid type %s", field, def.GetType(), expectedType)
	}
	return nil
}
