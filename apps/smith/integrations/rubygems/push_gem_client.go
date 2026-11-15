package rubygems

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultHost is rubygems.org.
const DefaultHost = "https://rubygems.org"

// ErrExists is returned for a gem version the host already has.
var ErrExists = errors.New("the host already has this gem version")

// Pusher sends gems to a gem host.
type Pusher struct {
	Host, Key string
	Client    *http.Client
}

// PusherFromEnvironment reads the host and key the way `gem push` does:
// RUBYGEMS_HOST (default rubygems.org) and GEM_HOST_API_KEY.
func PusherFromEnvironment() (Pusher, error) {
	p := Pusher{Host: os.Getenv("RUBYGEMS_HOST"), Key: os.Getenv("GEM_HOST_API_KEY"), Client: &http.Client{Timeout: 10 * time.Minute}}
	if p.Host == "" {
		p.Host = DefaultHost
	}
	if p.Key == "" {
		return p, errors.New("set GEM_HOST_API_KEY to push gems")
	}
	return p, nil
}

// Push sends one gem. A version the host already has is ErrExists, so a
// re-run completes a release.
func (p Pusher) Push(g Gem) error {
	content, err := os.ReadFile(g.File)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, strings.TrimSuffix(p.Host, "/")+"/api/v1/gems", bytes.NewReader(content))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", p.Key)
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	reply, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode/100 == 2 {
		return nil
	}
	// rubygems.org answers a repush with 409 "Repushing of gem versions is
	// not allowed."
	if resp.StatusCode == http.StatusConflict || strings.Contains(string(reply), "Repushing") {
		return ErrExists
	}
	return fmt.Errorf("pushing %s: %s: %s", filepath.Base(g.File), resp.Status, strings.TrimSpace(string(reply)))
}
