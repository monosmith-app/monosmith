package npmbinaries

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/npmpropagation"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func archive(t *testing.T, root, name, file string) string {
	t.Helper()
	f, err := os.Create(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create(file)
	_, _ = w.Write([]byte("binary " + file))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return name
}

func release(t *testing.T, root string) Release {
	return Release{
		Names: []string{"@acme/cli", "acme"}, Binary: "acme", Version: "5.0.0", OutDir: "out",
		Archives: []Archive{
			{Path: archive(t, root, "acme_windows_amd64.zip", "acme.exe"), OS: "windows", Arch: "amd64"},
			{Path: archive(t, root, "acme_linux_amd64.zip", "acme"), OS: "linux", Arch: "amd64"},
			{Path: archive(t, root, "acme_plan9_amd64.zip", "acme"), OS: "plan9", Arch: "amd64"},
		},
	}
}

// fakeNPM replaces the npm CLI for one test.
func fakeNPM(t *testing.T, f func(dir string, args ...string) (string, string, error)) {
	t.Helper()
	original := npm
	npm = f
	t.Cleanup(func() { npm = original })
}

// fakeRegistry is an npm registry in memory. A package it has answers view; a
// publish adds the package unless lose names it (npm said yes, the registry
// never got it). It never sleeps: the pause after a publish is a no-op.
type fakeRegistry struct {
	// have holds "name@version" of every package the registry serves.
	have map[string]bool
	lose map[string]bool
	// tags are the dist-tags of each package name.
	tags map[string]map[string]string
	// ran lists every npm call, space joined.
	ran []string
	// published lists "<dir> <args>" of each publish.
	published []string
}

func newRegistry(t *testing.T) *fakeRegistry {
	t.Helper()
	reg := &fakeRegistry{have: map[string]bool{}, lose: map[string]bool{}, tags: map[string]map[string]string{}}
	fakeNPM(t, reg.run)
	originalPause := pause
	pause = func(time.Duration) {}
	t.Cleanup(func() { pause = originalPause })
	return reg
}

func (r *fakeRegistry) run(dir string, args ...string) (string, string, error) {
	r.ran = append(r.ran, strings.Join(args, " "))
	switch args[0] {
	case "view":
		if len(args) > 2 && args[2] == "dist-tags" {
			raw, _ := json.Marshal(r.tags[args[1]])
			return string(raw), "", nil
		}
		if r.have[args[1]] {
			return args[1][strings.LastIndex(args[1], "@")+1:] + "\n", "", nil
		}
	case "publish":
		raw, _ := os.ReadFile(filepath.Join(dir, "package.json"))
		var m struct{ Name, Version string }
		_ = json.Unmarshal(raw, &m)
		r.published = append(r.published, filepath.Base(dir)+" "+strings.Join(args, " "))
		if !r.lose[m.Name] {
			r.have[m.Name+"@"+m.Version] = true
			if i := slices.Index(args, "--tag"); i >= 0 {
				if r.tags[m.Name] == nil {
					r.tags[m.Name] = map[string]string{}
				}
				r.tags[m.Name][args[i+1]] = m.Version
			}
		}
	case "dist-tag":
		name := strings.TrimSuffix(args[2], "@"+args[2][strings.LastIndex(args[2], "@")+1:])
		if args[1] == "add" {
			r.tags[name][args[3]] = args[2][strings.LastIndex(args[2], "@")+1:]
		} else {
			delete(r.tags[args[2]], args[3])
		}
	}
	return "", "", nil
}

// fakeAwait replaces the wait for the platform packages with a fixed answer.
func fakeAwait(t *testing.T, ready bool, waitedFor *[]string) {
	t.Helper()
	original := awaitPlatforms
	awaitPlatforms = func(_ string, _ Release, names []string) npmpropagation.Outcome {
		if waitedFor != nil {
			*waitedFor = names
		}
		return npmpropagation.Outcome{Ready: ready}
	}
	t.Cleanup(func() { awaitPlatforms = original })
}

