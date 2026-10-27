package dependencies

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PackagistMetadata is Packagist's package metadata root; tests replace
// FetchJSON.
const PackagistMetadata = "https://repo.packagist.org/p2"

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

// release is a stable version: dotted numbers, with the v some tags carry.
var release = regexp.MustCompile(`^v?\d+(\.\d+)*$`)

// Latest asks Packagist for the newest stable release of each package, a few
// at a time. Packagist lists versions newest first, so the first release
// (not a branch, alpha, beta or release candidate) wins. A package it does
// not have (a private repository's) is in neither list.
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
			var doc struct {
				Packages map[string][]struct {
					Version string `json:"version"`
				} `json:"packages"`
			}
			err := FetchJSON(PackagistMetadata+"/"+name+".json", &doc)
			version := ""
			for _, v := range doc.Packages[name] {
				if release.MatchString(v.Version) {
					version = strings.TrimPrefix(v.Version, "v")
					break
				}
			}
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil && strings.Contains(err.Error(), "404"):
			case err != nil:
				res.Unreachable[name] = err.Error()
			case version != "":
				res.Latest[name] = version
			}
		}(name)
	}
	wg.Wait()
	return res, nil
}
