package projectnaming

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFollowsTellsWhetherANameStartsWithTheNamespace(t *testing.T) {
	ws := protocol.WorkspaceInfo{Name: "shop", Namespace: "acme"}
	for _, c := range []struct {
		eco  Ecosystem
		pkg  string
		want bool
	}{
		{Npm, "@acme/core", true},
		{Npm, "@ACME/core", true},
		{Npm, "@other/core", false},
		{Npm, "core", false},
		{Cargo, "acme-core", true},
		{Cargo, "acmecore", false},
		{Python, "acme_core", true},
		{Maven, "shop-core", false},
		{Dart, "acme_core", true},
		{Gem, "core", false},
		{Composer, "acme/core", true},
		{Composer, "shop/core", false},
		{DotNet, "Acme.Core", true},
		{DotNet, "Shop.Core", false},
		{Swift, "Core", true},
	} {
		if got := Follows(ws, c.eco, c.pkg); got != c.want {
			t.Errorf("Follows(%s, %q) = %v, want %v", c.eco, c.pkg, got, c.want)
		}
	}
	if !Follows(protocol.WorkspaceInfo{}, Cargo, "core") {
		t.Error("a workspace with no namespace follows nothing")
	}
}

func TestEcosystemOfNamesTheLanguagesWithAPrefix(t *testing.T) {
	if eco, ok := EcosystemOf("php"); !ok || eco != Composer {
		t.Errorf("php: %v %v", eco, ok)
	}
	if _, ok := EcosystemOf("go"); ok {
		t.Error("go has no namespace prefix in a package name")
	}
}
