package updatenotice

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/directlink"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ForThisProcess are the options for the smith that is running: its home, its
// binary, and whether CI or the person's setting turn the notices off. The
// setting is SMITH_UPDATE_CHECK=on|off (a user setting, which wins), or
// [workspace] update_check of the workspace the command runs in; the older
// SMITH_NO_UPDATE_CHECK=1 still means off.
func ForThisProcess(version string, wait time.Duration) Options {
	exe, _ := os.Executable()
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	// Behind a link of the run folder, how smith was installed is read from the program the link was made from.
	exe = directlink.InstalledExe(plugintrust.Home(), exe)
	setting := SettingFrom(os.Getenv("SMITH_UPDATE_CHECK"), workspaceSetting())
	return Options{
		Home: plugintrust.Home(), Version: version, Exe: exe, GOOS: runtime.GOOS,
		Off:      plugintrust.InCI(os.Getenv),
		NoCheck:  setting == "off" || (os.Getenv("SMITH_NO_UPDATE_CHECK") != "" && os.Getenv("SMITH_NO_UPDATE_CHECK") != "0"),
		Announce: setting == "",
		Wait:     wait,
	}
}

// SettingFrom picks the update_check choice: the user's (the environment) wins
// over the workspace's. Anything but "on" or "off" is no choice.
//
// Parameters: env is SMITH_UPDATE_CHECK; file is [workspace] update_check.
// Returns: "on", "off", or "" when the person has not chosen.
func SettingFrom(env, file string) string {
	for _, v := range []string{env, file} {
		switch v = strings.ToLower(strings.Trim(strings.TrimSpace(v), `"'`)); v {
		case "on", "off":
			return v
		}
	}
	return ""
}

// workspaceSetting reads [workspace] update_check from the smith.toml above
// the current directory, as written; "" when there is none.
func workspaceSetting() string {
	root, err := workspace.FindRoot(".")
	if err != nil {
		return ""
	}
	raw, err := os.ReadFile(filepath.Join(root, workspace.FileName))
	if err != nil {
		return ""
	}
	v, _ := configedit.GetKey(string(raw), "workspace", "update_check")
	return v
}
