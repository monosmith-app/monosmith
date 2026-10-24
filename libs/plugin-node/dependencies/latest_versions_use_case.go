package dependencies

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/libs/plugin-node/releasing"
	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Lookup asks the registry for a package's latest version; tests replace it.
var Lookup = func(workspaceRoot, name string) protocol.PublishedResult {
	return releasing.Latest(workspaceRoot, name)
}

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

// Latest asks the registry (npm view, so .npmrc applies) for each package's
// latest version, a few at a time, and compares the Node.js smith pins with
// the newest LTS.
func Latest(workspaceRoot string, params protocol.LatestParams) (protocol.LatestResult, error) {
	res := protocol.LatestResult{Latest: map[string]string{}, Unreachable: map[string]string{}}
	var mu sync.Mutex
	var wg sync.WaitGroup
	slots := make(chan struct{}, 8)
	for _, name := range params.Names {
		wg.Add(1)
		slots <- struct{}{}
		go func(name string) {
			defer func() { <-slots; wg.Done() }()
			r := Lookup(workspaceRoot, name)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case r.Unreachable:
				res.Unreachable[name] = r.Detail
			case r.Version != "":
				res.Latest[name] = r.Version
			}
		}(name)
	}
	wg.Wait()
	if params.Tools {
		res.Tools = append(res.Tools, nodeLTS())
	}
	return res, nil
}

func nodeLTS() protocol.ToolPin {
	pin := protocol.ToolPin{Name: "Node.js (LTS)", Pinned: scaffolding.NodeVersion}
	var releases []struct {
		Version string          `json:"version"`
		LTS     json.RawMessage `json:"lts"`
	}
	if FetchJSON("https://nodejs.org/dist/index.json", &releases) != nil {
		return pin
	}
	for _, r := range releases {
		if string(r.LTS) != "false" && r.Version != "" {
			pin.Latest = strings.TrimPrefix(r.Version, "v")
			break
		}
	}
	major := func(v string) int { n, _ := strconv.Atoi(strings.SplitN(v, ".", 2)[0]); return n }
	pin.Newer = pin.Latest != "" && major(pin.Latest) > major(pin.Pinned)
	return pin
}
