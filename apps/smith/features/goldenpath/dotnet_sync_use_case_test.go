//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDotnetSync gives two .NET projects different versions of one package:
// `smith sync --check` fails, `smith sync` aligns them in the project files,
// and the check then passes.
func TestDotnetSync(t *testing.T) {
	if _, err := exec.LookPath("dotnet"); err != nil {
		t.Skip("the .NET SDK is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-dotnetsync", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-dotnetsync")
	w.smith(true, "add", "app", "api", "--lang", "dotnet")
	w.smith(true, "add", "package", "core", "--lang", "dotnet")
	for csproj, version := range map[string]string{
		"apps/api/SmithE2eDotnetsync.Api.csproj":       "12.0.3",
		"packages/core/SmithE2eDotnetsync.Core.csproj": "13.0.1",
	} {
		item := "  <ItemGroup>\n    <PackageReference Include=\"Newtonsoft.Json\" Version=\"" + version + "\" />\n  </ItemGroup>\n</Project>"
		w.write(csproj, strings.Replace(w.read(csproj), "</Project>", item, 1))
	}

	out := w.smith(false, "sync", "--check")
	if !strings.Contains(out, "Newtonsoft.Json (dotnet)") || !strings.Contains(out, "13.0.1") {
		t.Fatalf("sync --check names the package and the version it would use:\n%s", out)
	}
	w.smith(true, "sync", "--no-install")
	if got := w.read("apps/api/SmithE2eDotnetsync.Api.csproj"); !strings.Contains(got, `<PackageReference Include="Newtonsoft.Json" Version="13.0.1" />`) {
		t.Fatalf("api declares the aligned version:\n%s", got)
	}
	if out := w.smith(true, "sync", "--check"); !strings.Contains(out, "Every project agrees") {
		t.Fatalf("after sync the check passes:\n%s", out)
	}
}
