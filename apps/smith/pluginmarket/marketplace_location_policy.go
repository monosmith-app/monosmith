package pluginmarket

import (
	"os"
	"strings"
)

const (
	// DefaultLocation is the index of the repository monosmith-app/plugins.
	DefaultLocation = "https://raw.githubusercontent.com/monosmith-app/plugins/main/index.json"
	// LocationEnv names the environment variable that overrides the location.
	LocationEnv = "SMITH_MARKETPLACE"
)

// Location is the index to read: SMITH_MARKETPLACE when set, else configured
// (the [plugins] marketplace key of smith.toml), else DefaultLocation. It is
// an https URL, an http URL on this machine, or a file path.
func Location(configured string) string {
	if v := strings.TrimSpace(os.Getenv(LocationEnv)); v != "" {
		return v
	}
	if v := strings.TrimSpace(configured); v != "" {
		return v
	}
	return DefaultLocation
}
