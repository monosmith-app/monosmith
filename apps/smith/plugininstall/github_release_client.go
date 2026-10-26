package plugininstall

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginarchive"
)

// releaseAsset is the part of a release asset the installer uses.
type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// resolvedAsset is the archive a release offers, and the digest it publishes.
type resolvedAsset struct {
	Name string
	URL  string
	// SHA256 is the digest in the release's <asset>.sha256 asset, empty when it has none.
	SHA256 string
}

func githubAPI() string {
	if api := strings.TrimSuffix(os.Getenv("GITHUB_API_URL"), "/"); api != "" {
		return api
	}
	return "https://api.github.com"
}

// resolveRelease finds the plugin archive of a GitHub release by its tag, and
// the digest the release publishes for it (an asset named like the archive
// plus ".sha256", which `smith plugin pack` writes). GITHUB_TOKEN, when set,
// is sent to the API only, never to the asset downloads.
func resolveRelease(ctx context.Context, client *http.Client, src Source) (resolvedAsset, error) {
	api := githubAPI()
	if err := CheckURL(api); err != nil {
		return resolvedAsset{}, err
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", api, src.Owner, src.Repo, url.PathEscape(src.Tag))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return resolvedAsset{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "smith")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return resolvedAsset{}, fmt.Errorf("looking up %s: %w", src, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return resolvedAsset{}, fmt.Errorf("no release %s in %s/%s (or the repository is private: set GITHUB_TOKEN)", src.Tag, src.Owner, src.Repo)
	}
	if resp.StatusCode != http.StatusOK {
		return resolvedAsset{}, fmt.Errorf("looking up %s: %s", src, resp.Status)
	}
	var release struct {
		Assets []releaseAsset `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&release); err != nil {
		return resolvedAsset{}, fmt.Errorf("reading the release %s: %w", src, err)
	}
	asset, err := pickAsset(release.Assets, src)
	if err != nil {
		return resolvedAsset{}, err
	}
	out := resolvedAsset{Name: asset.Name, URL: asset.URL}
	for _, a := range release.Assets {
		if a.Name == asset.Name+".sha256" {
			out.SHA256, err = fetchDigest(ctx, client, a.URL, asset.Name)
			if err != nil {
				return resolvedAsset{}, err
			}
		}
	}
	return out, nil
}

// pickAsset chooses the archive: the one named in the source, or the only
// asset that ends in the plugin archive suffix.
func pickAsset(assets []releaseAsset, src Source) (releaseAsset, error) {
	var found []releaseAsset
	for _, a := range assets {
		if src.Asset != "" && a.Name == src.Asset || src.Asset == "" && strings.HasSuffix(a.Name, pluginarchive.Suffix) {
			found = append(found, a)
		}
	}
	switch len(found) {
	case 1:
		return found[0], nil
	case 0:
		if src.Asset != "" {
			return releaseAsset{}, fmt.Errorf("release %s has no asset %s", src.Tag, src.Asset)
		}
		return releaseAsset{}, fmt.Errorf("release %s has no asset ending in %s; run `smith plugin pack` and attach the archive", src.Tag, pluginarchive.Suffix)
	}
	names := make([]string, len(found))
	for i, a := range found {
		names[i] = a.Name
	}
	sort.Strings(names)
	return releaseAsset{}, fmt.Errorf("release %s has several plugin archives (%s); pick one: %s#<asset>", src.Tag, strings.Join(names, ", "), src)
}

// fetchDigest reads a "<sha256>  <name>" (or bare digest) asset.
func fetchDigest(ctx context.Context, client *http.Client, rawURL, archive string) (string, error) {
	var buf strings.Builder
	if _, _, err := download(ctx, client, rawURL, "", 4096, &buf); err != nil {
		return "", err
	}
	fields := strings.Fields(buf.String())
	if len(fields) == 0 || len(fields[0]) != 64 {
		return "", fmt.Errorf("the digest published for %s is not a sha256", archive)
	}
	return strings.ToLower(fields[0]), nil
}
