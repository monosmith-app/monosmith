package shellenv

// tempEnv is an Environment whose variables, operating system and parent
// program are what the test sets.
type tempEnv struct {
	OSEnvironment
	home   string
	vars   map[string]string
	goos   string
	parent string
}

func (e *tempEnv) Getenv(key string) string { return e.vars[key] }
func (e *tempEnv) OS() string               { return e.goos }
func (e *tempEnv) HomeDir() string          { return e.home }
func (e *tempEnv) ParentProgram() string    { return e.parent }
