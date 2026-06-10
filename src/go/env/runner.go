package env

type Runner interface {
	Usage()
	Description() string
	Init([]string) error
	Run(env *RuntimeEnvironment, state *PersistedState) error
	Name() string
}
