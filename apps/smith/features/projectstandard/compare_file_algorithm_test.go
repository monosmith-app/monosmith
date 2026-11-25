package projectstandard

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const exportsValue = `{".": {"source": "./src/index.ts", "types": "./dist/index.d.ts", "default": "./dist/index.js"}}`

func keysStandard() protocol.StandardFileContent {
	return protocol.StandardFileContent{
		Project: "models", Path: "libs/models/package.json", Mode: protocol.StandardKeys, Format: "json",
		Keys:   map[string]json.RawMessage{"exports": json.RawMessage(exportsValue)},
		Reason: "Ctrl+click opens the source.",
	}
}

func TestCompare(t *testing.T) {
	var root string
	beforeEach := func() { root = t.TempDir() }

	t.Run("a whole file that differs is replaced and kept as .pre-smith", func(t *testing.T) {
		beforeEach()
		write(t, root, "libs/a/tsconfig.json", `{"compilerOptions": {"outDir": "x"}}`)
		write(t, root, "libs/a/tsconfig.json.pre-smith", "an earlier copy")
		f := protocol.StandardFileContent{Project: "a", Path: "libs/a/tsconfig.json", Mode: protocol.StandardFile, Content: "{}\n"}
		d, differs, err := Compare(root, f)
		if err != nil || !differs {
			t.Fatalf("differs %v err %v", differs, err)
		}
		if d.Copy != "libs/a/tsconfig.json.pre-smith-2" || string(d.New) != "{}\n" || strings.Contains(string(d.New), "Generated") {
			t.Fatalf("an earlier copy is never overwritten and no header is written: %+v", d)
		}
	})

	t.Run("a whole file as the standard has it is no difference", func(t *testing.T) {
		beforeEach()
		write(t, root, "libs/a/tsconfig.json", "{}\r\n")
		f := protocol.StandardFileContent{Project: "a", Path: "libs/a/tsconfig.json", Mode: protocol.StandardFile, Content: "{}\n"}
		if _, differs, err := Compare(root, f); err != nil || differs {
			t.Fatalf("line endings are not a difference: %v %v", differs, err)
		}
	})

	t.Run("keys are set in place, keeping their order and every other key", func(t *testing.T) {
		beforeEach()
		write(t, root, "libs/models/package.json", "{\n  \"name\": \"@quotes/models\",\n  \"main\": \"./dist/index.js\"\n}\n")
		d, differs, err := Compare(root, keysStandard())
		if err != nil || !differs || d.Copy != "" {
			t.Fatalf("differs %v err %v copy %q", differs, err, d.Copy)
		}
		got := string(d.New)
		if !strings.Contains(got, `"name": "@quotes/models"`) || strings.Index(got, `"source"`) > strings.Index(got, `"default"`) {
			t.Fatalf("other keys stay and the conditions keep their order:\n%s", got)
		}
		write(t, root, "libs/models/package.json", got)
		if _, differs, err := Compare(root, keysStandard()); err != nil || differs {
			t.Fatalf("once written it is the standard: %v %v", differs, err)
		}
	})

	t.Run("an exports map in another order differs", func(t *testing.T) {
		beforeEach()
		write(t, root, "libs/models/package.json", `{"exports": {".": {"default": "./dist/index.js", "types": "./dist/index.d.ts", "source": "./src/index.ts"}}}`)
		if _, differs, err := Compare(root, keysStandard()); err != nil || !differs {
			t.Fatalf("order is meaning in exports: %v %v", differs, err)
		}
	})

	t.Run("a path outside the workspace is refused", func(t *testing.T) {
		beforeEach()
		f := protocol.StandardFileContent{Path: "../elsewhere.json", Mode: protocol.StandardFile, Content: "{}"}
		if _, _, err := Compare(root, f); err == nil {
			t.Fatal("want a refusal")
		}
	})
}

func TestFindingsAndApply(t *testing.T) {
	root := t.TempDir()
	write(t, root, "libs/models/package.json", `{"name": "@quotes/models", "types": "./dist/index.d.ts"}`)
	write(t, root, "libs/models/tsconfig.json", `{"compilerOptions": {"outDir": "x"}}`)
	projects := []protocol.Project{{Name: "models", Root: "libs/models"}, {Name: "web", Root: "apps/web"}}
	res := protocol.StandardResult{Files: []protocol.StandardFileContent{
		keysStandard(),
		{Project: "models", Path: "libs/models/tsconfig.json", Mode: protocol.StandardFile, Content: "{}\n", Reason: "One tsconfig shape."},
		{Project: "stranger", Path: "libs/stranger/tsconfig.json", Mode: protocol.StandardFile, Content: "{}\n"},
	}}
	a, err := Collect(root, projects, res)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Differences) != 2 {
		t.Fatalf("a project not asked about is left out: %+v", a.Differences)
	}

	t.Run("one warning per project, with a fix that keeps the copy", func(t *testing.T) {
		fs := Findings(a)
		if len(fs) != 1 {
			t.Fatalf("web has no standard file, so no finding: %+v", fs)
		}
		f := fs[0]
		if f.OK || !f.Warning || f.Check != "libs/models"+CheckSuffix || f.Fix == nil || f.Fix.Safe {
			t.Fatalf("finding %+v", f)
		}
		if len(f.Fix.Files) != 3 || f.Fix.Files[1].Path != "libs/models/tsconfig.json.pre-smith" || !strings.Contains(f.Remedy, "[[doctor.accept]]") {
			t.Fatalf("fix %+v remedy %q", f.Fix.Files, f.Remedy)
		}
	})

	t.Run("apply writes the copy and the standard", func(t *testing.T) {
		written, err := Apply(root, a.Differences, io.Discard)
		if err != nil || len(written) != 3 {
			t.Fatalf("%v %v", written, err)
		}
		if read(t, root, "libs/models/tsconfig.json.pre-smith") != `{"compilerOptions": {"outDir": "x"}}` || read(t, root, "libs/models/tsconfig.json") != "{}\n" {
			t.Fatal("the copy holds the old file and the file the standard")
		}
		again, err := Collect(root, projects, res)
		if err != nil || len(again.Differences) != 0 || !Findings(again)[0].OK {
			t.Fatalf("applied, the project passes: %+v %v", again.Differences, err)
		}
	})
}

func TestParseAnswer(t *testing.T) {
	for _, v := range []string{"", Replace, Keep} {
		if got, err := ParseAnswer(v); err != nil || got != v {
			t.Errorf("%q: %q %v", v, got, err)
		}
	}
	if _, err := ParseAnswer("both"); err == nil {
		t.Error("want a refusal")
	}
}
