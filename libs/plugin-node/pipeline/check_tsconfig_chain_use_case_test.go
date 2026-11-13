package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func chainWorkspace(t *testing.T, files map[string]string) string {
	t.Helper()
	ws := t.TempDir()
	for name, body := range files {
		p := filepath.Join(ws, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func TestCheckTsconfigChainPassesAProjectThatReachesTheBase(t *testing.T) {
	ws := chainWorkspace(t, map[string]string{
		tsconfigchain.Base:         tsconfigchain.BaseFile,
		"apps/web/tsconfig.json":   `{"extends": "../../tsconfig.base.json"}`,
		"apps/web/package.json":    `{}`,
		"apps/solo/tsconfig.json":  `{"files": [], "references": [{"path": "./tsconfig.lib.json"}]}`,
		"apps/solo/package.json":   `{}`,
		"apps/other/tsconfig.json": `{"extends": "./tsconfig.lib.json"}`,
	})
	if f, ok := CheckTsconfigChain(ws, protocol.Project{Name: "web", Root: "apps/web"}); !ok || !f.OK {
		t.Errorf("a project that extends the base passes: %+v", f)
	}
	if _, ok := CheckTsconfigChain(ws, protocol.Project{Name: "solo", Root: "apps/solo"}); ok {
		t.Error("a solution file is not judged")
	}
	if _, ok := CheckTsconfigChain(ws, protocol.Project{Name: "none", Root: "apps/none"}); ok {
		t.Error("a project without a tsconfig.json is not judged")
	}
}

func TestCheckTsconfigChainWarnsAndAsksBeforeItEdits(t *testing.T) {
	ws := chainWorkspace(t, map[string]string{
		tsconfigchain.Base:       tsconfigchain.BaseFile,
		"apps/fn/tsconfig.json":  "{\n  \"extends\": \"@monosmith/tsconfig/node.json\",\n  \"include\": [\"src\"]\n}\n",
		"apps/web/tsconfig.json": "{\r\n  \"compilerOptions\": { \"strict\": true }\r\n}\r\n",
	})
	f, ok := CheckTsconfigChain(ws, protocol.Project{Name: "fn", Root: "apps/fn"})
	if !ok || f.OK || !f.Warning || f.Fix == nil || f.Fix.Safe {
		t.Fatalf("a project that skips the base is a warning with a fix that asks: %+v", f)
	}
	if len(f.Fix.Choices) != 1 || len(f.Fix.Choices[0].Options) != 2 || len(f.Fix.Files) != 0 {
		t.Fatalf("extend or leave, nothing written unasked: %+v", f.Fix)
	}
	extend, leave := f.Fix.Choices[0].Options[0].Fix, f.Fix.Choices[0].Options[1].Fix
	if len(leave.Files) != 0 || len(extend.Files) != 1 || extend.Files[0].Path != "apps/fn/tsconfig.json" ||
		!strings.Contains(extend.Files[0].Content, `"extends": "../../tsconfig.base.json"`) || strings.Contains(extend.Files[0].Content, "@monosmith/tsconfig") {
		t.Errorf("the edit: %+v", extend.Files)
	}
	f, _ = CheckTsconfigChain(ws, protocol.Project{Name: "web", Root: "apps/web"})
	if f.Fix == nil || !strings.Contains(f.Fix.Choices[0].Options[0].Fix.Files[0].Content, "\r\n") {
		t.Errorf("the file keeps its line endings: %+v", f.Fix)
	}
}

func TestCheckTsconfigChainIsSilentWithoutABase(t *testing.T) {
	ws := chainWorkspace(t, map[string]string{"apps/web/tsconfig.json": `{}`})
	if _, ok := CheckTsconfigChain(ws, protocol.Project{Name: "web", Root: "apps/web"}); ok {
		t.Error("with no base there is nothing to reach")
	}
}