func TestPublishSendsPlatformPackagesBeforeWrappersAndSkipsWhatTheRegistryHas(t *testing.T) {
	root := t.TempDir()
	fakeAwait(t, true, nil)
	t.Setenv("NODE_AUTH_TOKEN", "secret")
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
	reg := newRegistry(t)
	reg.have["@acme/cli-linux-x64@5.0.0"] = true

	res, err := Publish(root, release(t, root), false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"acme-cli-win32-x64 publish --access public",
		"acme-cli publish --access public",
		"acme publish --access public",
	}
	if !slices.Equal(reg.published, want) {
		t.Errorf("published\n%v\nwant\n%v", reg.published, want)
	}
	if !slices.Equal(res.Existing, []string{"@acme/cli-linux-x64@5.0.0"}) || !slices.Equal(res.Skipped, []string{"plan9/amd64"}) {
		t.Errorf("existing %v skipped %v", res.Existing, res.Skipped)
	}
	bin, err := os.ReadFile(filepath.Join(root, "out", "acme-cli-win32-x64", "bin", "acme.exe"))
	if err != nil || string(bin) != "binary acme.exe" {
		t.Errorf("platform binary = %q, %v", bin, err)
	}
	if _, err := os.Stat(filepath.Join(root, "out", "acme", "bin", "acme.js")); err != nil {
		t.Error("the wrapper has no shim")
	}
	if rc, _ := os.ReadFile(filepath.Join(root, "out", "acme", ".npmrc")); !strings.Contains(string(rc), "${NODE_AUTH_TOKEN}") {
		t.Errorf(".npmrc = %q", rc)
	}
}

func TestPublishAddsProvenanceAndTagsPrereleases(t *testing.T) {
	root := t.TempDir()
	fakeAwait(t, true, nil)
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "https://token")
	reg := newRegistry(t)
	r := release(t, root)
	r.Version, r.Repository = "5.0.0-rc.1", "https://github.com/acme/cli"
	if _, err := Publish(root, r, false); err != nil {
		t.Fatal(err)
	}
	if got := reg.published[len(reg.published)-1]; got != "acme publish --access public --tag next --provenance" {
		t.Errorf("args = %s", got)
	}
}

func TestDryRunNamesThePackagesWithoutWriting(t *testing.T) {
	root := t.TempDir()
	fakeNPM(t, func(string, ...string) (string, string, error) {
		t.Error("npm ran on a dry run")
		return "", "", nil
	})
	res, err := Publish(root, release(t, root), true)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"@acme/cli-linux-x64@5.0.0", "@acme/cli-win32-x64@5.0.0", "@acme/cli@5.0.0", "acme@5.0.0"}
	if !slices.Equal(res.Published, want) {
		t.Errorf("published %v", res.Published)
	}
	if _, err := os.Stat(filepath.Join(root, "out")); !os.IsNotExist(err) {
		t.Error("a dry run wrote packages")
	}
}

func TestWrappersTakeThePendingTagAndAreReturnedAsMovesWhenThePlatformPackagesAreNotFetchable(t *testing.T) {
	for _, c := range []struct{ version, tag, moveTo string }{
		{"5.0.0-rc.1", "next", "next"}, {"5.0.0", "", "latest"},
	} {
		t.Run(c.version, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
			var waitedFor []string
			fakeAwait(t, false, &waitedFor)
			reg := newRegistry(t)
			r := release(t, root)
			r.Version = c.version
			r.PropagationWait = 15 * time.Minute
			res, err := Publish(root, r, false)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(waitedFor, []string{"@acme/cli-linux-x64", "@acme/cli-win32-x64"}) {
				t.Errorf("waited for %v", waitedFor)
			}
			wantPlatform := "publish --access public"
			if c.tag != "" {
				wantPlatform += " --tag next"
			}
			want := []string{"acme-cli-linux-x64 " + wantPlatform, "acme-cli-win32-x64 " + wantPlatform, "acme-cli publish --access public --tag pending", "acme publish --access public --tag pending"}
			if !slices.Equal(reg.published, want) {
				t.Errorf("publishes %v, want %v", reg.published, want)
			}
			var commands []string
			for _, m := range res.PendingMoves {
				commands = append(commands, m.Command())
			}
			wantTags := []string{"npm dist-tag add @acme/cli@" + c.version + " " + c.moveTo, "npm dist-tag add acme@" + c.version + " " + c.moveTo}
			if !slices.Equal(commands, wantTags) {
				t.Errorf("pending moves %v, want %v", commands, wantTags)
			}
		})
	}
}

