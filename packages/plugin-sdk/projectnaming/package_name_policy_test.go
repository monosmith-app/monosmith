package projectnaming

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestNamespaceIsTheNamespaceElseTheName(t *testing.T) {
	for _, c := range []struct {
		ws   protocol.WorkspaceInfo
		want string
	}{
		{protocol.WorkspaceInfo{Name: "opencraw", Namespace: "pargic"}, "pargic"},
		{protocol.WorkspaceInfo{Name: "opencraw", Namespace: " "}, "opencraw"},
		{protocol.WorkspaceInfo{Name: "opencraw", Scope: "@pargic"}, "opencraw"},
		{protocol.WorkspaceInfo{}, ""},
	} {
		if got := Namespace(c.ws); got != c.want {
			t.Errorf("%+v: got %q, want %q", c.ws, got, c.want)
		}
	}
}

func TestScopeIsExplicitElseTheNamespace(t *testing.T) {
	for _, c := range []struct {
		ws   protocol.WorkspaceInfo
		want string
	}{
		{protocol.WorkspaceInfo{Name: "x", Scope: "@pargic"}, "@pargic"},
		{protocol.WorkspaceInfo{Name: "x", Scope: "pargic"}, "@pargic"},
		{protocol.WorkspaceInfo{Name: "x", Namespace: "My Shop!"}, "@my-shop"},
		{protocol.WorkspaceInfo{Name: "Acme"}, "@acme"},
		{protocol.WorkspaceInfo{}, ""},
	} {
		if got := Scope(c.ws); got != c.want {
			t.Errorf("%+v: got %q, want %q", c.ws, got, c.want)
		}
	}
}

func TestNameFollowsTheEcosystemSyntax(t *testing.T) {
	ws := func(ns string) protocol.WorkspaceInfo { return protocol.WorkspaceInfo{Name: "ignored", Namespace: ns} }
	for _, c := range []struct {
		eco    Ecosystem
		ns, in string
		want   string
	}{
		{Npm, "acme", "core", "@acme/core"},
		{Npm, "", "core", "@ignored/core"},
		{Cargo, "acme", "core", "acme-core"},
		{Cargo, "jato.auto.packages", "rust-core", "jato-auto-packages-rust-core"},
		{Cargo, "My Shop!", "api", "my-shop-api"},
		{Cargo, "acme", "acme-core", "acme-core"},
		{Cargo, "2024.shop", "api", "x2024-shop-api"},
		{Python, "acme", "sdk", "acme-sdk"},
		{Python, "acme", "acme-sdk", "acme-sdk"},
		{Maven, "acme", "core", "acme-core"},
		{Dart, "acme", "core", "acme_core"},
		{Dart, "acme", "acme-core", "acme_core"},
		{Dart, "3d", "web--api", "x3d_web_api"},
		{CMake, "acme", "acme_core", "acme_core"},
		{Gem, "acme", "web-api", "acme_web_api"},
		{Composer, "Acme Corp", "web-api", "acme-corp/web-api"},
		{DotNet, "acme", "web-api", "Acme.WebApi"},
		{DotNet, "jato.auto.packages", "core", "JatoAutoPackages.Core"},
		{DotNet, "acme", "acme-core", "AcmeCore"},
		{Swift, "acme", "sales-report", "SalesReport"},
		{Swift, "acme", "2fa", "M2fa"},
		{ObjC, "acme", "string-kit", "StringKit"},
		{VBA, "acme", "sales-report", "sales_report"},
		{VBA, "acme", "2fa", "M2fa"},
		{Lua, "acme", "core", "core"},
	} {
		if got := Name(ws(c.ns), c.eco, c.in); got != c.want {
			t.Errorf("Name(%q, %s, %q) = %q, want %q", c.ns, c.eco, c.in, got, c.want)
		}
	}
}

func TestNameWithoutANamespaceIsTheProjectName(t *testing.T) {
	none := protocol.WorkspaceInfo{}
	for eco, want := range map[Ecosystem]string{Cargo: "core", Python: "core", Dart: "core", Gem: "core", DotNet: "Core", Composer: "core/core", Npm: "core"} {
		if got := Name(none, eco, "core"); got != want {
			t.Errorf("%s: got %q, want %q", eco, got, want)
		}
	}
}

func TestGoModulePrefixIsAValidPath(t *testing.T) {
	for in, want := range map[string]string{
		"acme":               "example.com/acme",
		"Jato.Auto.Packages": "example.com/jato.auto.packages",
		"My Shop!":           "example.com/my-shop",
		"":                   "example.com/workspace",
	} {
		if got := GoModulePrefix(protocol.WorkspaceInfo{Namespace: in}); got != want {
			t.Errorf("GoModulePrefix(%q) = %q, want %q", in, got, want)
		}
	}
	if got := GoModulePrefix(protocol.WorkspaceInfo{Name: "a", ModulePrefix: "github.com/x/y"}); got != "github.com/x/y" {
		t.Errorf("an explicit prefix wins: %q", got)
	}
}

func TestGroupIDIsDevAndTheNamespace(t *testing.T) {
	if got := GroupID(protocol.WorkspaceInfo{Name: "x", Namespace: "My-Shop"}); got != "dev.myshop" {
		t.Errorf("got %q", got)
	}
}
