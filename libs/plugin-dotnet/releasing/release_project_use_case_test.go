package releasing

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSetVersion(t *testing.T) {
	if got, ok := SetVersion("<Project>\n  <PropertyGroup>\n    <Version>0.0.0</Version>\n  </PropertyGroup>\n</Project>", "1.2.0"); !ok || !strings.Contains(got, "<Version>1.2.0</Version>") {
		t.Fatalf("%v %s", ok, got)
	}
	if got, ok := SetVersion("<Project>\n  <PropertyGroup>\n    <OutputType>Exe</OutputType>\n  </PropertyGroup>\n</Project>", "1.2.0"); !ok || !strings.Contains(got, "<PropertyGroup>\n    <Version>1.2.0</Version>\n    <OutputType>") {
		t.Fatalf("added:\n%s", got)
	}
}

func TestPublishPacksThenPushes(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "packages", "core"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "packages", "core", "Acme.Core.csproj"), []byte("<Project/>"), 0o644)
	var ran []string
	real := DotnetRunner
	defer func() { DotnetRunner = real }()
	DotnetRunner = func(_ string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	t.Setenv("NUGET_API_KEY", "k")
	core := protocol.Project{Name: "core", Root: "packages/core", Kind: protocol.ProjectPackage, Exports: []string{"Acme.Core"}}
	res, err := Publish(ws, protocol.PublishParams{Project: core, Version: "1.0.0"})
	if err != nil || !res.Published || len(ran) != 2 {
		t.Fatal(err, res, ran)
	}
	if !strings.HasPrefix(ran[0], "dotnet pack Acme.Core.csproj --nologo -c Release -o ") || !strings.HasSuffix(ran[0], "-p:Version=1.0.0") || !strings.Contains(ran[1], "Acme.Core.1.0.0.nupkg --source https://api.nuget.org/v3/index.json --skip-duplicate --api-key k") {
		t.Fatalf("ran:\n%s", strings.Join(ran, "\n"))
	}
}

func TestPublishPushesToTheFeedSmithTomlNames(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "packages", "core"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "packages", "core", "Acme.Core.csproj"), []byte("<Project/>"), 0o644)
	var ran []string
	real := DotnetRunner
	defer func() { DotnetRunner = real }()
	DotnetRunner = func(_ string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	t.Setenv("NUGET_API_KEY", "nuget-org-key")
	feed := "https://pkgs.dev.azure.com/acme/_packaging/libs/nuget/v3/index.json"
	core := protocol.Project{Name: "core", Root: "packages/core", Kind: protocol.ProjectPackage, Exports: []string{"Acme.Core"},
		Registries: &protocol.Registries{Publish: &protocol.Registry{Name: "libs", Auth: protocol.AuthAzureArtifacts, NuGet: feed}}}
	res, err := Publish(ws, protocol.PublishParams{Project: core, Version: "1.0.0"})
	if err != nil || !res.Published || res.Registry != feed {
		t.Fatal(err, res)
	}
	// Azure Artifacts: the credential provider signs the push, the key is a placeholder.
	if !strings.Contains(ran[1], "--source "+feed+" --skip-duplicate --api-key az") {
		t.Fatalf("ran:\n%s", strings.Join(ran, "\n"))
	}
}

func TestPublished(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/acme.core/index.json" {
			_, _ = w.Write([]byte(`{"versions": ["1.0.0", "1.10.0", "1.9.0", "2.0.0-rc.1"]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	NuGetIndex = srv.URL
	core := protocol.Project{Kind: protocol.ProjectPackage, Exports: []string{"Acme.Core"}}
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "1.10.0" {
		t.Fatalf("%+v", res)
	}
}
