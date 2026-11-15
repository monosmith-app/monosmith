package rubygems

import (
	"errors"
	"path/filepath"
	"sort"
)

// Result says what happened to each gem.
type Result struct {
	Published, Existing []string
	// Skipped are the archives no gem platform matches.
	Skipped []string
}

// push sends one gem; tests replace it.
var push = func(g Gem) error {
	p, err := PusherFromEnvironment()
	if err != nil {
		return err
	}
	return p.Push(g)
}

// Publish builds the release's platform gems and pushes them. A version the
// host already has is left alone, so a re-run completes a release. A dry run
// only names the gems.
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
			names, ok := Platforms(a.OS, a.Arch)
			if !ok {
				res.Skipped = append(res.Skipped, a.OS+"/"+a.Arch)
			}
			for _, p := range names {
				res.Published = append(res.Published, FileName(r.Name, version, p))
			}
		}
		return res, nil
	}
	gems, skipped, err := Build(root, r)
	res.Skipped = skipped
	if err != nil {
		return res, err
	}
	if len(gems) == 0 {
		return res, errors.New("no archive for a platform RubyGems has gems for")
	}
	for _, g := range gems {
		switch err := push(g); {
		case errors.Is(err, ErrExists):
			res.Existing = append(res.Existing, filepath.Base(g.File))
		case err != nil:
			return res, err
		default:
			res.Published = append(res.Published, filepath.Base(g.File))
		}
	}
	return res, nil
}
