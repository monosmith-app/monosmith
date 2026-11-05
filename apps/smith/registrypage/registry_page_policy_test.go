package registrypage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPublicPages(t *testing.T) {
	for _, c := range []struct{ plugin, name, want string }{
		{"node", "react", "https://www.npmjs.com/package/react"},
		{"node", "@types/node", "https://www.npmjs.com/package/@types/node"},
		{"python", "requests", "https://pypi.org/project/requests/"},
		{"dotnet", "Newtonsoft.Json", "https://www.nuget.org/packages/Newtonsoft.Json"},
		{"rust", "serde", "https://crates.io/crates/serde"},
		{"jvm", "org.slf4j:slf4j-api", "https://central.sonatype.com/artifact/org.slf4j/slf4j-api"},
		{"jvm", "no-colon", ""},
		{"dart", "http", "https://pub.dev/packages/http"},
		{"ruby", "rails", "https://rubygems.org/gems/rails"},
		{"php", "laravel/framework", "https://packagist.org/packages/laravel/framework"},
		{"go", "github.com/spf13/cobra", "https://pkg.go.dev/github.com/spf13/cobra"},
		{"unknown", "x", ""},
		{"node", "", ""},
	} {
		if got := Page(c.plugin, c.name, nil, false); got != c.want {
			t.Errorf("Page(%s, %s) = %q; want %q", c.plugin, c.name, got, c.want)
		}
	}
}

func TestPrivateRegistries(t *testing.T) {
	azure := func(url string) *protocol.Registry { return &protocol.Registry{NPM: url, PyPI: url, NuGet: url} }
	for _, c := range []struct {
		name, plugin, pkg string
		reg               *protocol.Registry
		want              string
	}{
		{"azure with project", "node", "@acme/ui", azure("https://pkgs.dev.azure.com/acme/web/_packaging/main/npm/registry/"), "https://dev.azure.com/acme/web/_artifacts/feed/main/Npm/@acme%2Fui/overview"},
		{"azure org feed", "python", "acme-lib", azure("https://pkgs.dev.azure.com/acme/_packaging/main/pypi/simple/"), "https://dev.azure.com/acme/_artifacts/feed/main/PyPI/acme-lib/overview"},
		{"azure view", "dotnet", "Acme.Core", azure("https://pkgs.dev.azure.com/acme/_packaging/main@Local/nuget/v3/index.json"), "https://dev.azure.com/acme/_artifacts/feed/main/NuGet/Acme.Core/overview"},
		{"other private registry", "node", "x", azure("https://npm.pkg.github.com"), ""},
		{"named public host", "node", "react", azure("https://registry.npmjs.org/"), "https://www.npmjs.com/package/react"},
		{"registry that sets nothing for the language", "python", "requests", &protocol.Registry{NPM: "https://npm.pkg.github.com"}, "https://pypi.org/project/requests/"},
		{"azure without a feed", "node", "x", azure("https://pkgs.dev.azure.com/acme/"), ""},
	} {
		if got := Page(c.plugin, c.pkg, c.reg, false); got != c.want {
			t.Errorf("%s: Page = %q; want %q", c.name, got, c.want)
		}
	}
}

func TestAnAmbientPrivateRegistryHidesThePublicPage(t *testing.T) {
	if got := Page("node", "react", nil, true); got != "" {
		t.Errorf("Page = %q", got)
	}
}

func TestAmbientPrivate(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if AmbientPrivate("node", root) {
		t.Error("no .npmrc is public")
	}
	write(".npmrc", "registry=https://registry.npmjs.org/\n")
	if AmbientPrivate("node", root) {
		t.Error("npmjs is public")
	}
	write(".npmrc", "@acme:registry=https://pkgs.dev.azure.com/acme/_packaging/f/npm/registry/\n")
	if !AmbientPrivate("node", root) {
		t.Error("a scoped private registry is private")
	}
	write("NuGet.Config", `<add key="a" value="https://api.nuget.org/v3/index.json" />`)
	if AmbientPrivate("dotnet", root) {
		t.Error("nuget.org is public")
	}
	write("NuGet.Config", `<add key="a" value="https://example.com/nuget" />`)
	if !AmbientPrivate("dotnet", root) {
		t.Error("another source is private")
	}
	t.Setenv("PIP_INDEX_URL", "https://example.com/simple/")
	if !AmbientPrivate("python", root) {
		t.Error("a pip index is private")
	}
	t.Setenv("PIP_INDEX_URL", "https://pypi.org/simple/")
	if AmbientPrivate("python", root) {
		t.Error("pypi is public")
	}
}
