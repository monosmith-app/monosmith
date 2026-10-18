package cipipeline

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// verdictClock is the day a verdict is made on; tests replace it. The day
// stands for the advisory database: a new advisory shows up within a day, so
// a verdict is never replayed on the next one.
var verdictClock = time.Now

// verdicts remembers which audit steps passed, in the task cache (so the
// remote cache shares them), keyed on the content of the files the step
// declares (protocol.Step.CacheFiles) and on the day.
type verdicts struct {
	store   taskcache.Storage
	root    string
	version string
	files   *taskcache.FileDigests
}

// openVerdicts returns the verdict store of a session, or nil when the cache
// cannot be opened: the audit then runs every step, as it always did.
func openVerdicts(s *session.Session) *verdicts {
	store, err := s.Storage()
	if err != nil {
		return nil
	}
	return &verdicts{store: store, root: s.Root, version: s.Version, files: taskcache.NewFileDigests(s.Root)}
}

// keyOf is the key of a step's verdict and what it is made of; ok is false
// when the step declares no file that exists, so nothing says its verdict
// still holds.
func (v *verdicts) keyOf(step protocol.Step) (key string, inputs taskcache.KeyInputs, ok bool) {
	var digests []taskcache.FileDigest
	for _, f := range step.CacheFiles {
		if _, err := os.Stat(filepath.Join(v.root, filepath.FromSlash(f))); err != nil {
			continue
		}
		digests = append(digests, taskcache.FileDigest{Path: f, Digest: v.files.Digest(f)})
	}
	if len(digests) == 0 {
		return "", inputs, false
	}
	inputs = taskcache.KeyInputs{
		Task: "audit:" + step.Name, Command: step.Command, Files: digests, Platform: taskrun.Platform, Tool: v.version,
		Env: map[string]string{"advisories": verdictClock().UTC().Format("2006-01-02"), "dir": path.Clean("/" + step.Dir)},
	}
	return taskcache.Key(inputs), inputs, true
}

// passed reports whether the step already passed for these files on this day.
func (v *verdicts) passed(key string) (time.Time, bool) {
	entry, err := taskcache.Lookup(v.store, key)
	if err != nil {
		return time.Time{}, false
	}
	return entry.Created, true
}

// remember records that the step passed.
func (v *verdicts) remember(key string, step protocol.Step, inputs taskcache.KeyInputs, took time.Duration) error {
	return taskcache.Save(v.store, v.root, key, inputs.Task, "", nil, nil, took, &inputs)
}

// replay says a step was skipped because its verdict holds.
func replay(out io.Writer, step protocol.Step, since time.Time) {
	fmt.Fprintf(out, "✓ %s: no change in %s since it passed on %s\n", step.Name, listFiles(step.CacheFiles), since.UTC().Format("2006-01-02"))
}

func listFiles(files []string) string {
	var b bytes.Buffer
	for i, f := range files {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(f)
	}
	return b.String()
}
