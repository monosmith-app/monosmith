package toolpin

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, root, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// native is a native pin file for each label, with the version it names.
var native = map[string]struct{ file, body, want string }{
	"node":   {".nvmrc", "v22.3.0\n", "22.3.0"},
	"go":     {"go.work", "go 1.27.0\n\nuse .\n", "1.27"},
	"python": {".python-version", "3.13\n", "3.13"},
	"rustc":  {"rust-toolchain.toml", "[toolchain]\nchannel = \"1.99\"\n", "1.99"},
	"java":   {".java-version", "17.0.2\n", "17"},
	"dotnet": {"global.json", `{"sdk":{"version":"9.0.300"}}`, "9.0"},
	"ruby":   {".ruby-version", "ruby-3.3.1\n", "3.3.1"},
	"php":    {".php-version", "8.3\n", "8.3"},
}

// TestEveryLanguageDecidesInOneOrder is the conformance table: for each
// language, smith.toml beats the native file, which beats the default.
func TestEveryLanguageDecidesInOneOrder(t *testing.T) {
	for label, n := range native {
		t.Run(label, func(t *testing.T) {
			ws := t.TempDir()
			if got := Version(ws, label); got != Default(label) || got == "" {
				t.Fatalf("no pin: want default %q, got %q", Default(label), got)
			}
			write(t, ws, n.file, n.body)
			if got := Version(ws, label); got != n.want {
				t.Fatalf("native file: want %q, got %q", n.want, got)
			}
			write(t, ws, "smith.toml", "[workspace]\nname = \"x\"\n\n[toolchains]\n# pin\n"+label+" = \"7.7\"\n\n[other]\n"+label+" = \"1\"\n")
			if got := Version(ws, label); got != "7.7" {
				t.Fatalf("smith.toml: want 7.7, got %q", got)
			}
		})
	}
}

func TestExtraCandidatesComeAfterTheNativeFile(t *testing.T) {
	ws := t.TempDir()
	if got := Version(ws, "go", "", "1.25"); got != "1.25" {
		t.Fatalf("extra: got %q", got)
	}
	write(t, ws, "go.work", "go 1.26\n")
	if got := Version(ws, "go", "1.25"); got != "1.26" {
		t.Fatalf("native: got %q", got)
	}
}

func TestNativeFilesThatNameNoVersionAreSkipped(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, ".nvmrc", "lts/*\n")
	write(t, ws, ".node-version", "20\n")
	if got := Version(ws, "node"); got != "20" {
		t.Fatalf("got %q", got)
	}
	ws = t.TempDir()
	write(t, ws, "Cargo.toml", "[workspace.package]\nrust-version = \"1.88\"\n")
	if got := Version(ws, "rustc"); got != "1.88" {
		t.Fatalf("cargo: got %q", got)
	}
	ws = t.TempDir()
	write(t, ws, "pom.xml", "<maven.compiler.release>17</maven.compiler.release>")
	if got := Version(ws, "java"); got != "17" {
		t.Fatalf("pom: got %q", got)
	}
}
