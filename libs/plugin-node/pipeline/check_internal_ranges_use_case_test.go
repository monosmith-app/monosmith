package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeManifest(t *testing.T, ws, dir, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(ws, dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, dir, "package.json"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckInternalRangeFindingsWarnsAndOffersOneGroupedFix(t *testing.T) {
	ws := t.TempDir()
	writeManifest(t, ws, ".", `{"workspaces":["packages/*"]}`)
	writeManifest(t, ws, "packages/env", `{"name":"@auto/env","version":"0.0.2"}`)
	writeManifest(t, ws, "packages/retry", `{"name":"@auto/retry"}`)
	writeManifest(t, ws, "packages/app", "{\n    \"name\": \"@auto/app\",\n    \"dependencies\": {\n        \"@auto/env\": \"^0.0.1\",\n        \"@auto/retry\": \"*\"\n    }\n}\n")
	writeManifest(t, ws, "packages/other", "{\n  \"name\": \"@auto/other\",\n  \"devDependencies\": {\n    \"@auto/retry\": \"~1.0.0\"\n  }\n}\n")
	_ = os.WriteFile(filepath.Join(ws, "package-lock.json"), []byte(`{}`), 0o644)
	projects := []protocol.Project{
		{Name: "env", Root: "packages/env"}, {Name: "retry", Root: "packages/retry"},
		{Name: "app", Root: "packages/app"}, {Name: "other", Root: "packages/other"},
	}

	findings := CheckInternalRangeFindings(ws, projects)
	if len(findings) != 2 {
		t.Fatalf("want a warning for app and for other, got %+v", findings)
	}
	app := findings[0]
	if app.OK || !app.Warning || app.Project != "app" || app.Area != "dependencies" || app.Check != CheckInternalRanges {
		t.Fatalf("finding %+v", app)
	}
	if !strings.Contains(app.Problem, "@auto/env ^0.0.1") || strings.Contains(app.Problem, "@auto/retry") {
		t.Errorf("problem %q", app.Problem)
	}
	if !strings.Contains(app.Remedy, "registry copy") {
		t.Errorf("remedy %q", app.Remedy)
	}
	fix := app.Fix
	if fix == nil || len(fix.Files) != 1 || fix.Files[0].Path != "packages/app/package.json" {
		t.Fatalf("fix %+v", fix)
	}
	want := "{\n    \"name\": \"@auto/app\",\n    \"dependencies\": {\n        \"@auto/env\": \"*\",\n        \"@auto/retry\": \"*\"\n    }\n}\n"
	if fix.Files[0].Content != want {
		t.Errorf("content %q", fix.Files[0].Content)
	}
	if len(fix.Steps) != 1 || strings.Join(fix.Steps[0].Argv, " ") != "npm install --package-lock-only --no-audit --no-fund" {
		t.Errorf("steps %+v", fix.Steps)
	}
	if findings[1].Fix == nil || findings[1].Fix.Description != fix.Description || findings[1].Fix.AppliesTo[0] != "other" {
		t.Errorf("the fixes must share a description so the host folds them: %+v", findings[1].Fix)
	}
}

func TestCheckInternalRangeFindingsPassesAndHasNoStepWithoutLockfile(t *testing.T) {
	ws := t.TempDir()
	writeManifest(t, ws, "packages/env", `{"name":"@auto/env"}`)
	writeManifest(t, ws, "packages/app", `{"name":"@auto/app","dependencies":{"@auto/env":"workspace:*","left-pad":"^1.0.0"},"peerDependencies":{"@auto/env":"^1.0.0"}}`)
	projects := []protocol.Project{{Name: "env", Root: "packages/env"}, {Name: "app", Root: "packages/app"}}
	got := CheckInternalRangeFindings(ws, projects)
	if len(got) != 1 || !got[0].OK || got[0].Check != CheckInternalRanges {
		t.Fatalf("want one passing finding, got %+v", got)
	}

	writeManifest(t, ws, "packages/app", `{"name":"@auto/app","dependencies":{"@auto/env":"1.0.0"}}`)
	got = CheckInternalRangeFindings(ws, projects)
	if len(got) != 1 || got[0].OK || got[0].Fix == nil || len(got[0].Fix.Steps) != 0 {
		t.Fatalf("no lockfile, no step: %+v", got)
	}
}