func TestPublishNeverWaitsForTheRegistry(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
	original := awaitPlatforms
	var waits []time.Duration
	awaitPlatforms = func(_ string, r Release, _ []string) npmpropagation.Outcome {
		waits = append(waits, r.PropagationWait)
		return npmpropagation.Outcome{}
	}
	t.Cleanup(func() { awaitPlatforms = original })
	newRegistry(t)
	r := release(t, root)
	r.PropagationWait = time.Hour
	if _, err := Publish(root, r, false); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(waits, []time.Duration{0}) {
		t.Errorf("publish asked for waits %v: it probes once and the final phase waits", waits)
	}
}

func TestNoProbeWhenEveryWrapperIsAlreadyPublished(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
	original := awaitPlatforms
	awaitPlatforms = func(string, Release, []string) npmpropagation.Outcome {
		t.Error("probed although nothing is left to publish")
		return npmpropagation.Outcome{}
	}
	t.Cleanup(func() { awaitPlatforms = original })
	reg := newRegistry(t)
	for _, name := range []string{"@acme/cli-linux-x64", "@acme/cli-win32-x64", "@acme/cli", "acme"} {
		reg.have[name+"@5.0.0"] = true
	}
	if _, err := Publish(root, release(t, root), false); err != nil {
		t.Fatal(err)
	}
	if len(reg.published) != 0 {
		t.Errorf("published %v", reg.published)
	}
}

func TestAPackageNpmHasNotServedYetDoesNotFailThePublish(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
	fakeAwait(t, false, nil)
	reg := newRegistry(t)
	reg.lose["@acme/cli-linux-x64"] = true
	var pauses int
	pause = func(time.Duration) { pauses++ }
	res, err := Publish(root, release(t, root), false)
	if err != nil {
		t.Fatalf("a slow registry is not an error: %v", err)
	}
	if pauses < readBackChecks-1 {
		t.Errorf("paused %d times, want at least %d: the read-back is quick", pauses, readBackChecks-1)
	}
	if !slices.Contains(res.Published, "@acme/cli-linux-x64@5.0.0") || len(res.PendingMoves) != 2 {
		t.Errorf("the release goes on and the wrappers wait under pending: %+v", res)
	}
	if got := reg.published[len(reg.published)-1]; !strings.HasSuffix(got, "--tag pending") {
		t.Errorf("last publish %q", got)
	}
}

// timedRegistry serves each package from a time on, by the fake clock.
type timedRegistry struct {
	clock *fakeClock
	from  map[string]time.Duration
	start time.Time
}

func (r *timedRegistry) Lookup(name, _ string) (bool, string) {
	from, ok := r.from[name]
	if !ok || r.clock.now.Sub(r.start) < from {
		return false, ""
	}
	return true, "https://registry.example/" + name
}
func (r *timedRegistry) Fetchable(string) bool { return true }

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time        { return c.now }
func (c *fakeClock) Sleep(d time.Duration) { c.now = c.now.Add(d) }

