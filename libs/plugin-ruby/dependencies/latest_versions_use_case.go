package dependencies

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RubyGemsAPI is rubygems.org's API root; tests replace FetchJSON.
const RubyGemsAPI = "https://rubygems.org/api/v1"

// FetchJSON reads a JSON document over HTTP; tests replace it.
var FetchJSON = func(url string, v any) error {
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// Latest asks rubygems.org for the newest release of each gem (never a
// pre-release), a few at a time. A gem it does not have (a private
// server's) is in neither list.
func Latest(_ string, params protocol.LatestParams) (protocol.LatestResult, error) {
	res := protocol.LatestResult{Latest: map[string]string{}, Unreachable: map[string]string{}}
	var mu sync.Mutex
	var wg sync.WaitGroup
	slots := make(chan struct{}, 8)
	for _, name := range params.Names {
		wg.Add(1)
		slots <- struct{}{}
		go func(name string) {
			defer func() { <-slots; wg.Done() }()
			var latest struct {
				Version string `json:"version"`
			}
			err := FetchJSON(RubyGemsAPI+"/versions/"+name+"/latest.json", &latest)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil && strings.Contains(err.Error(), "404"):
			case err != nil:
				res.Unreachable[name] = err.Error()
			case latest.Version != "" && latest.Version != "unknown":
				res.Latest[name] = latest.Version
			}
		}(name)
	}
	wg.Wait()
	return res, nil
}
