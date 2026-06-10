package m2cp

type Type int

var Types = []string{"string", "bool", "datetime", "int", "[]int", "double", "[]double", "binary", "json"}

const (
	TypeString Type = iota // ioto: auto increment value
	TypeBool
	TypeDateTime
	TypeInt
	TypeIntArray
	TypeDouble
	TypeDoubleArray
	TypeBinary
	TypeJson
	TypeInvalid
)

func NewType(s string) Type {
	switch s {
	case "string":
		return TypeString
	case "bool":
		return TypeBool
	case "datetime":
		return TypeDateTime
	case "int":
		return TypeInt
	case "[]int":
		return TypeIntArray
	case "double":
		return TypeDouble
	case "[]double":
		return TypeDoubleArray
	case "binary":
		return TypeBinary
	case "json":
		return TypeJson
	default:
		return TypeInvalid
	}
}

func (m Type) String() string {
	switch m {
	case TypeString:
		return "string"
	case TypeBool:
		return "bool"
	case TypeDateTime:
		return "datetime"
	case TypeInt:
		return "int"
	case TypeIntArray:
		return "[]int"
	case TypeDouble:
		return "double"
	case TypeDoubleArray:
		return "[]double"
	case TypeBinary:
		return "binary"
	case TypeJson:
		return "json"
	default:
		return "invalid"
	}
}
