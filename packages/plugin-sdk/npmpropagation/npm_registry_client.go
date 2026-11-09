package npmpropagation

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// NPM asks the registry through the npm CLI, then fetches tarballs over HTTP.
type NPM struct {
	// Dir is where npm runs; Args are added to every call (--registry <url>).
	Dir  string
	Args []string
	// Output runs a command and returns its stdout.
	Output func(dir string, argv []string) (string, error)
	// Head is the HTTP client for tarballs; a 15 s client when nil.
	Head *http.Client
}

// Lookup reads version and dist.tarball with `npm view --prefer-online`, so
// the cache of the machine cannot answer for the registry.
func (n NPM) Lookup(name, version string) (bool, string) {
	argv := append([]string{"npm", "view", name + "@" + version, "version", "dist.tarball", "--json", "--prefer-online"}, n.Args...)
	out, err := n.Output(n.Dir, argv)
	if err != nil || strings.TrimSpace(out) == "" {
		return false, ""
	}
	var got struct {
		Version string `json:"version"`
		Tarball string `json:"dist.tarball"`
	}
	if json.Unmarshal([]byte(out), &got) != nil || got.Version != version {
		return false, ""
	}
	return true, got.Tarball
}

// Fetchable sends a HEAD request and accepts a 2xx answer.
func (n NPM) Fetchable(url string) bool {
	c := n.Head
	if c == nil {
		c = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := c.Head(url)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// RealClock is the wall clock.
type RealClock struct{}

// Now returns the current time.
func (RealClock) Now() time.Time { return time.Now() }

// Sleep pauses for d.
func (RealClock) Sleep(d time.Duration) { time.Sleep(d) }
