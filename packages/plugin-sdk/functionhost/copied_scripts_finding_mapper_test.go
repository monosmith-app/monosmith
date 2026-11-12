package functionhost

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// assembleCopy is scripts/assemble.mjs as smith wrote it into function apps.
const assembleCopy = `// Copies the manifest files next to the compiled code in the workspace's dist/.
import { cpSync, existsSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'

const out = '../../dist/apps/api'
mkdirSync(out, { recursive: true })
for (const file of ['package.json', 'host.json']) {
  if (existsSync(file)) cpSync(file, join(out, file))
}
`

const oldManifest = `{
  "name": "api",
  "scripts": {
    "build": "tsc -p tsconfig.build.json && node scripts/assemble.mjs",
    "start": "node scripts/start.mjs",
    "package": "node scripts/package.mjs"
  }
}
`

func oldApp(t *testing.T, assemble string) (string, protocol.Project) {
	t.Helper()
	ws := t.TempDir()
	write(t, filepath.Join(ws, "apps", "api", "package.json"), oldManifest)
	write(t, filepath.Join(ws, "apps", "api", "scripts", "assemble.mjs"), assemble)
	return ws, protocol.Project{Name: "api", Root: "apps/api"}
}

func TestACopyAsSmithWroteItIsRemovedWithASafeFix(t *testing.T) {
	ws, p := oldApp(t, assembleCopy)
	f, ok := CopiedScriptsFinding(ws, p, Azure)
	if !ok || f.Fix == nil || !f.Fix.Safe || len(f.Fix.Choices) != 0 {
		t.Fatalf("a scaffolded copy has a safe fix without a question: %+v", f)
	}
	var manifest string
	var deleted []string
	for _, e := range f.Fix.Files {
		if e.Delete {
			deleted = append(deleted, e.Path)
		} else if strings.HasSuffix(e.Path, "package.json") {
			manifest = e.Content
		}
	}
	if len(deleted) != 1 || deleted[0] != "apps/api/scripts/assemble.mjs" {
		t.Errorf("deleted %v", deleted)
	}
	if !strings.Contains(manifest, "tsc -p tsconfig.build.json && smith func assemble") {
		t.Errorf("the build script runs smith:\n%s", manifest)
	}
	if !strings.Contains(manifest, "node scripts/start.mjs") {
		t.Errorf("a script of a copy that stays is not touched:\n%s", manifest)
	}
}

func TestAnEditedCopyIsKeptUnlessTheOwnerSaysOtherwise(t *testing.T) {
	ws, p := oldApp(t, assembleCopy+"console.log('mine')\n")
	f, ok := CopiedScriptsFinding(ws, p, Azure)
	if !ok || f.Fix == nil || f.Fix.Safe || len(f.Fix.Choices) != 1 {
		t.Fatalf("an edited copy is asked about, never removed silently: %+v", f)
	}
	options := f.Fix.Choices[0].Options
	if len(options[0].Fix.Files) != 0 {
		t.Errorf("the first option keeps the edited copy: %+v", options[0].Fix.Files)
	}
	if len(options[1].Fix.Files) != 2 {
		t.Errorf("the second option removes the copy and edits package.json: %+v", options[1].Fix.Files)
	}
}

func TestAnAppWithoutCopiesHasNoFinding(t *testing.T) {
	ws := t.TempDir()
	write(t, filepath.Join(ws, "apps", "api", "package.json"), `{}`)
	if _, ok := CopiedScriptsFinding(ws, protocol.Project{Name: "api", Root: "apps/api"}, Azure); ok {
		t.Error("nothing to report")
	}
}

func TestFingerprintIgnoresFilledInStringsAndIndentation(t *testing.T) {
	other := strings.ReplaceAll(assembleCopy, "../../dist/apps/api", "../dist/x")
	other = strings.ReplaceAll(other, "  if", "      if")
	if Fingerprint(assembleCopy) != Fingerprint(other) {
		t.Error("the same script with other paths is the same shape")
	}
	if !IsScaffolded(other) {
		t.Error("the scaffolded assemble script is known")
	}
}

func TestTheStartTaskRunsTheToolOfTheHostFromTheBuild(t *testing.T) {
	for _, c := range []struct{ host, trigger, want string }{
		{Azure, "http", "npm run build && func start --script-root ../../dist/apps/api"},
		{AWS, "http", "npm run build && sam local start-api"},
		{AWS, "queue", "npm run build && sam local invoke GreetingFunction --event events/greeting.json"},
		{GCP, "http", "npm run build && cd ../../dist/apps/api && npx functions-framework --target=hello"},
		{GCP, "queue", "npm run build && cd ../../dist/apps/api && npx functions-framework --target=processMessage --signature-type=cloudevent"},
		{Electron, "", "npm run build && cd ../../dist/electron/apps/api/app && npx electron ."},
	} {
		if got := StartTask(c.host, "apps/api", c.trigger); got != c.want {
			t.Errorf("%s %s: %q, want %q", c.host, c.trigger, got, c.want)
		}
	}
	if got := PackageTask(Azure, "apps/api"); got != "smith func package --with-dependencies" {
		t.Errorf("Azure packs its dependencies: %q", got)
	}
	if got := PackageTask(AWS, "apps/api"); got != "smith func package" {
		t.Errorf("a zip host: %q", got)
	}
}

// azureStartCopy is start.py of an Azure Python app, as older smith wrote it.
const azureStartCopy = `"""Starts the Functions host (Azure Functions Core Tools: func) with the workspace venv."""

import os
import subprocess
import sys

os.environ["languageWorkers__python__defaultExecutablePath"] = sys.executable
settings = "local.settings.json"
if not os.path.exists(settings) and os.path.exists("local.settings.sample.json"):
    print("No local.settings.json: copy local.settings.sample.json to it for your settings.")
sys.exit(subprocess.call(["func", "start"], shell=sys.platform == "win32"))
`

func TestAPythonCopyAsSmithWroteItIsRemovedAndAnEditedOneIsAskedAbout(t *testing.T) {
	ws := t.TempDir()
	write(t, filepath.Join(ws, "apps", "api", "start.py"), azureStartCopy)
	write(t, filepath.Join(ws, "apps", "api", "package.py"), "print('mine')\n")
	f, ok := CopiedPythonScriptsFinding(ws, protocol.Project{Name: "api", Root: "apps/api"})
	if !ok || f.Fix == nil || f.Fix.Safe || len(f.Fix.Choices) != 1 {
		t.Fatalf("an edited package.py is asked about: %+v", f)
	}
	options := f.Fix.Choices[0].Options
	if len(options[0].Fix.Files) != 1 || options[0].Fix.Files[0].Path != "apps/api/start.py" || !options[0].Fix.Files[0].Delete {
		t.Errorf("keeping the edited copy still removes the scaffolded one: %+v", options[0].Fix.Files)
	}
	if len(options[1].Fix.Files) != 2 {
		t.Errorf("removing all deletes both: %+v", options[1].Fix.Files)
	}
	only := t.TempDir()
	write(t, filepath.Join(only, "apps", "api", "start.py"), azureStartCopy)
	if f, _ := CopiedPythonScriptsFinding(only, protocol.Project{Name: "api", Root: "apps/api"}); f.Fix == nil || !f.Fix.Safe {
		t.Errorf("a copy as smith wrote it has a safe fix: %+v", f.Fix)
	}
}
