package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// manifestHas fails the test unless the file at rel under ws holds want.
func manifestHas(t *testing.T, ws, rel, want string) {
	t.Helper()
	src, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), want) {
		t.Fatalf("%s does not hold %q:\n%s", rel, want, src)
	}
}

func TestGenerateNamesTheArtifactAfterThePackageName(t *testing.T) {
	for _, stack := range []string{"", KotlinStack} {
		ws := t.TempDir()
		if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "acme"}, Stack: stack, PackageName: "pargic-core"}); err != nil {
			t.Fatal(err)
		}
		if stack == "" {
			manifestHas(t, ws, "libs/core/pom.xml", "<artifactId>pargic-core</artifactId>")
		} else {
			manifestHas(t, ws, "libs/core/build.gradle.kts", "dist/gradle/pargic-core")
		}
	}
}

// The artifact starts with the workspace scope without its @, on Maven and
// on a new Gradle build.
func TestGenerateBuildsTheArtifactFromTheScope(t *testing.T) {
	for _, stack := range []string{"", KotlinStack} {
		ws := t.TempDir()
		if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "opencraw", Namespace: "pargic"}, Stack: stack}); err != nil {
			t.Fatal(err)
		}
		if stack == "" {
			manifestHas(t, ws, "libs/core/pom.xml", "<artifactId>pargic-core</artifactId>")
		} else {
			manifestHas(t, ws, "libs/core/build.gradle.kts", "dist/gradle/pargic-core")
		}
	}
}
