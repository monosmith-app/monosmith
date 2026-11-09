package dependencies

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-go/pipeline"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestLatestTools(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "go.work"), []byte("go 1.26\n\nuse ./apps/api\n"), 0o644)
	real := FetchJSON
	defer func() { FetchJSON = real }()
	FetchJSON = func(url string, v any) error {
		if strings.Contains(url, "go.dev") {
			return json.Unmarshal([]byte(`[{"version":"go1.27rc1","stable":false},{"version":"go1.26.2","stable":true}]`), v)
		}
		return json.Unmarshal([]byte(`{"Version":"v2.16.0"}`), v)
	}
	res, _ := Latest(ws, protocol.LatestParams{Tools: true})
	if len(res.Tools) != 2 {
		t.Fatalf("%+v", res)
	}
	if g := res.Tools[0]; g.Pinned != "1.26" || g.Latest != "1.26.2" || g.Newer {
		t.Errorf("a patch release is covered by the go.work version: %+v", g)
	}
	if l := res.Tools[1]; l.Pinned != strings.TrimPrefix(pipeline.GolangciLintVersion, "v") || l.Latest != "2.16.0" || !l.Newer {
		t.Errorf("golangci-lint %+v", l)
	}
	if res, _ := Latest(ws, protocol.LatestParams{}); len(res.Tools) != 0 {
		t.Error("tools only when asked")
	}
}
