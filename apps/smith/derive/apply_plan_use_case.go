package derive

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/derivefile"
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Decision is what a person chose for files that need one, by workspace path.
type Decision struct {
	// Replace lets smith overwrite a drifted or diverged file with the
	// renderer's content. The caller has shown the diff.
	Replace []string
	// Adopt takes the hand-edited content of a keys-mode file as the new
	// baseline, when the file still has every derived key.
	Adopt []string
	// Reclaim gives an ejected, drifted or diverged file back to smith: it is
	// rewritten with the derived content and the record says owner "smith"
	// again. The caller has shown the diff.
	Reclaim []string
	// Eject gives the file to the person: the record says owner "you" and smith
	// stops touching it.
	Eject []string
}

// Outcome lists what Apply did, by path.
type Outcome struct {
	Written  []string
	Recorded []string
	Replaced []string
	Adopted  []string
	Ejected  []string
	// Reclaimed are the files given back to smith.
	Reclaimed []string
	// Held are the items Apply left alone because they need a decision.
	Held []Item
}

// Apply writes the files the plan says smith may write, updates the record in
// memory and returns what happened. A drifted, diverged or broken file stays
// as it is unless the decision names it. The caller saves the record. It
// returns an error when a decision names a file smith does not derive, adopt
// does not fit the file, or a write fails.
func Apply(root string, plan Plan, rec *derivelock.Record, d Decision) (Outcome, error) {
	var out Outcome
	for _, p := range slices.Concat(d.Replace, d.Adopt, d.Eject, d.Reclaim) {
		if len(itemsAt(plan, p)) == 0 && len(rec.EntriesFor(p)) == 0 {
			return out, problem.New("smith does not derive "+p, "smith derive --check lists the files it derives")
		}
	}
	changes, err := PlanReclaim(root, plan, d.Reclaim)
	if err != nil {
		return out, err
	}
	for _, item := range plan.Items {
		p := item.File.Path
		switch {
		case slices.Contains(d.Eject, p):
			rec.Set(entryOf(item, derivelock.OwnerYou))
			out.Ejected = appendOnce(out.Ejected, p)
		case slices.Contains(d.Reclaim, p):
			if c := changes[p]; c.Changed() && !reclaimed(out, p) {
				if err := write(root, Item{File: item.File, Merged: c.After}); err != nil {
					return out, err
				}
			}
			rec.Set(entryOf(item, derivelock.OwnerSmith))
			out.Reclaimed = appendOnce(out.Reclaimed, p)
		case item.Status == StatusCurrent || item.Status == StatusEjected:
		case item.Status == StatusUnrecorded:
			rec.Set(entryOf(item, derivelock.OwnerSmith))
			out.Recorded = append(out.Recorded, p)
		case item.Status == StatusStale || item.Status == StatusMissing:
			if err := write(root, item); err != nil {
				return out, err
			}
			rec.Set(entryOf(item, derivelock.OwnerSmith))
			out.Written = append(out.Written, p)
		case slices.Contains(d.Adopt, p):
			if err := adopt(root, item, rec); err != nil {
				return out, err
			}
			out.Adopted = append(out.Adopted, p)
		case slices.Contains(d.Replace, p) && item.Merged != nil && (item.Status == StatusDrift || item.Status == StatusDiverged):
			if err := write(root, item); err != nil {
				return out, err
			}
			rec.Set(entryOf(item, derivelock.OwnerSmith))
			out.Replaced = append(out.Replaced, p)
		default:
			out.Held = append(out.Held, item)
		}
	}
	for _, p := range d.Eject {
		if len(itemsAt(plan, p)) == 0 {
			for _, e := range rec.EntriesFor(p) {
				e.Owner = derivelock.OwnerYou
			}
			out.Ejected = appendOnce(out.Ejected, p)
		}
	}
	return out, nil
}

func itemsAt(plan Plan, path string) []Item {
	var out []Item
	for _, i := range plan.Items {
		if i.File.Path == path {
			out = append(out, i)
		}
	}
	return out
}

func appendOnce(list []string, s string) []string {
	if slices.Contains(list, s) {
		return list
	}
	return append(list, s)
}

// entryOf is the record entry for an item once smith wrote it (or an owner
// took it over): the hashes are those of the content the renderer wanted.
func entryOf(item Item, owner string) derivelock.Entry {
	h := derivelock.Hash(item.Wanted)
	e := derivelock.Entry{Path: item.File.Path, Owner: owner, Renderer: item.File.Renderer, Version: item.File.Version, Mode: item.File.Mode, Hash: h, Wanted: h}
	if owner == derivelock.OwnerYou && item.Present != nil {
		e.Hash = derivelock.Hash(item.Present)
	}
	return e
}

func write(root string, item Item) error {
	path := filepath.Join(root, filepath.FromSlash(item.File.Path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, item.Merged, 0o644)
}

// adopt records the hand-edited content of a drifted keys-mode file as the
// baseline, against what the renderer wants now.
func adopt(root string, item Item, rec *derivelock.Record) error {
	if item.Status != StatusDrift {
		return problem.New(item.File.Path+" has no hand edit to adopt ("+string(item.Status)+")", "use --replace to let smith write it, or --eject to own it")
	}
	if item.File.Mode != protocol.DeriveKeys {
		return problem.New(item.File.Path+" is a "+item.File.Mode+" file: adopt applies to files that smith owns by keys", "use --replace to let smith overwrite it (see the diff first), or --eject to own it")
	}
	disk, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(item.File.Path)))
	if err != nil {
		return err
	}
	all, err := derivefile.HasAllKeys(item.File, disk)
	if err != nil {
		return err
	}
	if !all {
		return problem.New(fmt.Sprintf("%s lost a derived key, so its edit cannot be adopted", item.File.Path), "use --replace to restore the key, or --eject to own the file")
	}
	e := entryOf(item, derivelock.OwnerSmith)
	e.Hash = derivelock.Hash(item.Present)
	rec.Set(e)
	return nil
}
