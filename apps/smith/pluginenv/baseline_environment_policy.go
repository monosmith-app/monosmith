package pluginenv

import "strings"

// commonBaseline are the variables every external plugin gets on every
// platform: enough to find programs, a home and a temporary directory, to
// speak the user's language and to reach the network through a proxy.
var commonBaseline = []string{
	"PATH", "HOME", "USERPROFILE", "TMPDIR", "TEMP", "TMP", "TZ",
	"LANG", "LANGUAGE", "LC_*", "TERM",
	"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY",
	"SSL_CERT_FILE", "SSL_CERT_DIR",
	// What smith itself defines for a plugin.
	"SMITH_PLUGIN", "SMITH_VERBOSE",
}

// windowsBaseline are the extra variables a Windows process needs to start
// at all (the system directory, the shell, the user's profile folders).
var windowsBaseline = []string{
	"SYSTEMROOT", "SYSTEMDRIVE", "WINDIR", "COMSPEC", "PATHEXT",
	"APPDATA", "LOCALAPPDATA", "PROGRAMDATA", "ALLUSERSPROFILE", "PUBLIC",
	"PROGRAMFILES", "PROGRAMFILES(X86)", "PROGRAMW6432",
	"COMMONPROGRAMFILES", "COMMONPROGRAMFILES(X86)", "COMMONPROGRAMW6432",
	"HOMEDRIVE", "HOMEPATH", "USERNAME", "USERDOMAIN", "COMPUTERNAME",
	"NUMBER_OF_PROCESSORS", "PROCESSOR_ARCHITECTURE", "OS",
}

// Baseline lists the variable names (or prefix patterns ending in *) an
// external plugin always gets on goos, when they are set.
func Baseline(goos string) []string {
	out := append([]string(nil), commonBaseline...)
	if goos == "windows" {
		out = append(out, windowsBaseline...)
	}
	return out
}

// IsSmithSecret reports whether a variable is one smith itself uses for a
// credential or a release: the cache signing key and token, the release
// settings, and any other SMITH_ variable that looks like a secret. A plugin's
// own manifest can never ask for these; only the workspace owner can, in
// smith.toml.
func IsSmithSecret(name string) bool {
	upper := strings.ToUpper(name)
	if !strings.HasPrefix(upper, "SMITH_") {
		return false
	}
	if strings.HasPrefix(upper, "SMITH_RELEASE_") || strings.HasPrefix(upper, "SMITH_CACHE_") {
		return true
	}
	for _, word := range []string{"KEY", "TOKEN", "SECRET", "PASSWORD", "CREDENTIAL"} {
		if strings.Contains(upper, word) {
			return true
		}
	}
	return false
}
