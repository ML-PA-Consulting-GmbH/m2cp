package helper

import (
	"fmt"
	"m2cpcli/graphql"
	"m2cpcli/tools"
	"regexp"
)

func ValidateRpcInput(rpcInput *graphql.ExecuteRpcInput) error {
	if rpcInput.Address == "" {
		return fmt.Errorf("address missing")
	}
	if tools.IsValidUuid(rpcInput.Address) {
		return fmt.Errorf("address must not be a UUID but a node address")
	}
	if rpcInput.Command == "" {
		return fmt.Errorf("command missing")
	}
	// check basic address format
	match, _ := regexp.MatchString(".+\\..+\\..+", rpcInput.Address)
	if !match {
		return fmt.Errorf("please provide an address in the format '<node>.<app-name>.<device-serial>' or '<node>.<app-name>.<device-name>'")
	}
	return nil
}

// FindFirst returns the index and the first element E in the list, where predicate(E) is true. Else it returns -1, the zero value of E and an error.
func FindFirst[E any](list []E, predicate func(E) bool) (int, E, error) {
	for i, item := range list {
		if predicate(item) {
			return i, item, nil
		}
	}
	return -1, *new(E), fmt.Errorf("no element matching predicate found")
}
