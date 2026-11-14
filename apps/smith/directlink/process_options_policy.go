package directlink

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
)

// ForThisProcess are the options for the smith that is running: its home, its
// program, the PATH it was started with and the version.
//
// Parameters: version is the running smith.
// Returns: the options.
func ForThisProcess(version string) Options {
	exe, _ := os.Executable()
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return Options{Home: plugintrust.Home(), Exe: exe, Version: version, GOOS: runtime.GOOS}
}

// FirstRunOff is true when the first run must not link: in CI, where nothing on
// the machine is changed, or when the person set SMITH_NO_LINK=1.
//
// Parameters: getenv reads the environment.
// Returns: true when the first run leaves the entries alone.
func FirstRunOff(getenv func(string) string) bool {
	return plugintrust.InCI(getenv) || (getenv("SMITH_NO_LINK") != "" && getenv("SMITH_NO_LINK") != "0")
}
