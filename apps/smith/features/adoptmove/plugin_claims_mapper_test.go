package adoptmove

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestClaimsOfKeepsThePluginsThatDiscoverByManifest(t *testing.T) {
	got := ClaimsOf([]protocol.InitializeResult{
		{Languages: []string{"php"}, Manifests: []string{"composer.json"}, Capabilities: protocol.Capabilities{Discover: true}},
		{Languages: []string{"docker"}, Capabilities: protocol.Capabilities{Discover: true}},
		{Languages: []string{"go"}, Manifests: []string{"go.mod"}},
		{Languages: []string{"dotnet", "csharp"}, Manifests: []string{"*.csproj"}, Ignores: []string{"bin/", "obj/"}, Capabilities: protocol.Capabilities{Discover: true}},
	})
	want := []Claim{{Language: "php", Manifests: []string{"composer.json"}}, {Language: "dotnet", Manifests: []string{"*.csproj"}, Ignores: []string{"bin/", "obj/"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
