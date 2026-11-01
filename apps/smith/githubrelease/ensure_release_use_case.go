package githubrelease

import (
	"fmt"
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
}

// Outcome reports what Ensure did.
type Outcome struct {
	URL      string
	Created  bool
	Uploaded []string
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
	have := map[string]bool{}
	for _, a := range r.Assets {
		have[a.Name] = true
	}
	for _, file := range req.Assets {
		name := filepath.Base(file)
		if have[name] {
			continue
		}
		if err := c.UploadAsset(r, file); err != nil {
			return out, fmt.Errorf("uploading %s: %w", name, err)
		}
		out.Uploaded = append(out.Uploaded, name)
	}
	return out, nil
}
