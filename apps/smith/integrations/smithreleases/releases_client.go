package smithreleases

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// PublicKey is the minisign key smith's releases are signed with (key ID
// 6B2CE84DD9F901); its private half is the repository secret SMITH_SIGNING_KEY.
const PublicKey = "RWQB+dlN6CxrAMNDX47n/BQK3mOTgrJ4yVIasbrbP6xRRZTQeB5yncav"

// githubDownloads is where the release files are; githubReleases lists them.
const (
	githubDownloads = "https://github.com/monosmith-app/monosmith/releases/download"
	githubReleases  = "https://api.github.com/repos/monosmith-app/monosmith/releases?per_page=100"
)

// Base is the address a release's files are under: SMITH_DOWNLOAD_URL (a
// mirror of the release archives), else GitHub.
//
// Returns: the base, without a trailing slash.
func Base() string {
	if mirror := strings.TrimRight(os.Getenv("SMITH_DOWNLOAD_URL"), "/"); mirror != "" {
		return mirror
	}
	return githubDownloads
}

// get fetches a URL and returns its body, or an error that names the URL.
func get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s answered %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// File downloads one file of a release: <Base>/smith@<version>/<name>.
//
// Parameters: version is an exact version; name is the file's name.
// Returns: the file's bytes, or an error when it cannot be downloaded.
func File(ctx context.Context, version, name string) ([]byte, error) {
	return get(ctx, Base()+"/smith%40"+version+"/"+name, nil)
}

// Newest is the newest smith on a channel. A mirror serves latest.txt (and
// next.txt); GitHub lists the smith@ releases, newest first: "latest" is the
// newest that is not a prerelease, "next" the newest of all. GITHUB_TOKEN or
// GH_TOKEN, when set, raises GitHub's rate limit.
//
// Parameters: channel is "latest" or "next".
// Returns: the version, or an error when no release answers.
func Newest(ctx context.Context, channel string) (string, error) {
	if os.Getenv("SMITH_DOWNLOAD_URL") != "" {
		name := "latest.txt"
		if channel == "next" {
			name = "next.txt"
		}
		body, err := get(ctx, Base()+"/"+name, nil)
		return strings.TrimSpace(string(body)), err
	}
	headers := map[string]string{"Accept": "application/vnd.github+json"}
	for _, env := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if token := os.Getenv(env); token != "" {
			headers["Authorization"] = "Bearer " + token
			break
		}
	}
	body, err := get(ctx, githubReleases, headers)
	if err != nil {
		return "", err
	}
	var releases []struct {
		Tag        string `json:"tag_name"`
		Prerelease bool   `json:"prerelease"`
		Draft      bool   `json:"draft"`
	}
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", err
	}
	for _, r := range releases {
		if v, ok := strings.CutPrefix(r.Tag, "smith@"); ok && !r.Draft && (channel == "next" || !r.Prerelease) {
			return v, nil
		}
	}
	return "", fmt.Errorf("GitHub lists no smith release on the %s channel", channel)
}
