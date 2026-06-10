package env

import (
	"context"
	"fmt"
)

type RuntimeEnvironment struct {
	ProgramName string
	Url         string
	Ctx         context.Context
	cancel      context.CancelFunc
	Commands    map[string]Runner // used for the `help` function
	//Gql         *gql.BackendClient
}

func NewRuntimeEnvironment(command string) (*RuntimeEnvironment, error) {
	ctx, cancel := context.WithCancel(context.Background())
	env := RuntimeEnvironment{
		ProgramName: command,
		Ctx:         ctx,
		cancel:      cancel,
	}
	return &env, nil
}

// EndpointURL is deprecated
func (e *RuntimeEnvironment) EndpointURL(api string) string {
	fmt.Println("EndpointURL is deprecated.")
	return fmt.Sprintf("%s%s", e.Url, api)
}

func (e *RuntimeEnvironment) PersistedStatePath() string {
	return fmt.Sprintf("~/.%s/state.json", e.ProgramName)
}

func (e *RuntimeEnvironment) RegisterCommands(commandsList []Runner) {
	e.Commands = map[string]Runner{}
	for _, command := range commandsList {
		e.Commands[command.Name()] = command
	}
}

//func (e *RuntimeEnvironment) SetGraphQlBackend(url, jwt string) error {
//	var err error
//	e.Gql, err = gql.NewBackendClient(url, jwt)
//	//e.Gql, err = gql.NewDebugBackendClient() // TODO: DEBUG
//	if err != nil {
//		return fmt.Errorf("could not set backend: %s", err)
//	}
//	return nil
//}
