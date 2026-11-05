package cachekey

import "strings"

// UserEnvironment sets SMITH_CACHE_KEY as an environment variable of the
// user. Only Windows has a command for it; elsewhere the key goes into a shell
// profile, which smith does not edit.
type UserEnvironment struct {
	// GOOS is runtime.GOOS.
	GOOS string
	// Run starts a program with text on its standard input.
	Run CommandRunner
}

const setUserVariable = "[Environment]::SetEnvironmentVariable('SMITH_CACHE_KEY', %s, 'User')"

// CanSet is true where smith can set the variable itself.
func (u UserEnvironment) CanSet() bool { return u.GOOS == "windows" }

// Preview is the exact command smith would run, or on other systems the line
// to add to a shell profile, with a placeholder where the key goes: the key
// is never shown.
func (u UserEnvironment) Preview() string {
	if u.CanSet() {
		return `powershell -NoProfile -Command "` + strings.Replace(setUserVariable, "%s", "'<the key>'", 1) + `"`
	}
	return `export SMITH_CACHE_KEY="<the key>"`
}

// Set stores the key as a user environment variable. The key goes in through
// the program's standard input, so it never appears in a process list.
//
// Errors: this system has no command for it, or the program failed.
func (u UserEnvironment) Set(key string) error {
	_, err := u.Run(key+"\n", "powershell", "-NoProfile", "-Command", strings.Replace(setUserVariable, "%s", "[Console]::In.ReadLine()", 1))
	return err
}
