package npmbinaries

import (
	"archive/zip"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/npmpropagation"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Archive is one platform's release archive, holding the binary at its root.
type Archive struct {
	// Path is workspace-relative.
	Path     string
	OS, Arch string
}

// Release is an app's release as npm packages.
type Release struct {
	// Names are the packages people install; the first also names the
	// platform packages (<first>-<os>-<cpu>).
	Names   []string
	Binary  string
	Version string
	// Description, License and Homepage describe every package.
	Description, License, Homepage string
	// Repository is the https URL of the source repository; npm provenance
	// checks the packages against it.
	Repository string
	Archives   []Archive
	// Tag is the dist-tag a prerelease publishes under (a release channel's
	// name). Default: next.
	Tag string
	// OutDir is the workspace-relative directory the packages are written to.
	OutDir string
	// PropagationWait is the longest Finalize waits for the platform packages
	// to be fetchable from the registry; 0 probes once. Publish never waits.
	PropagationWait time.Duration
	// Progress receives the progress lines of that wait; nil prints nothing.
	Progress func(line string)
}

// Result says what happened to each package, in publishing order.
type Result struct {
	Published, Existing []string
	// Skipped are the archives no npm platform matches.
	Skipped []string
	// PendingMoves are the dist-tags of the wrappers that sit under "pending"
	// because the platform packages were not fetchable yet; Finalize moves
	// them once every project has published.
	PendingMoves []protocol.TagMove
}

// npm runs the npm CLI and returns what it printed; tests replace it.
var npm = func(dir string, args ...string) (stdout, stderr string, err error) {
	return toolrun.Capture(dir, append([]string{"npm"}, args...))
}

// Publish writes a package per platform holding its binary, then the
// wrapper packages depending on them all, and publishes them in that order,
// so no wrapper ever points at a missing platform package. A version the
// registry already has is left alone, so a re-run completes a release. A dry
// run only names the packages. It never waits for the registry: a wrapper
// whose platform packages are not fetchable at one quick probe takes the
// dist-tag "pending" and is returned in PendingMoves. A package the registry
// does not have right after npm said it published it is an error.
func Publish(root string, r Release, dryRun bool) (Result, error) {
	var res Result
	if len(r.Names) == 0 {
		return res, nil
	}
	platforms := map[string]string{}
	type built struct{ name, dir string }
	var order []built
	sort.Slice(r.Archives, func(i, j int) bool { return r.Archives[i].Path < r.Archives[j].Path })
	for _, a := range r.Archives {
		name, ok := PlatformPackage(r.Names[0], a.OS, a.Arch)
		if !ok {
			res.Skipped = append(res.Skipped, a.OS+"/"+a.Arch)
			continue
		}
		platforms[npmOS[a.OS]+"-"+npmCPU[a.Arch]] = name
		dir := filepath.Join(root, filepath.FromSlash(r.OutDir), folder(name))
		order = append(order, built{name, dir})
		if dryRun {
			continue
		}
		if err := writePlatform(root, r, a, name, dir); err != nil {
			return res, err
		}
	}
	if len(platforms) == 0 {
		return res, fmt.Errorf("no archive for a platform npm knows")
	}
	for _, name := range r.Names {
		dir := filepath.Join(root, filepath.FromSlash(r.OutDir), folder(name))
		order = append(order, built{name, dir})
		if dryRun {
			continue
		}
		if err := writeWrapper(r, name, platforms, dir); err != nil {
			return res, err
		}
	}
	// The wrappers are published under the dist-tag "pending" unless every
	// platform package can be fetched the way a person fetches it at one quick
	// probe, so no dist-tag points at a launcher with no binary.
	tag := ""
	awaits := platformNames(platforms)
	for i, b := range order {
		if dryRun {
			res.Published = append(res.Published, b.name+"@"+r.Version)
			continue
		}
		wrapper := i >= len(order)-len(r.Names)
		if i == len(order)-len(r.Names) && slices.ContainsFunc(order[i:], func(w built) bool { return !registryHas(w.dir, w.name, r.Version) }) {
			once := r
			once.PropagationWait = 0
			if out := awaitPlatforms(b.dir, once, awaits); !out.Ready {
				tag = pendingTag
			}
		}
		published, err := publishOne(b.dir, b.name, r, cmp.Or(tag, intendedTag(r)))
		if err != nil {
			return res, err
		}
		if wrapper && ((published && tag == pendingTag) || (!published && hasPendingTag(b.dir, b.name, r.Version, cmp.Or(intendedTag(r), "latest")))) {
			res.PendingMoves = append(res.PendingMoves, protocol.TagMove{Package: b.name, Version: r.Version, Tag: cmp.Or(intendedTag(r), "latest"), Awaits: awaits})
		}
		if published {
			res.Published = append(res.Published, b.name+"@"+r.Version)
		} else {
			res.Existing = append(res.Existing, b.name+"@"+r.Version)
		}
	}
	return res, nil
}

// Finalized says what the last phase of a release did.
type Finalized struct {
	// Moved are the dist-tags moved ("name@version tag").
	Moved []string
	// Pending are the moves whose packages did not become fetchable in time.
	Pending []protocol.TagMove
	// Notes say why a move is pending and the command that does it by hand.
	Notes []string
}

// Finalize is the last phase of a release: for each move it waits, for up to
// r.PropagationWait, until the packages the move awaits are fetchable, then
// moves the dist-tag and removes "pending". It runs after every project has
// published, so the wait blocks nothing else.
//
// Parameters: root is the workspace; r carries the wait, the progress sink and
// the output directory; moves are the dist-tags to move.
// Returns: what was moved and what is still pending.
// Errors: npm refuses to move a tag.
func Finalize(root string, r Release, moves []protocol.TagMove) (Finalized, error) {
	var out Finalized
	for _, m := range moves {
		dir := filepath.Join(root, filepath.FromSlash(r.OutDir), folder(m.Package))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return out, err
		}
		if err := writeAuth(dir); err != nil {
			return out, err
		}
		step := r
		step.Version = m.Version
		if got := awaitPlatforms(dir, step, m.Awaits); !got.Ready {
			out.Pending = append(out.Pending, m)
			out.Notes = append(out.Notes, NotServedNote(m, got))
			continue
		}
		if _, stderr, err := npm(dir, "dist-tag", "add", m.Package+"@"+m.Version, m.Tag); err != nil {
			return out, fmt.Errorf("npm dist-tag add %s@%s %s: %w\n%s", m.Package, m.Version, m.Tag, err, strings.TrimSpace(stderr))
		}
		// A leftover "pending" would make a re-run look for work.
		if _, stderr, err := npm(dir, "dist-tag", "rm", m.Package, pendingTag); err != nil {
			out.Notes = append(out.Notes, fmt.Sprintf("the dist-tag %q of %s could not be removed: %v %s. Remove it by hand: npm dist-tag rm %s %s", pendingTag, m.Package, err, strings.TrimSpace(stderr), m.Package, pendingTag))
		}
		out.Moved = append(out.Moved, m.Package+"@"+m.Version+" "+m.Tag)
	}
	return out, nil
}

