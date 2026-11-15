package convertplan

import "testing"

func TestNpmrcFeedFindsTheAzureArtifactsFeedAScopeIsRoutedTo(t *testing.T) {
	cases := map[string]struct {
		npmrc, name, url string
	}{
		"a scoped feed": {
			npmrc: "; comments\n@auto:registry=https://pkgs.dev.azure.com/jato-jaas/Automation/_packaging/AUTO/npm/registry/\n",
			name:  "auto", url: "https://pkgs.dev.azure.com/jato-jaas/Automation/_packaging/AUTO/npm/registry/",
		},
		"an organisation-level feed": {
			npmrc: "@acme:registry=https://pkgs.dev.azure.com/acme/_packaging/libs.core/npm/registry/\nalways-auth=true\n",
			name:  "libs-core", url: "https://pkgs.dev.azure.com/acme/_packaging/libs.core/npm/registry/",
		},
		"a comment names nothing":  {npmrc: "; @auto:registry=https://pkgs.dev.azure.com/o/_packaging/AUTO/npm/registry/\n"},
		"npmjs":                    {npmrc: "@auto:registry=https://registry.npmjs.org/\n"},
		"a global registry":        {npmrc: "registry=https://pkgs.dev.azure.com/o/_packaging/AUTO/npm/registry/\n"},
		"not an npm feed":          {npmrc: "@auto:registry=https://pkgs.dev.azure.com/o/_packaging/AUTO/pypi/simple/\n"},
		"no .npmrc":                {npmrc: ""},
		"a token line is no feed":  {npmrc: "//pkgs.dev.azure.com/o/_packaging/AUTO/npm/registry/:_authToken=${TOKEN}\n"},
		"the first feed is chosen": {npmrc: "@a:registry=https://pkgs.dev.azure.com/o/_packaging/ONE/npm/registry/\n@b:registry=https://pkgs.dev.azure.com/o/_packaging/TWO/npm/registry/\n", name: "one", url: "https://pkgs.dev.azure.com/o/_packaging/ONE/npm/registry/"},
	}
	for name, c := range cases {
		gotName, gotURL, ok := NpmrcFeed(c.npmrc)
		if ok != (c.name != "") || gotName != c.name || gotURL != c.url {
			t.Errorf("%s: %q %q %v, want %q %q", name, gotName, gotURL, ok, c.name, c.url)
		}
	}
}
