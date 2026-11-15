package pypiwheels

import (
	"errors"
	"path/filepath"
	"sort"
)

// Result says what happened to each wheel.
type Result struct {
	Published, Existing []string
	// Skipped are the archives no wheel platform matches.
	Skipped []string
}

// upload sends one wheel; tests replace it.
var upload = func(w Wheel) error {
	u, err := UploaderFromEnvironment()
	if err != nil {
		return err
	}
	return u.Upload(w)
}

// Publish builds the release's wheels and uploads them. A wheel the index
// already has is left alone, so a re-run completes a release. A dry run only
// names the wheels.
func Publish(root string, r Release, dryRun bool) (Result, error) {
	var res Result
	if r.Name == "" {
		return res, nil
	}
	sort.Slice(r.Archives, func(i, j int) bool { return r.Archives[i].Path < r.Archives[j].Path })
	if dryRun {
		version, err := Version(r.Version)
		if err != nil {
			return res, err
		}
		for _, a := range r.Archives {
			if tags, ok := Tags(a.OS, a.Arch); ok {
				res.Published = append(res.Published, FileName(r.Name, version, tags))
			} else {
				res.Skipped = append(res.Skipped, a.OS+"/"+a.Arch)
			}
		}
		return res, nil
	}
	wheels, skipped, err := Build(root, r)
	res.Skipped = skipped
	if err != nil {
		return res, err
	}
	if len(wheels) == 0 {
		return res, errors.New("no archive for a platform PyPI has wheels for")
	}
	for _, w := range wheels {
		switch err := upload(w); {
		case errors.Is(err, ErrExists):
			res.Existing = append(res.Existing, filepath.Base(w.File))
		case err != nil:
			return res, err
		default:
			res.Published = append(res.Published, filepath.Base(w.File))
		}
	}
	return res, nil
}