// NotServedNote says why a dist-tag was not moved when the wait ended: npm said
// it published the packages, the registry still does not serve them. It names
// each package and says how to finish, by hand or by running the release again.
func NotServedNote(m protocol.TagMove, out npmpropagation.Outcome) string {
	missing := "the platform packages"
	if len(out.Missing) > 0 {
		var named []string
		for _, n := range out.Missing {
			named = append(named, n+"@"+m.Version)
		}
		missing = strings.Join(named, ", ")
	}
	return fmt.Sprintf("the dist-tag %q of %s was not moved: npm said it published %s but the registry still does not serve it after %s. "+
		"Run the release again: it republishes what is missing and moves the tag. Or, when npm serves it, run: %s",
		m.Tag, m.Package, missing, out.Waited.Round(time.Second), m.Command())
}

// Incomplete reports what a re-run of an already released version still has
// to do: the packages the registry does not have, and the wrappers still under
// "pending". With none of either, the release of the packages is done.
//
// Parameters: root is the workspace; r names the packages and the version; pairs
// are the "os/arch" platforms the release publishes.
// Returns: the missing packages and the moves of the pending wrappers.
func Incomplete(root string, r Release, pairs [][2]string) (missing []string, moves []protocol.TagMove) {
	var awaits []string
	for _, p := range pairs {
		if name, ok := PlatformPackage(r.Names[0], p[0], p[1]); ok {
			awaits = append(awaits, name)
		}
	}
	slices.Sort(awaits)
	for _, name := range append(slices.Clone(awaits), r.Names...) {
		if !fetchable(root, name, r.Version) {
			missing = append(missing, name)
		}
	}
	for _, name := range r.Names {
		if hasPendingTag(root, name, r.Version, cmp.Or(intendedTag(r), "latest")) {
			moves = append(moves, protocol.TagMove{Package: name, Version: r.Version, Tag: cmp.Or(intendedTag(r), "latest"), Awaits: awaits})
		}
	}
	return missing, moves
}

