package updatenotice

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
)

// ForThisProcess are the options for the smith that is running: its home, its
// binary, and whether CI or SMITH_NO_UPDATE_CHECK=1 turn the notices off.
func ForThisProcess(version string, wait time.Duration) Options {
	exe, _ := os.Executable()
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return Options{
		Home: plugintrust.Home(), Version: version, Exe: exe, GOOS: runtime.GOOS,
		Off:     plugintrust.InCI(os.Getenv),
		NoCheck: os.Getenv("SMITH_NO_UPDATE_CHECK") != "" && os.Getenv("SMITH_NO_UPDATE_CHECK") != "0",
		Wait:    wait,
	}
}
