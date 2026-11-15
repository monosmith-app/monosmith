package updatenotice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DistTagsURL is the one request smith makes for the update notice: the
// dist-tags of the monosmith package on the npm registry (latest and next).
// Nothing about the person or the workspace is sent.
const DistTagsURL = "https://registry.npmjs.org/-/package/monosmith/dist-tags"

// ReleasesURL is where the release notes are.
const ReleasesURL = "https://github.com/monosmith-app/monosmith/releases"

// FetchLatest asks the registry for the newest version on a channel, giving up
// after timeout.
//
// Parameters: channel is the dist-tag: "latest" or "next".
// Returns: the version the tag points to; an error when the registry does not
// answer in time, answers badly or has no such tag.
func FetchLatest(ctx context.Context, timeout time.Duration, channel string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, DistTagsURL, nil)
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
	var tags map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", err
	}
	if tags[channel] == "" {
		return "", fmt.Errorf("the registry has no %q tag", channel)
	}
	return tags[channel], nil
}
