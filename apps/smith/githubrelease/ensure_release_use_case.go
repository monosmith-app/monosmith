package githubrelease

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// Request describes the release of one tag.
type Request struct {
	Tag   string
	Title string
	Notes string
	// Prerelease marks the release as a prerelease, which GitHub never offers
	// as the latest release.
	Prerelease bool
	// Assets are absolute paths of files to attach.
	Assets []string
	// Verify compares an asset that is attached already, by content, with the
	// local file of the same name, and replaces it when they differ. A package
	// file packed again by a re-run may differ from the one an earlier run
	// attached; its checksum file must describe what is attached. Without it an
	// attached asset is kept as it is.
	Verify bool
}

// Outcome reports what Ensure did.
type Outcome struct {
	URL      string
	Created  bool
	Uploaded []string
	// Replaced lists the attached assets that differed from the local file
	// and were uploaded again (Request.Verify).
	Replaced []string
}

// Ensure makes sure the tag has a release with every asset attached. It is
// safe to repeat: an existing release keeps its notes and receives only the
// assets it lacks.
func Ensure(c *Client, repo Repository, req Request) (Outcome, error) {
	r, found, err := c.ReleaseByTag(repo, req.Tag)
	if err != nil {
		return Outcome{}, err
	}
	out := Outcome{}
	if !found {
		if r, err = c.CreateRelease(repo, req.Tag, req.Title, req.Notes, req.Prerelease); err != nil {
			return out, err
		}
		out.Created = true
	}
	out.URL = r.HTMLURL
	have := map[string]Asset{}
	for _, a := range r.Assets {
		have[a.Name] = a
	}
	for _, file := range req.Assets {
		name := filepath.Base(file)
		if attached, ok := have[name]; ok {
			if !req.Verify {
				continue
			}
			same, err := sameContent(c, repo, attached, file)
			if err != nil {
				return out, fmt.Errorf("checking %s: %w", name, err)
			}
			if same {
				continue
			}
			if err := c.DeleteAsset(repo, attached); err != nil {
				return out, fmt.Errorf("replacing %s: %w", name, err)
			}
			out.Replaced = append(out.Replaced, name)
		}
		if err := c.UploadAsset(r, file); err != nil {
			return out, fmt.Errorf("uploading %s: %w", name, err)
		}
		out.Uploaded = append(out.Uploaded, name)
	}
	return out, nil
}

// sameContent reports whether an attached asset holds the bytes of a local
// file. The sizes are compared first: only equal sizes are downloaded.
func sameContent(c *Client, repo Repository, a Asset, file string) (bool, error) {
	local, err := os.ReadFile(file)
	if err != nil {
		return false, err
	}
	if a.Size != 0 && a.Size != int64(len(local)) {
		return false, nil
	}
	tmp, err := os.MkdirTemp("", "smith-asset-")
	if err != nil {
		return false, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	got := filepath.Join(tmp, "asset")
	if err := c.DownloadAsset(repo, a, got); err != nil {
		return false, err
	}
	remote, err := os.ReadFile(got)
	if err != nil {
		return false, err
	}
	return sha256.Sum256(local) == sha256.Sum256(remote), nil
}
