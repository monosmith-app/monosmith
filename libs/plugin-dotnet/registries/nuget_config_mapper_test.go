package registries

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var azure = &protocol.Registry{Name: "libs", Auth: protocol.AuthAzureArtifacts, NuGet: "https://pkgs.dev.azure.com/acme/_packaging/libs/nuget/v3/index.json"}

func TestConfigForAzureArtifactsListsTheFeedAlone(t *testing.T) {
	got := Config(&protocol.Registries{Install: azure, Publish: azure})
	if !strings.Contains(got, `<add key="libs" value="`+azure.NuGet+`" />`) || strings.Contains(got, "nuget.org") || strings.Contains(got, "packageSourceCredentials") {
		t.Errorf("config:\n%s", got)
	}
	var doc struct{}
	if err := xml.Unmarshal([]byte(got), &doc); err != nil {
		t.Errorf("not XML: %v", err)
	}
}

func TestConfigForATokenFeedKeepsNuGetOrgAndReadsTheToken(t *testing.T) {
	tok := &protocol.Registry{Name: "feed", Auth: protocol.AuthToken, TokenEnv: "FEED_TOKEN", NuGet: "https://nuget.example.com/v3/index.json?a=1&b=2"}
	got := Config(&protocol.Registries{Install: tok})
	for _, want := range []string{`value="https://nuget.example.com/v3/index.json?a=1&amp;b=2"`, `<add key="nuget.org" value="https://api.nuget.org/v3/index.json" />`, `<add key="ClearTextPassword" value="%FEED_TOKEN%" />`} {
		if !strings.Contains(got, want) {
			t.Errorf("lacks %s:\n%s", want, got)
		}
	}
	if Config(nil) != "" || Config(&protocol.Registries{Install: &protocol.Registry{Name: "n", NPM: "https://n/"}}) != "" {
		t.Error("no NuGet feed: no config")
	}
}

func TestCheckRewritesSmithsConfigButNotTheTeams(t *testing.T) {
	projects := []protocol.Project{{Registries: &protocol.Registries{Install: azure, Publish: azure}}}
	root := t.TempDir()
	f, ok := CheckNuGetConfig(root, projects)
	if !ok || f.OK || !f.Fix.Safe {
		t.Fatalf("missing: %+v", f)
	}
	_ = os.WriteFile(filepath.Join(root, FileName), []byte(f.Fix.Files[0].Content), 0o644)
	if f, _ := CheckNuGetConfig(root, projects); !f.OK {
		t.Errorf("after the fix: %+v", f)
	}
	_ = os.WriteFile(filepath.Join(root, FileName), []byte("<configuration />\n"), 0o644)
	if f, _ := CheckNuGetConfig(root, projects); f.OK || f.Fix == nil || f.Fix.Safe {
		t.Errorf("the team's config must not be rewritten safely: %+v", f)
	}
	if _, ok := CheckNuGetConfig(root, []protocol.Project{{}}); ok {
		t.Error("no feed named and a team's config: nothing to check")
	}
}

func TestCIAndPushSignInPerFeed(t *testing.T) {
	r, public := CI([]protocol.Project{{Kind: protocol.ProjectPackage, Registries: &protocol.Registries{Install: azure, Publish: azure}}})
	if public || len(r.Authenticate) != 1 || r.Authenticate[0].Azure.Uses != "NuGetAuthenticate@1" || len(r.Credentials) != 0 {
		t.Errorf("azure: %+v public=%v", r, public)
	}
	source, key, ok := Push(protocol.Project{Registries: &protocol.Registries{Publish: azure}}, func(string) string { return "" })
	if !ok || source != azure.NuGet || key != "az" {
		t.Errorf("azure push: %s %s %v", source, key, ok)
	}
	gh := &protocol.Registry{Name: "gh", Auth: protocol.AuthGitHubPackages, NuGet: "https://nuget.pkg.github.com/acme/index.json"}
	_, key, _ = Push(protocol.Project{Registries: &protocol.Registries{Publish: gh}}, func(k string) string { return map[string]string{"GITHUB_TOKEN": "gh-token"}[k] })
	if key != "gh-token" {
		t.Errorf("github push key %q", key)
	}
	if _, _, ok := Push(protocol.Project{}, os.Getenv); ok {
		t.Error("no feed named: nuget.org, as before")
	}
}
