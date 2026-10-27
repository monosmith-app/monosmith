package dependencies

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const rootManifest = `[workspace]
members = ["crates/api", "crates/cli"]

[workspace.dependencies]
anyhow = "1.0.80"
tokio = { version = "1.30", features = ["full"] }
`

const apiManifest = `[package]
name = "api"
version = "0.1.0"

[dependencies]
serde = { version = "1.0.150", features = ["derive"] } # wire format
anyhow.workspace = true
tokio = { workspace = true }
core = { path = "../core", version = "0.1.0" }
fork = { git = "https://example.com/fork" }
rand = "~0.8.4"
http_client = { package = "reqwest", version = ">=0.11, <0.13" }

[dev-dependencies]
tempfile = "3"

[build-dependencies.cc]
version = "1.0.80"
`

const cliManifest = `[package]
name = "cli"
version = "0.1.0"

[dependencies]
serde = "1.0.200"
`

func writeWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	for file, body := range map[string]string{
		"Cargo.toml": rootManifest, "crates/api/Cargo.toml": apiManifest, "crates/cli/Cargo.toml": cliManifest,
	} {
		path := filepath.Join(ws, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

var projects = []protocol.Project{{Name: "api", Root: "crates/api"}, {Name: "cli", Root: "crates/cli"}}

func TestDeclaredListsRegistryDependenciesOfEveryManifest(t *testing.T) {
	res, err := Declared(writeWorkspace(t), protocol.DeclaredParams{Projects: projects})
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		Project, Name, Spec, Section string
		Rewritable                   bool
	}
	var got []row
	for _, d := range res.Declarations {
		got = append(got, row{d.Project, d.Name, d.Spec, d.Section, d.Rewritable})
	}
	want := []row{
		{"", "anyhow", "1.0.80", "workspace.dependencies", true},
		{"", "tokio", "1.30", "workspace.dependencies", true},
		{"api", "cc", "1.0.80", "build-dependencies", true},
		{"api", "rand", "~0.8.4", "dependencies", true},
		{"api", "reqwest", ">=0.11, <0.13", "dependencies", false},
		{"api", "serde", "1.0.150", "dependencies", true},
		{"api", "tempfile", "3", "dev-dependencies", true},
		{"cli", "serde", "1.0.200", "dependencies", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("declarations:\n got %+v\nwant %+v", got, want)
	}
}

func TestRewriteChangesOnlyTheVersionString(t *testing.T) {
	ws := writeWorkspace(t)
	declared, _ := Declared(ws, protocol.DeclaredParams{Projects: projects})
	var changes []protocol.Rewrite
	for _, d := range declared.Declarations {
		switch {
		case d.Name == "serde" && d.Project == "api":
			changes = append(changes, protocol.Rewrite{Declaration: d, Spec: "1.0.200"})
		case d.Name == "cc", d.Name == "tokio", d.Name == "tempfile":
			changes = append(changes, protocol.Rewrite{Declaration: d, Spec: "9.9.9"})
		}
	}
	res, err := Rewrite(ws, protocol.RewriteParams{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 2 {
		t.Errorf("files: %v", res.Files)
	}
	api, _ := os.ReadFile(filepath.Join(ws, "crates", "api", "Cargo.toml"))
	for _, want := range []string{
		`serde = { version = "1.0.200", features = ["derive"] } # wire format`,
		`tempfile = "9.9.9"`,
		"[build-dependencies.cc]\nversion = \"9.9.9\"",
		`rand = "~0.8.4"`,
		`http_client = { package = "reqwest", version = ">=0.11, <0.13" }`,
	} {
		if !strings.Contains(string(api), want) {
			t.Errorf("api manifest lacks %q:\n%s", want, api)
		}
	}
	root, _ := os.ReadFile(filepath.Join(ws, "Cargo.toml"))
	if !strings.Contains(string(root), `tokio = { version = "9.9.9", features = ["full"] }`) || !strings.Contains(string(root), `anyhow = "1.0.80"`) {
		t.Errorf("root manifest:\n%s", root)
	}
}

func TestRewriteMatchesARenamedDependencyByItsCrateName(t *testing.T) {
	src := "[dependencies]\nhttp_client = { package = \"reqwest\", version = \"0.11.0\" }\n"
	out, ok := SetSpec(src, "dependencies", "reqwest", "0.12.0")
	if !ok || !strings.Contains(out, `version = "0.12.0"`) {
		t.Fatalf("%v %s", ok, out)
	}
	if _, ok := SetSpec(src, "dependencies", "http_client", "0.12.0"); ok {
		t.Error("the key of a renamed dependency is not its crate name")
	}
}

func TestRewriteRefusesADependencyThatIsGone(t *testing.T) {
	ws := writeWorkspace(t)
	_, err := Rewrite(ws, protocol.RewriteParams{Changes: []protocol.Rewrite{{
		Declaration: protocol.Declaration{File: "crates/cli/Cargo.toml", Name: "left-pad", Section: "dependencies"}, Spec: "1.0.0",
	}}})
	if err == nil || !strings.Contains(err.Error(), "no longer declares left-pad") {
		t.Fatalf("%v", err)
	}
}

func TestLatestAsksCratesIoForEveryName(t *testing.T) {
	real := Lookup
	t.Cleanup(func() { Lookup = real })
	Lookup = func(name string) protocol.PublishedResult {
		switch name {
		case "serde":
			return protocol.PublishedResult{Version: "1.0.228"}
		case "offline":
			return protocol.PublishedResult{Unreachable: true, Detail: "no route"}
		}
		return protocol.PublishedResult{}
	}
	res, err := Latest("", protocol.LatestParams{Names: []string{"serde", "offline", "unknown"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Latest["serde"] != "1.0.228" || len(res.Latest) != 1 || res.Unreachable["offline"] != "no route" || len(res.Unreachable) != 1 {
		t.Fatalf("%+v", res)
	}
}

func TestInstallAndUninstallRunCargo(t *testing.T) {
	var ran []string
	real := Cargo
	t.Cleanup(func() { Cargo = real })
	Cargo = func(dir string, argv []string) error {
		ran = append(ran, strings.Join(argv, " "))
		return nil
	}
	api := protocol.Project{Name: "api", Root: "crates/api"}
	_, _ = Install("ws", protocol.InstallParams{Projects: projects})
	_, _ = Install("ws", protocol.InstallParams{Project: &api, Packages: []string{"serde@1"}})
	_, _ = Install("ws", protocol.InstallParams{Project: &api, Packages: []string{"tempfile"}, Dev: true})
	_, _ = Uninstall("ws", protocol.UninstallParams{Project: api, Packages: []string{"serde"}})
	want := []string{
		"fetch",
		"add --manifest-path crates/api/Cargo.toml serde@1",
		"add --manifest-path crates/api/Cargo.toml --dev tempfile",
		"remove --manifest-path crates/api/Cargo.toml serde",
	}
	if !reflect.DeepEqual(ran, want) {
		t.Fatalf("cargo ran:\n%v", ran)
	}
}
