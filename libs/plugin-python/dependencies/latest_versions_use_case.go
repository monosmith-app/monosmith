package dependencies

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/monosmith-app/monosmith/libs/plugin-python/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/registry"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolpin"
)

// Newest asks a simple index for a distribution's newest release; tests
// replace it.
var Newest = releasing.NewestOn

// FetchJSON reads a JSON document over HTTP; tests replace it.
var FetchJSON = registry.GetJSON

// index is where pip installs from: PIP_INDEX_URL, else PyPI.
func index() string {
	if u := os.Getenv("PIP_INDEX_URL"); u != "" {
		return strings.TrimSuffix(u, "/") + "/"
	}
	return "https://pypi.org/simple/"
}

// Latest asks the index pip uses for each distribution's newest release, a
// few at a time, and compares the Python smith pins with the newest release
// series.
func Latest(workspaceRoot string, params protocol.LatestParams) (protocol.LatestResult, error) {
	res := protocol.LatestResult{Latest: map[string]string{}, Unreachable: map[string]string{}}
	idx := index()
	var mu sync.Mutex
	var wg sync.WaitGroup
	slots := make(chan struct{}, 8)
	for _, name := range params.Names {
		wg.Add(1)
		slots <- struct{}{}
		go func(name string) {
			defer func() { <-slots; wg.Done() }()
			v, err := Newest(idx, name)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil:
				res.Unreachable[name] = err.Error()
			case v != "":
				res.Latest[name] = v
			}
		}(name)
	}
	wg.Wait()
	if params.Tools {
		res.Tools = append(res.Tools, pythonRelease(workspaceRoot))
	}
	return res, nil
}

func pythonRelease(workspaceRoot string) protocol.ToolPin {
	pin := protocol.ToolPin{Name: "Python", Pinned: toolpin.Version(workspaceRoot, "python")}
	var cycles []struct {
		Cycle string `json:"cycle"`
	}
	if FetchJSON("https://endoflife.date/api/python.json", &cycles) != nil || len(cycles) == 0 {
		return pin
	}
	pin.Latest = cycles[0].Cycle
	minor := func(v string) int {
		parts := strings.SplitN(v, ".", 3)
		if len(parts) < 2 {
			return 0
		}
		n, _ := strconv.Atoi(parts[1])
		return n
	}
	pin.Newer = minor(pin.Latest) > minor(pin.Pinned)
	return pin
}
