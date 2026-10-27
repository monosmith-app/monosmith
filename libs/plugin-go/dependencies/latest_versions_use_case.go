package dependencies

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/libs/plugin-go/pipeline"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// FetchJSON reads a JSON document over HTTP; tests replace it.
var FetchJSON = func(url string, v any) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// Latest asks the module proxy for the newest release of each module, and
// compares the tools the Go plugin pins with their newest releases: Go (the
// version go.work names) and golangci-lint (the version CI installs).
func Latest(workspaceRoot string, params protocol.LatestParams) (protocol.LatestResult, error) {
	var res protocol.LatestResult
	if len(params.Names) > 0 {
		res.Latest, res.Unreachable = latestModules(params.Names)
	}
	if !params.Tools {
		return res, nil
	}
	src, _ := os.ReadFile(filepath.Join(workspaceRoot, "go.work"))
	goPin := protocol.ToolPin{Name: "Go", Pinned: pipeline.ParseGoWork(string(src)).GoVersion}
	var releases []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if FetchJSON("https://go.dev/dl/?mode=json", &releases) == nil {
		for _, r := range releases {
			if r.Stable {
				goPin.Latest = strings.TrimPrefix(r.Version, "go")
				break
			}
		}
	}
	goPin.Newer = goPin.Pinned != "" && goPin.Latest != "" && newer(goPin.Latest, goPin.Pinned, 2)
	lint := protocol.ToolPin{Name: "golangci-lint", Pinned: strings.TrimPrefix(pipeline.GolangciLintVersion, "v")}
	var module struct{ Version string }
	if FetchJSON("https://proxy.golang.org/github.com/golangci/golangci-lint/v2/@latest", &module) == nil {
		lint.Latest = strings.TrimPrefix(module.Version, "v")
	}
	lint.Newer = lint.Latest != "" && newer(lint.Latest, lint.Pinned, 3)
	res.Tools = []protocol.ToolPin{goPin, lint}
	return res, nil
}

// newer compares the first parts of two dotted versions.
func newer(a, b string, parts int) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < parts; i++ {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			return x > y
		}
	}
	return false
}
