package cachekey

import (
	"errors"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
)

// UserEnvironment keeps SMITH_CACHE_KEY for the user of this machine. On
// Windows it is a user environment variable that PowerShell sets. Elsewhere it
// is a marked block in the profile of the shell, which smith writes only after
// the person confirmed the file and the lines.
type UserEnvironment struct {
	// GOOS is runtime.GOOS.
	GOOS string
	// Run starts a program with text on its standard input.
	Run CommandRunner
	// Machine reads and writes the profile files. Nil means smith writes no
	// profile and only shows the line to add.
	Machine shellenv.Environment
}

const (
	setUserVariable   = "[Environment]::SetEnvironmentVariable('SMITH_CACHE_KEY', %s, 'User')"
	unsetUserVariable = "[Environment]::SetEnvironmentVariable('SMITH_CACHE_KEY', $null, 'User')"
)

// CanSet is true where a command sets the variable: Windows.
func (u UserEnvironment) CanSet() bool { return u.GOOS == "windows" }

// Profile is the profile file that would hold the key.
//
// Returns: the file, and false on Windows, without a machine, or when the
// shell is not zsh, bash or fish.
func (u UserEnvironment) Profile() (ProfileTarget, bool) {
	if u.CanSet() || u.Machine == nil {
		return ProfileTarget{}, false
	}
	return TargetFor(u.Machine)
}

// Preview is the exact command smith would run on Windows, or on other
// systems the line to add to a shell profile, with a placeholder where the key
// goes: the key is never shown.
func (u UserEnvironment) Preview() string {
	if u.CanSet() {
		return `powershell -NoProfile -Command "` + strings.Replace(setUserVariable, "%s", "'<the key>'", 1) + `"`
	}
	if t, ok := u.Profile(); ok {
		return t.Line("<the key>")
	}
	return ProfileTarget{Shell: shellenv.Bash}.Line("<the key>")
}

// Set stores the key as a user environment variable on Windows. The key goes
// in through the program's standard input, so it never appears in a process
// list. PowerShell tells the open windows about the change.
//
// Errors: this system has no command for it, or the program failed. No error
// holds the key.
func (u UserEnvironment) Set(key string) error {
	if !u.CanSet() {
		return errors.New("this system has no command that sets a user variable")
	}
	_, err := u.Run(key+"\n", "powershell", "-NoProfile", "-Command", strings.Replace(setUserVariable, "%s", "[Console]::In.ReadLine()", 1))
	if err != nil {
		return errors.New(scrub(err, key))
	}
	return nil
}

// Unset removes the user environment variable on Windows.
//
// Errors: this system has no command for it, or the program failed.
func (u UserEnvironment) Unset() error {
	if !u.CanSet() {
		return errors.New("this system has no command that sets a user variable")
	}
	_, err := u.Run("", "powershell", "-NoProfile", "-Command", unsetUserVariable)
	return err
}
