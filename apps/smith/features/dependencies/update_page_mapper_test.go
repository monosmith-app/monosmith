package dependencies

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAnUpdateLinksToTheRegistryOfItsProject(t *testing.T) {
	feed := &protocol.Registries{Install: &protocol.Registry{NPM: "https://pkgs.dev.azure.com/acme/_packaging/main/npm/registry/"}}
	projects := []protocol.Project{{Name: "web", Registries: feed}, {Name: "docs"}}
	for _, c := range []struct {
		name    string
		project string
		ambient bool
		want    string
	}{
		{"private feed", "web", false, "https://dev.azure.com/acme/_artifacts/feed/main/Npm/react/overview"},
		{"public registry", "docs", false, "https://www.npmjs.com/package/react"},
		{"root manifest takes the first feed", "", false, "https://dev.azure.com/acme/_artifacts/feed/main/Npm/react/overview"},
		{"an .npmrc registry hides the link", "docs", true, ""},
	} {
		u := Update{Plugin: "node", Name: "react", Sites: []protocol.Declaration{{Project: c.project}}}
		if got := pageOf(u, projects, c.ambient); got != c.want {
			t.Errorf("%s: %q; want %q", c.name, got, c.want)
		}
	}
}
