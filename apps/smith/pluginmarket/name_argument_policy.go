package pluginmarket

import (
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginarchive"
)

// ParseName decides whether an install argument names a marketplace plugin and
// splits "name@version". It is a name when it is a plugin name with an
// optional "@version" and nothing exists at that path: a local archive or
// directory of that name keeps meaning the path. A path, a URL or
// github:owner/repo@tag is never a name.
func ParseName(arg string) (name, version string, ok bool) {
	name, version, hasVersion := strings.Cut(arg, "@")
	if pluginarchive.ValidatePluginName(name) != nil || hasVersion && (version == "" || strings.ContainsAny(version, "/\\:@")) {
		return "", "", false
	}
	if _, err := os.Lstat(arg); err == nil {
		return "", "", false
	}
	return name, version, true
}
