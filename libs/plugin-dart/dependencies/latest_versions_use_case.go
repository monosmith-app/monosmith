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

// PubAPI is pub.dev's API root; tests replace FetchJSON.
const PubAPI = "https://pub.dev/api"

// FetchJSON reads a JSON document over HTTP; tests replace it.
var FetchJSON = func(url string, v any) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/vnd.pub.v2+json")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// Latest asks pub.dev for the newest stable release of each package, a few
// at a time. A package pub.dev does not have (a private server's) is in
// neither list.
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
			var pkg struct {
				Latest struct {
					Version string `json:"version"`
				} `json:"latest"`
			}
			err := FetchJSON(PubAPI+"/packages/"+name, &pkg)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil && strings.Contains(err.Error(), "404"):
			case err != nil:
				res.Unreachable[name] = err.Error()
			case pkg.Latest.Version != "":
				res.Latest[name] = pkg.Latest.Version
			}
		}(name)
	}
	wg.Wait()
	return res, nil
}
