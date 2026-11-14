package functionhost

import (
	"archive/zip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func write(t *testing.T, file, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func zipNames(t *testing.T, file string) []string {
	t.Helper()
	zr, err := zip.OpenReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	slices.Sort(names)
	return names
}

func TestAssembleFunctionAppCopiesTheManifestFilesThatExist(t *testing.T) {
	project, build := t.TempDir(), filepath.Join(t.TempDir(), "dist", "app")
	write(t, filepath.Join(project, "package.json"), `{"name":"app"}`)
	write(t, filepath.Join(project, "host.json"), `{"version":"2.0"}`)
	if err := AssembleFunctionApp(project, build); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"package.json", "host.json"} {
		if _, err := os.Stat(filepath.Join(build, name)); err != nil {
			t.Errorf("%s was not copied: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(build, "local.settings.json")); err == nil {
		t.Error("a file the project lacks is not made")
	}
}

func TestPackFunctionAppZipsTheBuildWithoutDependenciesForAZipHost(t *testing.T) {
	project, build := t.TempDir(), filepath.Join(t.TempDir(), "out")
	write(t, filepath.Join(build, "dist", "functions", "a.js"), "x")
	write(t, filepath.Join(build, "package.json"), `{"name":"app","dependencies":{"left-pad":"1"}}`)
	write(t, filepath.Join(build, "host.json"), `{}`)
	var ran [][]string
	tools := Tools{Run: func(dir string, argv []string) error { ran = append(ran, argv); return nil }}
	zipPath, err := PackFunctionApp(PackOptions{ProjectDir: project, BuildDir: build, Name: "app", Tools: tools})
	if err != nil {
		t.Fatal(err)
	}
	if got := zipNames(t, zipPath); !slices.Equal(got, []string{"dist/functions/a.js", "host.json", "package.json"}) {
		t.Errorf("zip holds %v", got)
	}
	if len(ran) != 1 || strings.Join(ran[0], " ") != "npm run build" {
		t.Errorf("only the build runs for a zip host: %v", ran)
	}
	if _, err := os.Stat(filepath.Join(build, "stage")); err == nil {
		t.Error("the stage folder is removed")
	}
}

func TestPackFunctionAppInstallsDependenciesWhenAsked(t *testing.T) {
	project, build := t.TempDir(), filepath.Join(t.TempDir(), "out")
	write(t, filepath.Join(build, "package.json"), `{"name":"app","dependencies":{"left-pad":"1"}}`)
	var ran []string
	tools := Tools{Run: func(dir string, argv []string) error {
		ran = append(ran, strings.Join(argv[:2], " "))
		if argv[1] == "install" {
			write(t, filepath.Join(dir, "node_modules", "left-pad", "index.js"), "x")
		}
		return nil
	}}
	zipPath, err := PackFunctionApp(PackOptions{ProjectDir: project, BuildDir: build, Name: "app", WithDependencies: true, Tools: tools})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(zipNames(t, zipPath), "node_modules/left-pad/index.js") {
		t.Errorf("the installed dependency is in the zip: %v", zipNames(t, zipPath))
	}
	if !slices.Equal(ran, []string{"npm run", "npm install"}) {
		t.Errorf("ran %v", ran)
	}
}

func TestPackFunctionAppVendorsALibraryOfTheWorkspace(t *testing.T) {
	ws := t.TempDir()
	lib := filepath.Join(ws, "libs", "greeter")
	write(t, filepath.Join(lib, "package.json"), `{"name":"@acme/greeter","main":"index.js","scripts":{"build":"x"}}`)
	write(t, filepath.Join(lib, "index.js"), "module.exports = 1")
	project := filepath.Join(ws, "apps", "fn")
	write(t, filepath.Join(project, "package.json"), `{"name":"fn"}`)
	if err := os.MkdirAll(filepath.Join(ws, "node_modules", "@acme"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(lib, filepath.Join(ws, "node_modules", "@acme", "greeter")); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	build := filepath.Join(ws, "dist", "apps", "fn")
	write(t, filepath.Join(build, "package.json"), `{"name":"fn","dependencies":{"@acme/greeter":"*"}}`)
	var packed []string
	tools := Tools{
		Run: func(string, []string) error { return nil },
		Output: func(dir string, argv []string) (string, error) {
			packed = argv
			return "acme-greeter-1.0.0.tgz\n", nil
		},
	}
	if _, err := PackFunctionApp(PackOptions{ProjectDir: project, BuildDir: build, Name: "fn", Tools: tools}); err != nil {
		t.Fatal(err)
	}
	if len(packed) < 3 || packed[1] != "pack" || packed[2] != "./acme-greeter" {
		t.Errorf("the library is packed as a tarball: %v", packed)
	}
}

func TestPackFunctionAppInstallsWithTheWorkspaceNpmrcAndNeverZipsIt(t *testing.T) {
	ws := t.TempDir()
	write(t, filepath.Join(ws, "smith.toml"), "")
	write(t, filepath.Join(ws, ".npmrc"), "@acme:registry=https://feed.example/npm/\n//feed.example/npm/:_authToken=${FEED_TOKEN}\n")
	project := filepath.Join(ws, "apps", "api")
	write(t, filepath.Join(project, "package.json"), `{"name":"api"}`)
	build := filepath.Join(ws, "dist", "apps", "api")
	write(t, filepath.Join(build, "package.json"), `{"name":"api","dependencies":{"@acme/core":"1"}}`)
	var seen string
	tools := Tools{Run: func(dir string, argv []string) error {
		if argv[1] == "install" {
			raw, err := os.ReadFile(filepath.Join(dir, ".npmrc"))
			if err != nil {
				t.Errorf("the install ran without the workspace .npmrc: %v", err)
			}
			seen = string(raw)
		}
		return nil
	}}
	zipPath, err := PackFunctionApp(PackOptions{ProjectDir: project, BuildDir: build, Name: "api", WithDependencies: true, Tools: tools})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(seen, "@acme:registry=https://feed.example/npm/") {
		t.Errorf("npm saw %q", seen)
	}
	for _, name := range zipNames(t, zipPath) {
		if strings.HasSuffix(name, ".npmrc") {
			t.Errorf("%s is in the zip: a token line must never ship", name)
		}
	}
}

func TestNearestNpmrcStopsAtTheWorkspaceRoot(t *testing.T) {
	outer := t.TempDir()
	write(t, filepath.Join(outer, ".npmrc"), "outside the workspace\n")
	ws := filepath.Join(outer, "ws")
	write(t, filepath.Join(ws, "smith.toml"), "")
	project := filepath.Join(ws, "apps", "api")
	write(t, filepath.Join(project, "package.json"), "{}")
	if got := nearestNpmrc(project); got != "" {
		t.Errorf("found %s above the workspace root", got)
	}
	write(t, filepath.Join(project, ".npmrc"), "own\n")
	if got := nearestNpmrc(project); got != filepath.Join(project, ".npmrc") {
		t.Errorf("the project's own .npmrc was not first: %q", got)
	}
}
