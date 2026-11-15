//go:build e2e

package goldenpath

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestJvmSync gives two Maven modules different versions of one dependency:
// `smith sync --check` fails, `smith sync` aligns the poms, and the check
// then passes. Nothing is built, so no Maven install is needed.
func TestJvmSync(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-jvmsync", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-jvmsync")
	w.smith(true, "add", "app", "api", "--lang", "jvm")
	w.smith(true, "add", "package", "core", "--lang", "jvm")
	for pom, version := range map[string]string{"apps/api/pom.xml": "3.12.0", "packages/core/pom.xml": "3.14.0"} {
		dep := "  <dependencies>\n    <dependency>\n      <groupId>org.apache.commons</groupId>\n      <artifactId>commons-lang3</artifactId>\n      <version>" + version + "</version>\n    </dependency>\n"
		w.write(pom, strings.Replace(w.read(pom), "  <dependencies>\n", dep, 1))
	}

	out := w.smith(false, "sync", "--check")
	if !strings.Contains(out, "org.apache.commons:commons-lang3 (jvm)") || !strings.Contains(out, "3.14.0") {
		t.Fatalf("sync --check names the dependency and the version it would use:\n%s", out)
	}
	w.smith(true, "sync")
	if got := w.read("apps/api/pom.xml"); !strings.Contains(got, "<version>3.14.0</version>") || strings.Contains(got, "3.12.0") {
		t.Fatalf("api declares the aligned version:\n%s", got)
	}
	if out := w.smith(true, "sync", "--check"); !strings.Contains(out, "Every project agrees") {
		t.Fatalf("after sync the check passes:\n%s", out)
	}
}
