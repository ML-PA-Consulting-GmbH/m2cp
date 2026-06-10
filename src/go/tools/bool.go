package tools

func MaybeBoolToBool(maybeBool *bool, ifNil bool) bool {
	if maybeBool == nil {
		return ifNil
	}
	return *maybeBool
}
