package nugettool

import (
	"errors"
	"path/filepath"
	"sort"
)

// Result says what happened to each package.
type Result struct {
	Published, Existing []string
	// Skipped are the archives no runtime identifier matches.
	Skipped []string
}

// push sends one package; tests replace it.
var push = func(pkg Package) error {
	p, err := PusherFromEnvironment()
	if err != nil {
		return err
	}
	return p.Push(pkg)
}

// Publish builds the release's packages and pushes them, each platform's
// before the tool package that points at them. A version the feed already
// has is left alone, so a re-run completes a release. A dry run only names
// the packages.
func Publish(root string, r Release, dryRun bool) (Result, error) {
	var res Result
	if r.ID == "" {
		return res, nil
	}
	if dryRun {
		archives := append([]Archive(nil), r.Archives...)
		sort.Slice(archives, func(i, j int) bool { return archives[i].Path < archives[j].Path })
		for _, a := range archives {
			if rid, ok := RID(a.OS, a.Arch); ok {
				res.Published = append(res.Published, FileName(r.ID+"."+rid, r.Version))
			} else {
				res.Skipped = append(res.Skipped, a.OS+"/"+a.Arch)
			}
		}
		if len(res.Published) > 0 {
			res.Published = append(res.Published, FileName(r.ID, r.Version))
		}
		return res, nil
	}
	packages, skipped, err := Build(root, r)
	res.Skipped = skipped
	if err != nil {
		return res, err
	}
	if len(packages) == 0 {
		return res, errors.New("no archive for a platform .NET has a runtime identifier for")
	}
	for _, pkg := range packages {
		switch err := push(pkg); {
		case errors.Is(err, ErrExists):
			res.Existing = append(res.Existing, filepath.Base(pkg.File))
		case err != nil:
			return res, err
		default:
			res.Published = append(res.Published, filepath.Base(pkg.File))
		}
	}
	return res, nil
}
