package functions

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// A library of the workspace is not on a registry, so the packaging script
// must carry its code into the zip (checked by hand with the built binary).
func TestThePackagingScriptShipsTheLibrariesOfTheWorkspace(t *testing.T) {
	for _, host := range []string{functionhost.Azure, functionhost.AWS, functionhost.GCP} {
		script := Templates(host)["[[.Dir]]/package.py"]
		for _, want := range []string{"def workspace_packages", "direct_url.json", "workspace_packages()"} {
			if !strings.Contains(script, want) {
				t.Errorf("%s: package.py lacks %q", host, want)
			}
		}
	}
}
