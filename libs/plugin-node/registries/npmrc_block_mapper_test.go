package registries

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var azure = &protocol.Registry{Name: "libs", Auth: protocol.AuthAzureArtifacts, NPM: "https://pkgs.dev.azure.com/acme/_packaging/libs/npm/registry/"}

func TestBlockForEachSignIn(t *testing.T) {
	cases := map[string]struct {
		r     *protocol.Registries
		scope string
		want  string
	}{
		"azure":   {&protocol.Registries{Install: azure, Publish: azure}, "@acme", "registry=https://pkgs.dev.azure.com/acme/_packaging/libs/npm/registry/\n"},
		"github":  {&protocol.Registries{Install: &protocol.Registry{Auth: protocol.AuthGitHubPackages, NPM: "https://npm.pkg.github.com"}}, "@acme", "@acme:registry=https://npm.pkg.github.com\n//npm.pkg.github.com/:_authToken=${GITHUB_TOKEN}\n"},
		"token":   {&protocol.Registries{Install: &protocol.Registry{Auth: protocol.AuthToken, TokenEnv: "FEED", NPM: "https://npm.example.com/repo/"}}, "", "registry=https://npm.example.com/repo/\n//npm.example.com/repo/:_authToken=${FEED}\n"},
		"publish": {&protocol.Registries{Publish: &protocol.Registry{Auth: protocol.AuthToken, TokenEnv: "PUB", NPM: "https://npm.example.com/"}}, "", "//npm.example.com/:_authToken=${PUB}\n"},
	}
	for name, c := range cases {
		got := Block(c.r, c.scope)
		if got != BlockStart+"\n"+c.want+BlockEnd+"\n" {
			t.Errorf("%s: %q", name, got)
		}
	}
	if Block(nil, "") != "" || Block(&protocol.Registries{Install: &protocol.Registry{Auth: protocol.AuthToken, NuGet: "https://n/"}}, "") != "" {
		t.Error("no npm registry: no block")
	}
}

func TestMergeKeepsTheTeamsLines(t *testing.T) {
	team := "//registry.npmjs.org/:_authToken=${NODE_AUTH_TOKEN}\nsave-exact=true\n"
	block := Block(&protocol.Registries{Install: azure}, "")
	merged := Merge(team, block)
	if !strings.HasPrefix(merged, block) || !strings.HasSuffix(merged, team) {
		t.Fatalf("merged = %q", merged)
	}
	// Rewriting replaces the block in place; removing it leaves the rest.
	other := Block(&protocol.Registries{Install: &protocol.Registry{Auth: protocol.AuthToken, TokenEnv: "T", NPM: "https://x/"}}, "")
	if got := Merge(merged, other); got != other+team {
		t.Errorf("rewritten = %q", got)
	}
	if got := Merge(merged, ""); got != team {
		t.Errorf("removed = %q", got)
	}
	if Merge(team, "") != team {
		t.Error("no block, nothing to remove")
	}
}

func TestCheckNpmrcFixesOnlySmithsBlock(t *testing.T) {
	root := t.TempDir()
	team := "save-exact=true\n"
	_ = os.WriteFile(filepath.Join(root, ".npmrc"), []byte(team), 0o644)
	projects := []protocol.Project{{Name: "a", Exports: []string{"@acme/a"}, Registries: &protocol.Registries{Install: azure, Publish: azure}}}
	f, ok := CheckNpmrc(root, projects)
	if !ok || f.OK || f.Fix == nil || !f.Fix.Safe {
		t.Fatalf("finding = %+v", f)
	}
	if got := f.Fix.Files[0].Content; !strings.Contains(got, "registry="+azure.NPM) || !strings.HasSuffix(got, team) {
		t.Errorf("fix = %q", got)
	}
	_ = os.WriteFile(filepath.Join(root, ".npmrc"), []byte(f.Fix.Files[0].Content), 0o644)
	if f, _ := CheckNpmrc(root, projects); !f.OK {
		t.Errorf("after the fix: %+v", f)
	}
	// No registry named and no block: nothing to check.
	if _, ok := CheckNpmrc(t.TempDir(), []protocol.Project{{Name: "a"}}); ok {
		t.Error("checked a workspace with no registries")
	}
}

func TestCISignsInTheProviderNativeWay(t *testing.T) {
	gh := &protocol.Registry{Auth: protocol.AuthGitHubPackages, NPM: "https://npm.pkg.github.com"}
	tok := &protocol.Registry{Auth: protocol.AuthToken, TokenEnv: "FEED", NPM: "https://f/"}
	r, public := CI([]protocol.Project{{Kind: protocol.ProjectPackage, Registries: &protocol.Registries{Install: azure, Publish: azure}}})
	if public || len(r.Authenticate) != 1 || r.Authenticate[0].Azure.Uses != "npmAuthenticate@0" || len(r.Credentials) != 0 {
		t.Errorf("azure: %+v public=%v", r, public)
	}
	r, _ = CI([]protocol.Project{{Kind: protocol.ProjectPackage, Registries: &protocol.Registries{Install: gh, Publish: gh}}})
	if len(r.Credentials) != 1 || r.Credentials[0].Env != "GITHUB_TOKEN" || !r.Credentials[0].Setup || r.Permissions[0] != "packages: write" {
		t.Errorf("github: %+v", r)
	}
	r, public = CI([]protocol.Project{{Kind: protocol.ProjectLib, Registries: &protocol.Registries{Install: tok, Publish: tok}}, {Kind: protocol.ProjectPackage}})
	if !public || len(r.Credentials) != 1 || r.Credentials[0].Env != "FEED" || !r.Credentials[0].Setup {
		t.Errorf("token: %+v public=%v", r, public)
	}
}
