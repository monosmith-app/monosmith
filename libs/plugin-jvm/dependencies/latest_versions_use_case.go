package dependencies

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// MavenCentral is where newest releases are asked: the repository
// dependencies resolve from unless a build says otherwise.
const MavenCentral = "https://repo1.maven.org/maven2"

// FetchMetadata reads a repository's maven-metadata.xml; tests replace it.
// It returns nil, nil for a 404: the repository does not have the artifact.
var FetchMetadata = func(url string) ([]byte, error) {
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(resp.Body)
	case http.StatusNotFound:
		return nil, nil
	}
	return nil, fmt.Errorf("%s: %s", url, resp.Status)
}

// stable is a release version, as opposed to a milestone, release candidate
// or snapshot: dotted numbers, optionally .RELEASE, .Final or .GA.
var stable = regexp.MustCompile(`^\d+(\.\d+)*([-.](RELEASE|Final|GA))?$`)

// newestStable is the highest stable version in a maven-metadata.xml. The
// order of its <versions> is publishing order, so a backport shipped after a
// newer line is last there; the highest number wins instead.
func newestStable(metadata []byte) (string, error) {
	var doc struct {
		Versions []string `xml:"versioning>versions>version"`
	}
	if err := xml.Unmarshal(metadata, &doc); err != nil {
		return "", err
	}
	best := ""
	for _, v := range doc.Versions {
		v = strings.TrimSpace(v)
		if stable.MatchString(v) && (best == "" || higher(v, best)) {
			best = v
		}
	}
	return best, nil
}

// higher compares dotted version numbers part by part.
func higher(a, b string) bool {
	number := func(v string) []int {
		var parts []int
		for _, p := range strings.Split(strings.SplitN(v, "-", 2)[0], ".") {
			n, err := strconv.Atoi(p)
			if err != nil {
				break
			}
			parts = append(parts, n)
		}
		return parts
	}
	x, y := number(a), number(b)
	for i := 0; i < len(x) || i < len(y); i++ {
		var xi, yi int
		if i < len(x) {
			xi = x[i]
		}
		if i < len(y) {
			yi = y[i]
		}
		if xi != yi {
			return xi > yi
		}
	}
	return false
}

// Latest asks Maven Central for the newest stable release of each
// groupId:artifactId, a few at a time. An artifact it does not have (an
// internal repository's) is in neither list.
func Latest(_ string, params protocol.LatestParams) (protocol.LatestResult, error) {
	res := protocol.LatestResult{Latest: map[string]string{}, Unreachable: map[string]string{}}
	var mu sync.Mutex
	var wg sync.WaitGroup
	slots := make(chan struct{}, 8)
	for _, name := range params.Names {
		group, artifact, ok := strings.Cut(name, ":")
		if !ok {
			continue
		}
		wg.Add(1)
		slots <- struct{}{}
		go func(name, group, artifact string) {
			defer func() { <-slots; wg.Done() }()
			metadata, err := FetchMetadata(MavenCentral + "/" + strings.ReplaceAll(group, ".", "/") + "/" + artifact + "/maven-metadata.xml")
			var version string
			if err == nil && metadata != nil {
				version, err = newestStable(metadata)
			}
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil:
				res.Unreachable[name] = err.Error()
			case version != "":
				res.Latest[name] = version
			}
		}(name, group, artifact)
	}
	wg.Wait()
	return res, nil
}
