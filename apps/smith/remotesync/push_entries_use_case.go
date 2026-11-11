package remotesync

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"sync"

	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// ErrReadOnly is returned by Push when this machine only reads the remote.
var ErrReadOnly = errors.New("the remote cache is read-only here")

// PushOptions change how Push works.
type PushOptions struct {
	// DryRun asks the remote what it lacks and writes nothing.
	DryRun bool
	// Parallel is how many entries are sent at once (0: 4).
	Parallel int
	// Progress, when set, is told about each entry as it is done, one call at a time.
	Progress func(PushStep)
}

// The outcomes of one entry.
const (
	// StepUploaded: the entry was sent.
	StepUploaded = "uploaded"
	// StepWouldUpload: a dry run found that the remote lacks the entry.
	StepWouldUpload = "would-upload"
	// StepPresent: the remote holds the entry already; nothing was sent.
	StepPresent = "present"
	// StepFailed: the entry could not be read or sent.
	StepFailed = "failed"
)

// PushStep is what happened to one local entry.
type PushStep struct {
	// Key is the entry's cache key; Task is the task that made it.
	Key  string `json:"key"`
	Task string `json:"task,omitempty"`
	// Outcome is StepUploaded, StepWouldUpload, StepPresent or StepFailed.
	Outcome string `json:"outcome"`
	// Blobs is how many of its blobs were (or would be) sent, Bytes their size
	// with the entry itself.
	Blobs int   `json:"blobs,omitempty"`
	Bytes int64 `json:"bytes,omitempty"`
	// Error says why a failed entry failed.
	Error string `json:"error,omitempty"`
}

// PushResult is the outcome of a push.
type PushResult struct {
	// Store names the remote (its URL, as configured).
	Store  string `json:"store"`
	DryRun bool   `json:"dryRun"`
	// Entries counts the local entries; the others split them by outcome.
	Entries  int `json:"entries"`
	Uploaded int `json:"uploaded"`
	Present  int `json:"present"`
	Failed   int `json:"failed"`
	// Blobs and Bytes total what was (or would be) sent.
	Blobs int        `json:"blobs"`
	Bytes int64      `json:"bytes"`
	Steps []PushStep `json:"steps"`
}

// Push sends the remote every local action-cache entry it lacks, with the
// blobs it names, signed as a run signs them (remotecache.Upload). It asks the
// remote first (HEAD) and never sends an entry or a blob the remote holds.
// Blobs go before their entry, so a reader never finds an entry whose blobs
// are missing.
//
// Parameters: ctx stops the push between entries; local is the local cache;
// remote is the remote cache (Connect); opts change the work.
// Returns: what happened; ErrReadOnly for a write on a read-only remote, the
// store's error when it refuses the credentials, ctx.Err() when stopped.
// Other failures are counted per entry and do not stop the push.
func Push(ctx context.Context, local taskcache.Storage, remote Remote, opts PushOptions) (PushResult, error) {
	result := PushResult{Store: remote.URL, DryRun: opts.DryRun, Steps: []PushStep{}}
	if remote.ReadOnly && !opts.DryRun {
		return result, ErrReadOnly
	}
	keys, _, err := localEntries(local, 0)
	if err != nil {
		return result, err
	}
	sort.Strings(keys)
	result.Entries = len(keys)
	parallel := opts.Parallel
	if parallel <= 0 {
		parallel = 4
	}
	var (
		mu      sync.Mutex
		refused error
		wg      sync.WaitGroup
		work    = make(chan string)
	)
	for range parallel {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for key := range work {
				step, err := pushEntry(local, remote, key, opts.DryRun)
				mu.Lock()
				if remotecache.Refused(err) {
					// A refusal stops the push and is said once, by the caller.
					if refused == nil {
						refused = err
					}
					mu.Unlock()
					continue
				}
				result.record(step)
				if opts.Progress != nil {
					opts.Progress(step)
				}
				mu.Unlock()
			}
		}()
	}
feed:
	for _, key := range keys {
		mu.Lock()
		stop := refused != nil
		mu.Unlock()
		if stop {
			break
		}
		select {
		case <-ctx.Done():
			break feed
		case work <- key:
		}
	}
	close(work)
	wg.Wait()
	sort.Slice(result.Steps, func(i, j int) bool { return result.Steps[i].Key < result.Steps[j].Key })
	if refused != nil {
		return result, refused
	}
	return result, ctx.Err()
}

// record counts a step in the result.
func (r *PushResult) record(step PushStep) {
	r.Steps = append(r.Steps, step)
	switch step.Outcome {
	case StepUploaded, StepWouldUpload:
		r.Uploaded++
		r.Blobs += step.Blobs
		r.Bytes += step.Bytes
	case StepPresent:
		r.Present++
	case StepFailed:
		r.Failed++
	}
}

// pushEntry sends one entry and the blobs it names that the remote lacks.
// The error is the store's, for the caller to tell a refusal; the step says
// what happened in every case.
func pushEntry(local taskcache.Storage, remote Remote, key string, dryRun bool) (PushStep, error) {
	step := PushStep{Key: key}
	fail := func(err error) (PushStep, error) {
		step.Outcome, step.Error = StepFailed, err.Error()
		return step, err
	}
	data, err := read(local, taskcache.AC, key)
	if err != nil {
		return fail(err)
	}
	var entry taskcache.Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return fail(errors.New("the local entry is damaged: run smith cache verify --fix"))
	}
	step.Task = entry.Task
	has, err := remote.Store.Has(remotecache.ObjectName(taskcache.AC, key))
	if err != nil {
		return fail(err)
	}
	if has {
		step.Outcome = StepPresent
		return step, nil
	}
	digests := []string{entry.Log}
	for _, o := range entry.Outputs {
		digests = append(digests, o.Digest)
	}
	seen := map[string]bool{}
	for _, digest := range digests {
		if digest == "" || seen[digest] {
			continue
		}
		seen[digest] = true
		has, err := remote.Store.Has(remotecache.ObjectName(taskcache.CAS, digest))
		if err != nil {
			return fail(err)
		}
		if has {
			continue
		}
		blob, err := read(local, taskcache.CAS, digest)
		if err != nil {
			return fail(err)
		}
		if !dryRun {
			if err := remotecache.Upload(remote.Store, remote.Secret, taskcache.CAS, digest, blob); err != nil {
				return fail(err)
			}
		}
		step.Blobs++
		step.Bytes += int64(len(blob))
	}
	step.Bytes += int64(len(data))
	if dryRun {
		step.Outcome = StepWouldUpload
		return step, nil
	}
	if err := remotecache.Upload(remote.Store, remote.Secret, taskcache.AC, key, data); err != nil {
		return fail(err)
	}
	step.Outcome = StepUploaded
	return step, nil
}

// read returns a local entry's bytes.
func read(local taskcache.Storage, kind taskcache.Kind, key string) ([]byte, error) {
	r, err := local.Get(kind, key)
	if errors.Is(err, taskcache.ErrMiss) {
		return nil, errors.New("the local cache lacks " + string(kind) + "/" + key + ": run smith cache verify --fix")
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}
