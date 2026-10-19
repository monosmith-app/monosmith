package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestWriteGradleVersion(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	read := func(rel string) string {
		raw, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		return string(raw)
	}
	write("libs/core/build.gradle.kts", "group = \"dev.acme\"\nversion = \"0.0.0\"\n")
	write("libs/util/build.gradle", "group = 'dev.acme'\n")
	write("libs/util/gradle.properties", "version=0.1.0\n")
	res, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "libs/core", Exports: []string{":libs:core", "dev.acme:acme-core"}}, Version: "1.0.0"})
	if err != nil || len(res.Files) != 1 || res.Files[0] != "libs/core/build.gradle.kts" || !strings.Contains(read("libs/core/build.gradle.kts"), `version = "1.0.0"`) {
		t.Fatalf("%+v %v", res, err)
	}
	// The version lives in gradle.properties when only that sets one.
	res, err = WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "libs/util"}, Version: "0.2.0"})
	if err != nil || res.Files[0] != "libs/util/gradle.properties" || read("libs/util/gradle.properties") != "version=0.2.0\n" || read("libs/util/build.gradle") != "group = 'dev.acme'\n" {
		t.Fatalf("%+v %v", res, err)
	}
	// A dry run writes nothing.
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "libs/core"}, Version: "2.0.0", DryRun: true}); err != nil || strings.Contains(read("libs/core/build.gradle.kts"), "2.0.0") {
		t.Fatal("a dry run wrote the version")
	}
}

func TestPublishGradle(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "libs", "core"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "libs", "core", "build.gradle.kts"), nil, 0o644)
	real := GradleRunner
	defer func() { GradleRunner = real }()
	var ran []string
	GradleRunner = func(_ string, argv []string, _ ...string) error { ran = argv; return nil }
	t.Setenv("MAVEN_REPOSITORY_URL", "https://maven.example/releases")
	res, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "libs/core", Kind: protocol.ProjectPackage}})
	if err != nil || !res.Published || strings.Join(ran, " ") != "gradle --console=plain -q :libs:core:publishAllPublicationsToSmithRepository" {
		t.Fatalf("%+v %v %v", res, err, ran)
	}
}
