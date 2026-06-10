package tools

func MaybeIntToInt(n *int, defaultValue int) int {
	if n == nil {
		return defaultValue
	}
	return *n
}

func BoolPtr(b bool) *bool {
	return &b
}