// finalizeWith runs Finalize with the real wait on a fake clock and registry.
func finalizeWith(t *testing.T, from map[string]time.Duration) (Finalized, *fakeRegistry, []string) {
	t.Helper()
	clock := &fakeClock{now: time.Unix(1000, 0)}
	timed := &timedRegistry{clock: clock, from: from, start: clock.now}
	original := awaitPlatforms
	var progress []string
	awaitPlatforms = func(_ string, r Release, names []string) npmpropagation.Outcome {
		return npmpropagation.Await(timed, clock, npmpropagation.Config{Max: r.PropagationWait, Progress: func(l string) { progress = append(progress, l) }}, names, r.Version)
	}
	t.Cleanup(func() { awaitPlatforms = original })
	reg := newRegistry(t)
	reg.tags["acme"] = map[string]string{"pending": "5.0.0"}
	r := release(t, t.TempDir())
	r.PropagationWait = 15 * time.Minute
	got, err := Finalize(t.TempDir(), r, []protocol.TagMove{{Package: "acme", Version: "5.0.0", Tag: "next", Awaits: []string{"@acme/cli-linux-x64"}}})
	if err != nil {
		t.Fatal(err)
	}
	return got, reg, progress
}

func TestAPackageThatAppearsAtMinuteEightGetsItsTagMoved(t *testing.T) {
	got, reg, progress := finalizeWith(t, map[string]time.Duration{"@acme/cli-linux-x64": 8 * time.Minute})
	if !slices.Equal(got.Moved, []string{"acme@5.0.0 next"}) || len(got.Pending) != 0 || reg.tags["acme"]["next"] != "5.0.0" {
		t.Errorf("finalized %+v tags %v", got, reg.tags)
	}
	if len(progress) == 0 || !strings.Contains(progress[0], "waiting for npm") {
		t.Errorf("progress %v", progress)
	}
}

func TestAPackageThatNeverAppearsLeavesTheTagPendingAndSaysHowToFinish(t *testing.T) {
	got, reg, _ := finalizeWith(t, map[string]time.Duration{})
	if len(got.Moved) != 0 || len(got.Pending) != 1 || reg.tags["acme"]["next"] != "" || reg.tags["acme"]["pending"] != "5.0.0" {
		t.Errorf("finalized %+v tags %v", got, reg.tags)
	}
	note := strings.ToLower(strings.Join(got.Notes, "\n"))
	for _, want := range []string{"@acme/cli-linux-x64@5.0.0", "still does not serve it after 15m0s", "run the release again", "npm dist-tag add acme@5.0.0 next"} {
		if !strings.Contains(note, strings.ToLower(want)) {
			t.Errorf("note lacks %q:\n%s", want, note)
		}
	}
}

func TestAWrapperAlreadyUnderPendingIsReturnedAsAMoveOnARerun(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
	fakeAwait(t, true, nil)
	reg := newRegistry(t)
	for _, name := range []string{"@acme/cli", "acme"} {
		reg.have[name+"@5.0.0"] = true
		reg.tags[name] = map[string]string{"pending": "5.0.0"}
	}
	res, err := Publish(root, release(t, root), false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(reg.published, []string{"acme-cli-linux-x64 publish --access public", "acme-cli-win32-x64 publish --access public"}) {
		t.Errorf("a re-run publishes only what is missing: %v", reg.published)
	}
	if len(res.PendingMoves) != 2 || res.PendingMoves[0].Package != "@acme/cli" {
		t.Errorf("pending moves %v", res.PendingMoves)
	}
}

func TestFinalizeWaitsThenMovesTheTagAndKeepsTheLateOnePending(t *testing.T) {
	root := t.TempDir()
	reg := newRegistry(t)
	reg.tags["acme"] = map[string]string{"pending": "5.0.0"}
	reg.tags["late"] = map[string]string{"pending": "5.0.0"}
	original := awaitPlatforms
	var clock []time.Duration
	awaitPlatforms = func(_ string, r Release, names []string) npmpropagation.Outcome {
		clock = append(clock, r.PropagationWait)
		if names[0] == "late-linux" {
			return npmpropagation.Outcome{Missing: names, Waited: 15 * time.Minute}
		}
		return npmpropagation.Outcome{Ready: true}
	}
	t.Cleanup(func() { awaitPlatforms = original })
	r := release(t, root)
	r.PropagationWait = 15 * time.Minute
	moves := []protocol.TagMove{
		{Package: "acme", Version: "5.0.0", Tag: "next", Awaits: []string{"acme-linux"}},
		{Package: "late", Version: "5.0.0", Tag: "next", Awaits: []string{"late-linux"}},
	}
	got, err := Finalize(root, r, moves)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Moved, []string{"acme@5.0.0 next"}) || len(got.Pending) != 1 || got.Pending[0].Package != "late" {
		t.Errorf("finalized %+v", got)
	}
	if reg.tags["acme"]["next"] != "5.0.0" || reg.tags["acme"]["pending"] != "" || reg.tags["late"]["next"] != "" {
		t.Errorf("tags %v", reg.tags)
	}
	if len(got.Notes) != 1 || !strings.Contains(got.Notes[0], "npm dist-tag add late@5.0.0 next") || !slices.Equal(clock, []time.Duration{15 * time.Minute, 15 * time.Minute}) {
		t.Errorf("notes %v waits %v", got.Notes, clock)
	}
}