// hasPendingTag reports whether the dist-tag "pending" of name is version and
// the real dist-tag (tag) is not: the wrapper was published and its real
// dist-tag was never moved. A "pending" left behind because the registry would
// not remove it (npm answers E403 to some credentials) is not unfinished work.
func hasPendingTag(dir, name, version, tag string) bool {
	out, _, err := npm(dir, "view", name, "dist-tags", "--json", "--prefer-online")
	if err != nil {
		return false
	}
	var tags map[string]string
	return json.Unmarshal([]byte(out), &tags) == nil && tags[pendingTag] == version && tags[tag] != version
}

// folder is a package's directory name: @acme/cli → acme-cli.
func folder(name string) string {
	return strings.ReplaceAll(strings.TrimPrefix(name, "@"), "/", "-")
}

func writePlatform(root string, r Release, a Archive, name, dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		return err
	}
	file := executable(r.Binary, a.OS)
	if err := extract(filepath.Join(root, filepath.FromSlash(a.Path)), file, filepath.Join(dir, "bin", file)); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "package.json"), []byte(PlatformManifest(r, name, a.OS, a.Arch)), 0o644)
}

func writeWrapper(r Release, name string, platforms map[string]string, dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		return err
	}
	files := map[string]string{
		"package.json":                   WrapperManifest(r, name, platforms),
		"README.md":                      Readme(r, name),
		path.Join("bin", r.Binary+".js"): Shim(r.Binary, platforms),
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), []byte(content), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// extract copies one file out of a zip archive, executable.
func extract(archive, name, to string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer func() { _ = zr.Close() }()
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		in, err := f.Open()
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		out, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	}
	return fmt.Errorf("%s has no %s", filepath.Base(archive), name)
}

// pendingTag is the dist-tag the wrappers take while the platform packages are
// not fetchable from the registry yet.
const pendingTag = "pending"

// intendedTag is the dist-tag the release moves: a prerelease's channel (npm
// refuses a prerelease without a tag, so it never becomes latest), and none
// for a stable version, which takes latest.
func intendedTag(r Release) string {
	if strings.Contains(r.Version, "-") {
		return cmp.Or(r.Tag, "next")
	}
	return ""
}

// registryHas reports whether the registry has name@version. An unpublished
// package is a 404, an unpublished version prints nothing.
func registryHas(dir, name, version string) bool {
	out, _, err := npm(dir, "view", name+"@"+version, "version")
	return err == nil && strings.TrimSpace(out) == version
}

