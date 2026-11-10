package shiplink

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// workspace makes a workspace with the given files (slash paths to contents).
func workspace(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		write(t, root, name, body)
	}
	return root
}

func write(t *testing.T, root, name, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func sameFile(t *testing.T, root, a, b string) bool {
	t.Helper()
	ia, err := os.Stat(filepath.Join(root, filepath.FromSlash(a)))
	if err != nil {
		t.Fatal(err)
	}
	ib, err := os.Stat(filepath.Join(root, filepath.FromSlash(b)))
	if err != nil {
		t.Fatal(err)
	}
	return os.SameFile(ia, ib)
}

// swap replaces a file-system function for one test.
func swap[T any](t *testing.T, target *T, with T) {
	t.Helper()
	old := *target
	*target = with
	t.Cleanup(func() { *target = old })
}

func TestShipLinksTheToolsFileIntoTheShippedFolder(t *testing.T) {
	ws := workspace(t, map[string]string{"dist/cargo/debug/api": "v1", "dist/cargo/debug/api.d": "deps"})
	r := Ship(ws, "apps/api", []string{"../../dist/cargo/debug/api", "../../dist/cargo/debug/api.exe"})
	if r.Folder != "dist/apps/api" || r.Shipped != 1 || r.Copied != 0 || r.NothingMatched || len(r.InUse)+len(r.Failed) != 0 {
		t.Fatalf("report %+v", r)
	}
	if !sameFile(t, ws, "dist/cargo/debug/api", "dist/apps/api/api") {
		t.Fatal("the shipped file is not a hard link to the tool's file")
	}
	if _, err := os.Stat(filepath.Join(ws, "dist", "apps", "api", "api.d")); !os.IsNotExist(err) {
		t.Fatal("only the named file ships")
	}
}

func TestShipLinksAgainWhenTheToolReplacedTheFile(t *testing.T) {
	ws := workspace(t, map[string]string{"dist/cargo/debug/api": "v1"})
	Ship(ws, "apps/api", []string{"../../dist/cargo/debug/api"})
	// Cargo writes a new file, it does not write into the old one.
	src := filepath.Join(ws, "dist", "cargo", "debug", "api")
	if err := os.Remove(src); err != nil {
		t.Fatal(err)
	}
	write(t, ws, "dist/cargo/debug/api", "v2")
	r := Ship(ws, "apps/api", []string{"../../dist/cargo/debug/api"})
	if r.Shipped != 1 || len(r.InUse) != 0 {
		t.Fatalf("report %+v", r)
	}
	if got := read(t, ws, "dist/apps/api/api"); got != "v2" {
		t.Fatalf("the shipped file holds %q, want the new build", got)
	}
	if !sameFile(t, ws, "dist/cargo/debug/api", "dist/apps/api/api") {
		t.Fatal("the new file is not linked")
	}
}

func TestShipKeepsALinkThatIsStillCurrent(t *testing.T) {
	ws := workspace(t, map[string]string{"dist/cargo/debug/api": "v1"})
	Ship(ws, "apps/api", []string{"../../dist/cargo/debug/api"})
	swap(t, &removeFile, func(string) error { t.Fatal("a current link is not removed"); return nil })
	if r := Ship(ws, "apps/api", []string{"../../dist/cargo/debug/api"}); r.Shipped != 1 {
		t.Fatalf("report %+v", r)
	}
}

func TestShipCopiesWhenTheLinkFails(t *testing.T) {
	ws := workspace(t, map[string]string{"dist/gradle/api/libs/api.jar": "jar"})
	swap(t, &linkFile, func(string, string) error { return errors.New("cross-device link") })
	r := Ship(ws, "apps/api", []string{"../../dist/gradle/api/libs/*.jar"})
	if r.Shipped != 1 || r.Copied != 1 || len(r.Failed) != 0 {
		t.Fatalf("report %+v", r)
	}
	if read(t, ws, "dist/apps/api/api.jar") != "jar" || sameFile(t, ws, "dist/gradle/api/libs/api.jar", "dist/apps/api/api.jar") {
		t.Fatal("want a copy of the jar")
	}
}

func TestShipKeepsTheOldFileWhenARunningProgramHoldsIt(t *testing.T) {
	ws := workspace(t, map[string]string{"dist/cargo/debug/api.exe": "v2", "dist/apps/api/api.exe": "v1"})
	swap(t, &removeFile, func(string) error {
		return &os.PathError{Op: "remove", Path: "api.exe", Err: errors.New("Access is denied.")}
	})
	r := Ship(ws, "apps/api", []string{"../../dist/cargo/debug/api.exe"})
	if len(r.InUse) != 1 || r.InUse[0] != "dist/apps/api/api.exe" || r.Shipped != 0 || len(r.Failed) != 0 {
		t.Fatalf("report %+v", r)
	}
	if read(t, ws, "dist/apps/api/api.exe") != "v1" {
		t.Fatal("the old file must stay")
	}
}

func TestShipLinksEveryFileOfAFolder(t *testing.T) {
	ws := workspace(t, map[string]string{
		"dist/dotnet/bin/Api/debug/Api.dll":            "dll",
		"dist/dotnet/bin/Api/debug/runtimes/win/x.dll": "native",
	})
	r := Ship(ws, "apps/api", []string{"../../dist/dotnet/bin/Api/debug/*"})
	if r.Shipped != 2 {
		t.Fatalf("report %+v", r)
	}
	if !sameFile(t, ws, "dist/dotnet/bin/Api/debug/runtimes/win/x.dll", "dist/apps/api/runtimes/win/x.dll") {
		t.Fatal("a file of a sub-folder keeps its place below the shipped folder")
	}
}

func TestShipSaysWhenNothingMatched(t *testing.T) {
	ws := workspace(t, nil)
	if r := Ship(ws, "apps/api", []string{"../../dist/cargo/debug/api"}); !r.NothingMatched || r.Shipped != 0 {
		t.Fatalf("report %+v", r)
	}
	if r := Ship(ws, "apps/api", nil); r.NothingMatched {
		t.Fatal("a task without Ship ships nothing and says nothing")
	}
}

func TestShipIgnoresAGlobOutsideTheWorkspace(t *testing.T) {
	ws := workspace(t, nil)
	if r := Ship(ws, "apps/api", []string{"../../../elsewhere/api"}); r.Shipped != 0 || !r.NothingMatched {
		t.Fatalf("report %+v", r)
	}
}
