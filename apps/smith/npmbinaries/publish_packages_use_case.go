package npmbinaries

import (
	"archive/zip"
	"cmp"
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
	// PropagationWait is the longest the wrappers wait for the platform
	// packages to be fetchable from the registry; 0 probes once.
	PropagationWait time.Duration
	// Progress receives the progress lines of that wait; nil prints nothing.
	Progress func(line string)
}

// Result says what happened to each package, in publishing order.
type Result struct {
	Published, Existing []string
	// Skipped are the archives no npm platform matches.
	Skipped []string
	// PendingTags are the exact commands that move the dist-tag of each
	// wrapper that was published under "pending" because the platform
	// packages were not fetchable in time.
	PendingTags []string
}

// npm runs the npm CLI and returns what it printed; tests replace it.
var npm = func(dir string, args ...string) (stdout, stderr string, err error) {
	return toolrun.Capture(dir, append([]string{"npm"}, args...))
}

// Publish writes a package per platform holding its binary, then the
// wrapper packages depending on them all, and publishes them in that order,
// so no wrapper ever points at a missing platform package. A version the
// registry already has is left alone, so a re-run completes a release. A dry
// run only names the packages.
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
	// The wrappers are published only when every platform package can be
	// fetched the way a person fetches it; else they take the dist-tag
	// "pending", so no dist-tag points at a launcher with no binary.
	tag := ""
	for i, b := range order {
		if dryRun {
			res.Published = append(res.Published, b.name+"@"+r.Version)
			continue
		}
		if i == len(order)-len(r.Names) && slices.ContainsFunc(order[i:], func(w built) bool { return !registryHas(w.dir, w.name, r.Version) }) {
			if out := awaitPlatforms(b.dir, r, platformNames(platforms)); !out.Ready {
				tag = pendingTag
			}
		}
		published, err := publishOne(b.dir, b.name, r, cmp.Or(tag, intendedTag(r)))
		if published && tag == pendingTag {
			res.PendingTags = append(res.PendingTags, fmt.Sprintf("npm dist-tag add %s@%s %s", b.name, r.Version, cmp.Or(intendedTag(r), "latest")))
		}
		if err != nil {
			return res, err
		}
		if published {
			res.Published = append(res.Published, b.name+"@"+r.Version)
		} else {
			res.Existing = append(res.Existing, b.name+"@"+r.Version)
		}
	}
	return res, nil
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

// publishOne publishes a package directory unless the registry has its
// version. With NODE_AUTH_TOKEN set, the directory gets an .npmrc using it
// (npm never packs an .npmrc); otherwise npm's own login is used. On GitHub
// Actions with an OIDC token the package carries provenance.
func publishOne(dir, name string, r Release, tag string) (bool, error) {
	if registryHas(dir, name, r.Version) {
		return false, nil
	}
	if os.Getenv("NODE_AUTH_TOKEN") != "" {
		if err := os.WriteFile(filepath.Join(dir, ".npmrc"), []byte("//registry.npmjs.org/:_authToken=${NODE_AUTH_TOKEN}\n"), 0o600); err != nil {
			return false, err
		}
	}
	args := []string{"publish", "--access", "public"}
	if tag != "" {
		args = append(args, "--tag", tag)
	}
	if os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL") != "" && r.Repository != "" {
		args = append(args, "--provenance")
	}
	if _, stderr, err := npm(dir, args...); err != nil {
		return false, fmt.Errorf("npm publish %s: %w\n%s", name, err, strings.TrimSpace(stderr))
	}
	return true, nil
}
