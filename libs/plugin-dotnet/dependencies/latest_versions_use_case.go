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

// FlatContainer is nuget.org's package-content API; tests replace FetchJSON.
const FlatContainer = "https://api.nuget.org/v3-flatcontainer"

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

// newestStable is the last version without a pre-release label in a list
// nuget.org keeps in ascending order.
func newestStable(versions []string) string {
	for i := len(versions) - 1; i >= 0; i-- {
		if !strings.Contains(versions[i], "-") {
			return versions[i]
		}
	}
	return ""
}

// Latest asks nuget.org for the newest stable release of each package, a few
// at a time. A package it does not have (a private feed's) is in neither
// list; one it cannot be asked about is Unreachable.
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
			var index struct {
				Versions []string `json:"versions"`
			}
			err := FetchJSON(FlatContainer+"/"+strings.ToLower(name)+"/index.json", &index)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil && strings.Contains(err.Error(), "404"):
			case err != nil:
				res.Unreachable[name] = err.Error()
			default:
				if v := newestStable(index.Versions); v != "" {
					res.Latest[name] = v
				}
			}
		}(name)
	}
	wg.Wait()
	return res, nil
}