func platformNames(platforms map[string]string) []string {
	return slices.Sorted(maps.Values(platforms))
}

// awaitPlatforms waits for the platform packages to be fetchable from the
// registry (see npmpropagation.Await); tests replace it with a fake registry.
var awaitPlatforms = func(dir string, r Release, names []string) npmpropagation.Outcome {
	reg := npmpropagation.NPM{Dir: dir, Output: func(dir string, argv []string) (string, error) {
		out, _, err := npm(dir, argv[1:]...)
		return out, err
	}}
	return npmpropagation.Await(reg, npmpropagation.RealClock{}, npmpropagation.Config{Max: r.PropagationWait, Progress: r.Progress}, names, r.Version)
}

// readBackChecks and readBackDelay bound the quick read-back after a publish;
// pause is replaced by tests so they never sleep.
const (
	readBackChecks = 3
	readBackDelay  = 2 * time.Second
)

// fetchable reports whether a person can fetch name@version now: npm view
// finds it and a HEAD request of its tarball answers. A package npm never
// stored fails both; tests replace it.
var fetchable = func(dir, name, version string) bool {
	reg := npmpropagation.NPM{Dir: dir, Output: func(dir string, argv []string) (string, error) {
		out, _, err := npm(dir, argv[1:]...)
		return out, err
	}}
	found, tarball := reg.Lookup(name, version)
	return found && tarball != "" && reg.Fetchable(tarball)
}

var pause = time.Sleep

// writeAuth gives a package directory an .npmrc using NODE_AUTH_TOKEN, when set
// (npm never packs an .npmrc); otherwise npm's own login is used.
func writeAuth(dir string) error {
	if os.Getenv("NODE_AUTH_TOKEN") == "" {
		return nil
	}
	return os.WriteFile(filepath.Join(dir, ".npmrc"), []byte("//registry.npmjs.org/:_authToken=${NODE_AUTH_TOKEN}\n"), 0o600)
}

// publishOne publishes a package directory unless the registry has its
// version, then reads it back a few times. A package that is not there yet is
// not an error: npm takes minutes to serve a new package, so the release goes
// on, the wrappers take "pending" and the last phase waits (see Finalize); only
// the end of that wait calls the package missing. On GitHub Actions with an
// OIDC token the package carries provenance.
func publishOne(dir, name string, r Release, tag string) (bool, error) {
	if registryHas(dir, name, r.Version) {
		return false, nil
	}
	if err := writeAuth(dir); err != nil {
		return false, err
	}
	args := []string{"publish", "--access", "public"}
	if tag != "" {
		args = append(args, "--tag", tag)
	}
	if os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL") != "" && r.Repository != "" {
		args = append(args, "--provenance")
	}
	_, stderr, err := npm(dir, args...)
	if err != nil {
		if strings.Contains(stderr, "TLOG_CREATE_ENTRY_ERROR") {
			return false, fmt.Errorf("npm publish %s: %s\n%s", name, tlogConflictAdvice, strings.TrimSpace(stderr))
		}
		return false, fmt.Errorf("npm publish %s: %w\n%s", name, err, strings.TrimSpace(stderr))
	}
	for i := 0; i < readBackChecks; i++ {
		if i > 0 {
			pause(readBackDelay)
		}
		if registryHas(dir, name, r.Version) {
			break
		}
	}
	return true, nil
}

// tlogConflictAdvice explains the provenance conflict npm reports when an
// earlier publish of the same package stopped after its attestation was
// written to the transparency log and before the registry stored the package:
// the attestation cannot be written twice, so this version is lost for good.
const tlogConflictAdvice = "the transparency log already holds the provenance attestation of this exact package, but the registry never stored it (an earlier publish stopped between the two steps). This version can no longer be published with provenance. Release the next version of the project: git commit --allow-empty -m \"version(<project>)[<next version>]\" and push it"
