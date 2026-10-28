package updatenotice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RegistryURL answers with the newest published smith.
const RegistryURL = "https://registry.npmjs.org/monosmith/latest"

// ReleasesURL is where the release notes are.
const ReleasesURL = "https://github.com/monosmith-app/monosmith/releases"

// FetchLatest asks the registry for the newest version, giving up after timeout.
func FetchLatest(ctx context.Context, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, RegistryURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry answered %s", resp.Status)
	}
	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.Version, nil
}
