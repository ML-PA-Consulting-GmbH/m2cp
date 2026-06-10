package helper

import (
	"fmt"
	gql "m2cpcli/graphql"
	"m2cpcli/structs"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseModelAssertion(assertion string) (*structs.AssertionModel, error) {
	// split at \n\n
	sections := strings.SplitN(assertion, "\n\n", 2)
	var assertionModel = structs.AssertionModel{}
	if err := yaml.Unmarshal([]byte(sections[0]), &assertionModel); err != nil {
		return nil, err
	}
	return &assertionModel, nil
}

// FindAssertion finds a specific assertion in a list of assertions by type and search string
func FindAssertion(assertions []*gql.AssertionPlain, assertionType string, searchString string) (string, error) {
	typeString := fmt.Sprintf("type: %s", assertionType)
	for _, assertion := range assertions {
		if strings.Contains(assertion.Assertion, typeString) && strings.Contains(assertion.Assertion, searchString) {
			return assertion.Assertion, nil
		}
	}
	return "", fmt.Errorf("assertion not found for type %s and search string %s", assertionType, searchString)
}

// GetAssertionFieldValue returns the value of a field in an assertion
func GetAssertionFieldValue(assertion string, fieldName string) (string, error) {
	// Field starts with 'fieldName : ' and ends with '\n'. We have to find the first occurrence of the field name and then return it
	fieldStart := fmt.Sprintf("%s: ", fieldName)
	fieldEnd := "\n"
	startIndex := strings.Index(assertion, fieldStart)
	if startIndex == -1 {
		return "", fmt.Errorf("field %s not found in assertion", fieldName)
	}
	startIndex += len(fieldStart)
	endIndex := strings.Index(assertion[startIndex:], fieldEnd)
	if endIndex == -1 {
		return "", fmt.Errorf("field %s not found in assertion", fieldName)
	}
	return assertion[startIndex : startIndex+endIndex], nil
}

// GetAssertionFieldValues returns the values of all fields with a specific name in an assertion
func GetAssertionFieldValues(assertion, fieldName string) ([]string, error) {
	var fieldValues []string
	fieldStart := fmt.Sprintf("%s: ", fieldName)
	fieldEnd := "\n"
	for {
		startIndex := strings.Index(assertion, fieldStart)
		if startIndex == -1 {
			break
		}
		startIndex += len(fieldStart)
		endIndex := strings.Index(assertion[startIndex:], fieldEnd)
		if endIndex == -1 {
			return nil, fmt.Errorf("field %s not found in assertion", fieldName)
		}
		fieldValues = append(fieldValues, assertion[startIndex:startIndex+endIndex])
		assertion = assertion[startIndex+endIndex:]
	}
	return fieldValues, nil
}

// FormatAssertions formats the assertions into a single string, parsable by snapstore/snapd
func FormatAssertions(assertions []*gql.AssertionPlain) string {
	var formattedAssertions string
	for i, assertion := range assertions {
		if i < len(assertions)-1 {
			formattedAssertions += fmt.Sprintf("%s\n", assertion.Assertion)
		} else {
			formattedAssertions += assertion.Assertion
		}
	}
	return formattedAssertions
}
