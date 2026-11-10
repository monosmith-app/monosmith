package scaffolding

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-jvm/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerateGradle(t *testing.T) {
	ws := t.TempDir()
	read := func(rel string) string {
		raw, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		return string(raw)
	}
	_ = os.WriteFile(filepath.Join(ws, ".gitignore"), []byte("/dist/\n.gradle/"), 0o644)
	info := protocol.WorkspaceInfo{Name: "acme-corp"}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web-api", Directory: "apps/web-api", Workspace: info, Stack: KotlinStack})
	if err != nil || res.Notes[0] != `Depend on it from another module with implementation(project(":apps:web-api")).` {
		t.Fatal(err, res)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core", Workspace: protocol.WorkspaceInfo{Name: "renamed"}, Stack: KotlinStack}); err != nil {
		t.Fatal(err)
	}
	if s := read("settings.gradle.kts"); !strings.Contains(s, `rootProject.name = "acme-corp"`) || !strings.HasSuffix(s, "\ninclude(\":apps:web-api\")\ninclude(\":packages:core\")\n") {
		t.Fatalf("settings:\n%s", s)
	}
	if g := read(".gitignore"); g != "/dist/\n.gradle/\n\n# Gradle\n.kotlin/\n" {
		t.Fatalf("only the missing caches are ignored: %q", g)
	}
	app := read("apps/web-api/build.gradle.kts")
	for _, want := range []string{"    application\n}", `mainClass = "dev.acmecorp.webapi.MainKt"`, `dir("dist/gradle/acme-corp-web-api")`, "JvmTarget.JVM_" + JavaRelease} {
		if !strings.Contains(app, want) {
			t.Errorf("the app's script has %q:\n%s", want, app)
		}
	}
	if strings.Contains(app, "publishing") || strings.Contains(read("packages/core/build.gradle.kts"), "application") {
		t.Error("only a package publishes, only an app runs")
	}
	if _, err := os.Stat(filepath.Join(ws, "apps/web-api/src/main/kotlin/dev/acmecorp/webapi/Main.kt")); err != nil {
		t.Fatal(err)
	}
	found, err := discovery.Discover(ws, []string{"build.gradle.kts", "apps/web-api/build.gradle.kts", "packages/core/build.gradle.kts"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatalf("%v %+v", err, found.Projects)
	}
	api, core := found.Projects[0], found.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Stack != KotlinStack || api.Tasks["lint"].Command[3] != ":apps:web-api:ktlintCheck" || api.Tasks["start"].Command == nil {
		t.Errorf("api %+v", api)
	}
	if core.Kind != protocol.ProjectPackage || core.Version != "0.0.0" || core.Exports[1] != "dev.acmecorp:acme-corp-core" || core.Tasks["build"].Outputs[0] != "../../dist/packages/core" || core.Tasks["build"].Ship[0] != "../../dist/gradle/acme-corp-core/libs/*.jar" {
		t.Errorf("core %+v", core)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "packages/core", Workspace: info, Stack: KotlinStack}); err == nil {
		t.Fatal("an existing directory is refused")
	}
}

func TestGenerateAndroid(t *testing.T) {
	ws := t.TempDir()
	read := func(rel string) string {
		raw, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		return string(raw)
	}
	info := protocol.WorkspaceInfo{Name: "acme"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: info, Stack: KotlinStack}); err != nil {
		t.Fatal(err)
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "mobile", Directory: "apps/mobile", Workspace: info, Stack: AndroidStack})
	if err != nil {
		t.Fatal(err)
	}
	if root := read("build.gradle.kts"); !strings.Contains(root, androidPlugin) || !slices.Contains(res.Files, "build.gradle.kts") {
		t.Fatalf("the root script pins AGP:\n%s", root)
	}
	if s := read("settings.gradle.kts"); !strings.Contains(s, "google()") || !strings.Contains(s, `include(":apps:mobile")`) {
		t.Fatalf("settings:\n%s", s)
	}
	if script := read("apps/mobile/build.gradle.kts"); !strings.Contains(script, `namespace = "dev.acme.mobile"`) || !strings.Contains(script, `dir("dist/gradle/acme-mobile")`) {
		t.Fatalf("script:\n%s", script)
	}
	if read("apps/mobile/src/main/kotlin/dev/acme/mobile/MainActivity.kt") == "" || read("apps/mobile/src/main/AndroidManifest.xml") == "" {
		t.Fatal("the app has a screen and a manifest")
	}
	found, err := discovery.Discover(ws, []string{"build.gradle.kts", "apps/mobile/build.gradle.kts", "libs/core/build.gradle.kts"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatalf("%v %+v", err, found.Projects)
	}
	mobile := found.Projects[0]
	if mobile.Stack != AndroidStack || strings.Join(mobile.Tasks["lint"].Command[3:], " ") != ":apps:mobile:ktlintCheck :apps:mobile:lint" {
		t.Errorf("mobile %+v", mobile.Tasks["lint"])
	}
	if mobile.Tasks["lint"].DependsOn[0] != "test" {
		t.Error("an Android module's tasks run one after another")
	}
	if _, ok := mobile.Tasks["start"]; ok {
		t.Error("an Android app has no start task")
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "ui", Directory: "libs/ui", Workspace: info, Stack: AndroidStack}); err == nil || !strings.Contains(err.Error(), "--stack kotlin") {
		t.Errorf("an Android lib is refused with a hint: %v", err)
	}
}
