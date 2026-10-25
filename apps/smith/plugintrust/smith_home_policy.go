package plugintrust

import (
	"os"
	"path/filepath"
)

// Home is smith's per-user directory: SMITH_HOME, or ~/.smith. Installed
// plugins and trust decisions live under it.
func Home() string {
	if h := os.Getenv("SMITH_HOME"); h != "" {
		return h
	}
	if u, err := os.UserHomeDir(); err == nil {
		return filepath.Join(u, ".smith")
	}
	return ""
}

// StorePath is the file trust decisions are kept in.
func StorePath() string {
	home := Home()
	if home == "" {
		return ""
	}
	return filepath.Join(home, "trust.json")
}
