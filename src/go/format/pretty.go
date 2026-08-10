package format

import (
	"encoding/json"
	"fmt"
	"m2cpcli/graphql"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type OutputFormatter interface {
	String() (string, error)
}

// JsonOutputMode is set to true if the --json flag is used by [m2cpcli/cmd.RootCmd].
// It should be used to decide whether to print output as JSON or not.
var JsonOutputMode bool

func IsJsonMode() bool {
	return JsonOutputMode
}

func AsJsonString(tree any) (string, error) {
	//bytes, err := json.Marshal(query)
	bytes, err := json.MarshalIndent(tree, "", "    ")
	return string(bytes), err
}

func AsYamlString(tree any) (string, error) {
	bytes, err := yaml.Marshal(tree)
	return string(bytes), err
}

//func PrintFormattedString(cmd *cobra.Command, str string) {
//	str = fmt.Sprintf("Task-Id: %s\n%s", graphql.TaskId, str)
//	cmd.SetOut(os.Stdout)
//	cmd.Println(str)
//}

func PrintFormattedOutput[T any](cmd *cobra.Command, originalTree T, formatter func(T) (string, error)) error {
	var str string
	var err error

	type AugmentedTree struct {
		TaskId string `json:"task-id" yaml:"task-id"`
		Output any    `json:"output" yaml:"output"`
	}
	var newTree AugmentedTree
	newTree.Output = originalTree
	newTree.TaskId = graphql.TaskId

	var tree any
	if graphql.TaskId != "" {
		tree = newTree
	} else {
		tree = originalTree
	}

	if IsJsonMode() {
		str, err = AsJsonString(newTree)
		if err != nil {
			return err
		}
	} else {
		if formatter != nil {
			str, err = formatter(originalTree)
			if err != nil {
				return fmt.Errorf("failed to format the printing")
			}

			if graphql.TaskId != "" {
				taskIdStr := fmt.Sprintf("Task-Id: %s", graphql.TaskId)
				cmd.SetOut(os.Stderr)
				cmd.Println(taskIdStr)
				cmd.SetOut(os.Stdout)
			}
		} else {
			str, err = AsYamlString(tree)
		}
		if err != nil {
			return err
		}
	}
	cmd.SetOut(os.Stdout)
	cmd.Println(str)
	return nil
}

func FormattedListFromNonNestedStruct(query any) (string, error) {
	v := reflect.ValueOf(query)
	typeOfTree := v.Type()

	l := NewList()
	for i := 0; i < v.NumField(); i++ {
		fieldKey := typeOfTree.Field(i).Name
		fieldValue := v.Field(i).Interface()
		//fieldType := typeOfTree.Field(i).Type
		//fmt.Printf("key: %s value: %v type: %s\n", fieldKey, fieldValue, fieldType)

		//if fieldType.String() == "*env.PersistedData" {
		//	// TODO: how to access the structs contents?
		//}

		l.Add(fieldKey, fmt.Sprintf("%v", fieldValue))
	}
	return strings.TrimSpace(l.String()), nil
}

// TODO: flattened list?
// TODO: well formatted table?

//func printTypeDetails(item T) error {
//	s := reflect.ValueOf(&item).Elem()
//	if !s.CanAddr() {
//		return fmt.Errorf("`item` is not addressable: must be pointer or interface")
//	}
//
//	typeOfT := s.Type()
//	for i := 0; i < s.NumField(); i++ {
//		f := s.Field(i)
//		fmt.Printf("%d: %s %s = %v\n", i,
//			typeOfT.Field(i).Alias, f.Type(), f.Interface())
//
//		tag, ok := typeOfT.Field(i).Tag.Lookup("my")
//		if ok {
//			fmt.Printf("%d: '%s'\n", i, tag)
//		}
//
//		fmt.Println(typeOfT.Field(i).Tag)
//	}
//	return nil
//}

func Marshal(in interface{}) (out []byte, err error) {
	value := reflect.ValueOf(in)
	inType := value.Type()

	kind := inType.Kind().String()
	fmt.Printf("# kind: %s\n", kind)
	for i := 0; i < value.NumField(); i++ {
		//f := value.Field(i)
		//fmt.Printf("%d: %s %s = %v\n", i,
		//	inType.Field(i).Alias, f.Type(), f.Interface())

		fieldKey := inType.Field(i).Name
		fieldValue := value.Field(i).Interface() //?
		fieldType := inType.Field(i).Type
		fieldTypeKind := inType.Field(i).Type.Kind()
		//fieldTag := inType.Field(i).Tag // whole tag string
		cliTag, ok := inType.Field(i).Tag.Lookup("cli")
		if !ok {
			return nil, fmt.Errorf("could not find cli tag")
		}
		fmt.Printf("# key: %s, value: %v, type: %s (%s), cli: %s\n", fieldKey, fieldValue, fieldType, fieldTypeKind, cliTag)

		tags := strings.Split(cliTag, ",")
		sectionName := tags[0]
		fmt.Println(fmt.Sprintf("%s:", sectionName))
		if len(tags) > 1 {
			for _, tag := range tags[1:] {
				switch tag {
				case "table":
					fmt.Println("# format as a table!")
					fmt.Println(fieldValue)
				case "listing":
					fmt.Println("# format as a listing!")
					//str, err := FormattedListFromNonNestedStruct(fieldValue)
					//if err != nil {
					//	return nil, fmt.Errorf("could format listing: %s", err) // TODO: panic: reflect: call of reflect.Value.NumField on slice Value
					//}
					//fmt.Println(str)
				default:
					fmt.Println("# default format?")
				}
			}
		}

	}
	//return strings.TrimSpace(l.String()), nil

	//switch inType.Kind() {
	////case "string":
	//case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
	//	//fmt.Println(inType.String())
	//case reflect.Struct:
	//	fmt.Println("nice:" + inType.String())
	//default:
	//	//fmt.Print(inType.Kind().String())
	//	//fmt.Println(inType.Alias())
	//	//fmt.Println(inType.String())
	//	//out = []byte(inType.Alias())
	//}
	return out, nil
}
