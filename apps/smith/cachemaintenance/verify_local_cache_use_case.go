package cachemaintenance

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// Verify checks the local store in dir: every blob hashes to its name, every
// entry is readable and names only blobs that exist, no blob lacks an entry,
// and tmp/ holds nothing left behind. With fix it removes what is broken: the
// corrupt blobs and the entries that name them, entries that cannot be read or
// name a missing blob, and the blobs and tmp files nothing uses; a problem a
// removal fixed is marked Fixed. A blob written, or a file in tmp/, within the
// last minutes may be a write in progress and is not reported.
func Verify(dir string, fix bool, now time.Time) (VerifyReport, error) {
	report := VerifyReport{Dir: dir, Problems: []Problem{}}
	store, ok, err := openExisting(dir)
	if err != nil || !ok {
		return report, err
	}
	type blob struct {
		last    time.Time
		corrupt bool
	}
	blobs := map[string]*blob{}
	if err := store.Walk(taskcache.CAS, func(key string, _ int64, last time.Time) error {
		blobs[key] = &blob{last: last}
		return nil
	}); err != nil {
		return report, err
	}
	report.Blobs = len(blobs)
	for key, b := range blobs {
		sum, err := hashBlob(store, key)
		if err != nil {
			delete(blobs, key) // removed since the walk
			report.Blobs--
			continue
		}
		if sum != key {
			b.corrupt = true
			report.add(Problem{Kind: CorruptBlob, Key: key, Detail: "its content hashes to " + short(sum) + ", not to its name"})
		}
	}

	type entry struct {
		key   string
		blobs []string
	}
	var good []entry
	var broken []string
	if err := store.Walk(taskcache.AC, func(key string, _ int64, _ time.Time) error {
		e, state := readEntry(store, key)
		switch state {
		case entryGone:
			return nil
		case entryBad:
			report.Entries++
			report.add(Problem{Kind: UnreadableEntry, Key: key, Detail: "it is not a cache entry"})
			broken = append(broken, key)
			return nil
		}
		report.Entries++
		var bad []string
		for _, digest := range entryBlobs(e) {
			b, present := blobs[digest]
			switch {
			case !present:
				report.add(Problem{Kind: MissingBlob, Key: key, Detail: "the entry for " + e.Task + " names blob " + short(digest) + ", which is gone"})
				bad = append(bad, digest)
			case b.corrupt:
				report.add(Problem{Kind: DamagedEntry, Key: key, Detail: "the entry for " + e.Task + " names blob " + short(digest) + ", which is corrupt"})
				bad = append(bad, digest)
			}
		}
		if len(bad) > 0 {
			broken = append(broken, key)
			return nil
		}
		good = append(good, entry{key: key, blobs: entryBlobs(e)})
		return nil
	}); err != nil {
		return report, err
	}
	used := map[string]bool{}
	for _, e := range good {
		for _, d := range e.blobs {
			used[d] = true
		}
	}
	var orphans []string
	for key, b := range blobs {
		if used[key] || b.corrupt || now.Sub(b.last) < OrphanGrace {
			continue
		}
		orphans = append(orphans, key)
		report.add(Problem{Kind: OrphanBlob, Key: key, Detail: "no entry uses it"})
	}
	var tmps []string
	if files, err := os.ReadDir(filepath.Join(store.Root(), "tmp")); err == nil {
		for _, f := range files {
			if info, err := f.Info(); err == nil && now.Sub(info.ModTime()) >= OrphanGrace {
				tmps = append(tmps, f.Name())
				report.add(Problem{Kind: LeftoverTmp, Key: f.Name(), Detail: "a write that did not finish"})
			}
		}
	}

	if fix {
		removed := map[string]bool{}
		for _, key := range broken {
			if store.Delete(taskcache.AC, key) == nil {
				removed["entry:"+key] = true
			}
		}
		for key, b := range blobs {
			if b.corrupt && store.Delete(taskcache.CAS, key) == nil {
				removed["blob:"+key] = true
			}
		}
		for _, key := range orphans {
			if store.Delete(taskcache.CAS, key) == nil {
				removed["blob:"+key] = true
			}
		}
		for _, name := range tmps {
			if os.Remove(filepath.Join(store.Root(), "tmp", name)) == nil {
				removed["tmp:"+name] = true
			}
		}
		for i, p := range report.Problems {
			switch p.Kind {
			case CorruptBlob, OrphanBlob:
				report.Problems[i].Fixed = removed["blob:"+p.Key]
			case MissingBlob, UnreadableEntry, DamagedEntry:
				report.Problems[i].Fixed = removed["entry:"+p.Key]
			case LeftoverTmp:
				report.Problems[i].Fixed = removed["tmp:"+p.Key]
			}
		}
	}
	sort.SliceStable(report.Problems, func(i, j int) bool {
		a, b := report.Problems[i], report.Problems[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Key < b.Key
	})
	return report, nil
}

func (r *VerifyReport) add(p Problem) { r.Problems = append(r.Problems, p) }

// hashBlob is the SHA-256 of a stored blob.
func hashBlob(store taskcache.Storage, key string) (string, error) {
	r, err := store.Get(taskcache.CAS, key)
	if err != nil {
		return "", err
	}
	defer func() { _ = r.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// short is the start of a digest, enough to find the file by.
func short(digest string) string {
	if len(digest) > 12 {
		return digest[:12] + "…"
	}
	return digest
}
