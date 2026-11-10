package shellenv

// Environment is the port to the machine: the file system and what smith
// reads about the shell and the user. The real one is OSEnvironment. A test
// supplies one that works in a temporary folder.
type Environment interface {
	// Getenv reads an environment variable; "" when it is not set.
	Getenv(key string) string
	// OS is the operating system, as runtime.GOOS spells it.
	OS() string
	// HomeDir is the home folder of the user; "" when it is not known.
	HomeDir() string
	// ReadFile reads a file. The error satisfies errors.Is(err, fs.ErrNotExist)
	// when the file is not there.
	ReadFile(path string) ([]byte, error)
	// WriteFile writes a file and creates the folders above it.
	WriteFile(path string, data []byte) error
	// Remove deletes a file; a file that is not there is not an error.
	Remove(path string) error
	// Exists is true when a file or a folder is at the path.
	Exists(path string) bool
	// ParentProgram is the name of the program that started smith ("zsh",
	// "pwsh.exe"); "" when it is not known.
	ParentProgram() string
}
