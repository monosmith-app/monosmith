package releasepack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
)

// AttachBundle attaches the bundle (release-bundle.json and INSTALL.md) to the
// GitHub Release of every package in it, so a consumer who opens any one of
// them finds the whole set. A release that holds a bundle of an earlier,
// partial run keeps the packages the new bundle does not list; one whose
// bundle differs is replaced. The files are also kept under
// dist/release/bundle/<tag>/ in the workspace.
//
// Parameters: c and repo reach GitHub; root is the workspace; b is the bundle
// of this run.
// Returns: the release tags that received a bundle.
// Errors: a release cannot be found or read, or a file cannot be written or
// uploaded. A release the run did not create is an error: the bundle goes on
// releases that exist.
func AttachBundle(c *githubrelease.Client, repo githubrelease.Repository, root string, b Bundle) ([]string, error) {
	var attached []string
	done := map[string]bool{}
	for _, p := range b.Packages {
		if done[p.Tag] {
			continue
		}
		done[p.Tag] = true
		release, found, err := c.ReleaseByTag(repo, p.Tag)
		if err != nil {
			return attached, err
		}
		if !found {
			return attached, fmt.Errorf("the GitHub Release %s does not exist", p.Tag)
		}
		merged, err := mergeWithAttached(c, repo, release, b)
		if err != nil {
			return attached, err
		}
		dir := filepath.Join(root, "dist", "release", "bundle", safeName(p.Tag))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return attached, err
		}
		raw, err := BundleJSON(merged)
		if err != nil {
			return attached, err
		}
		files := []string{filepath.Join(dir, BundleFile), filepath.Join(dir, BundleMarkdown)}
		if err := os.WriteFile(files[0], raw, 0o644); err != nil {
			return attached, err
		}
		if err := os.WriteFile(files[1], []byte(BundleInstructions(merged)), 0o644); err != nil {
			return attached, err
		}
		if _, err := githubrelease.Ensure(c, repo, githubrelease.Request{Tag: p.Tag, Assets: files, Verify: true}); err != nil {
			return attached, fmt.Errorf("bundle for %s: %w", p.Tag, err)
		}
		attached = append(attached, p.Tag)
	}
	return attached, nil
}

// mergeWithAttached reads the bundle a release holds already, if any, and
// merges the packages it lists that the new bundle does not.
func mergeWithAttached(c *githubrelease.Client, repo githubrelease.Repository, r githubrelease.Release, b Bundle) (Bundle, error) {
	for _, a := range r.Assets {
		if a.Name != BundleFile {
			continue
		}
		dir, err := os.MkdirTemp("", "smith-bundle-")
		if err != nil {
			return b, err
		}
		defer func() { _ = os.RemoveAll(dir) }()
		to := filepath.Join(dir, BundleFile)
		if err := c.DownloadAsset(repo, a, to); err != nil {
			return b, err
		}
		raw, err := os.ReadFile(to)
		if err != nil {
			return b, err
		}
		var older Bundle
		if err := json.Unmarshal(raw, &older); err != nil {
			// A bundle that cannot be read is replaced, not trusted.
			return b, nil
		}
		return MergeBundles(older, b), nil
	}
	return b, nil
}

// safeName makes a tag usable as a folder name: core@0.5.2 and a/b@1 both stay readable.
func safeName(tag string) string {
	return strings.NewReplacer("/", "-", "\\", "-", ":", "-").Replace(tag)
}