func TestIncompleteSeesWhatARerunStillHasToDo(t *testing.T) {
	root := t.TempDir()
	reg := newRegistry(t)
	pairs := [][2]string{{"linux", "amd64"}, {"windows", "amd64"}}
	original := fetchable
	fetchable = func(_, name, version string) bool { return reg.have[name+"@"+version] }
	t.Cleanup(func() { fetchable = original })
	for _, name := range []string{"@acme/cli-win32-x64", "@acme/cli", "acme"} {
		reg.have[name+"@5.0.0"] = true
	}
	reg.tags["acme"] = map[string]string{"pending": "5.0.0"}
	missing, moves := Incomplete(root, release(t, root), pairs)
	if !slices.Equal(missing, []string{"@acme/cli-linux-x64"}) || len(moves) != 1 || moves[0].Package != "acme" || !slices.Equal(moves[0].Awaits, []string{"@acme/cli-linux-x64", "@acme/cli-win32-x64"}) {
		t.Errorf("missing %v moves %v", missing, moves)
	}
	reg.have["@acme/cli-linux-x64@5.0.0"] = true
	delete(reg.tags["acme"], "pending")
	if missing, moves := Incomplete(root, release(t, root), pairs); len(missing) != 0 || len(moves) != 0 {
		t.Errorf("a finished release has nothing to do: %v %v", missing, moves)
	}
}

func TestAProvenanceConflictSaysTheVersionIsLostAndHowToReleaseTheNext(t *testing.T) {
	fakeNPM(t, func(dir string, args ...string) (string, string, error) {
		if args[0] == "view" {
			return "", "", errors.New("not found")
		}
		return "", "npm error code TLOG_CREATE_ENTRY_ERROR\nnpm error (409) an equivalent entry already exists in the transparency log", errors.New("exit status 1")
	})
	_, err := publishOne(t.TempDir(), "@acme/cli-linux-x64", Release{Version: "1.0.0"}, "")
	if err == nil {
		t.Fatal("a failed publish must be an error")
	}
	for _, want := range []string{"can no longer be published with provenance", `version(<project>)[<next version>]`, "TLOG_CREATE_ENTRY_ERROR"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error lacks %q: %v", want, err)
		}
	}
}

func TestAPendingTagThatTheRegistryWillNotRemoveIsNotUnfinishedWork(t *testing.T) {
	fakeNPM(t, func(dir string, args ...string) (string, string, error) {
		return `{"next":"1.0.0","pending":"1.0.0"}`, "", nil
	})
	if hasPendingTag(t.TempDir(), "acme", "1.0.0", "next") {
		t.Error("pending equal to next means the tag was moved")
	}
	if !hasPendingTag(t.TempDir(), "acme", "1.0.0", "latest") {
		t.Error("pending without the real tag is unfinished")
	}
}
