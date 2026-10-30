package workspace

import "runtime"

// CommandPlatforms are the names a release with a build command runs it for,
// one runner each.
var CommandPlatforms = []string{"windows-x64", "macos-arm64", "macos-x64", "linux-x64", "linux-arm64"}

// HostCommandPlatform is the CommandPlatforms name of this machine, or "" when
// it is none of them.
func HostCommandPlatform() string {
	names := map[string]string{
		"windows/amd64": "windows-x64", "darwin/arm64": "macos-arm64", "darwin/amd64": "macos-x64",
		"linux/amd64": "linux-x64", "linux/arm64": "linux-arm64",
	}
	return names[runtime.GOOS+"/"+runtime.GOARCH]
}
