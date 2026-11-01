package npmbinaries

import (
	"archive/zip"
	"cmp"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

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
}

// Result says what happened to each package, in publishing order.
type Result struct {
	Published, Existing []string
	// Skipped are the archives no npm platform matches.
	Skipped []string
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
	for _, b := range order {
		if dryRun {
			res.Published = append(res.Published, b.name+"@"+r.Version)
			continue
		}
		published, err := publishOne(b.dir, b.name, r)
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

// publishOne publishes a package directory unless the registry has its
// version. With NODE_AUTH_TOKEN set, the directory gets an .npmrc using it
// (npm never packs an .npmrc); otherwise npm's own login is used. On GitHub
// Actions with an OIDC token the package carries provenance.
func publishOne(dir, name string, r Release) (bool, error) {
	// An unpublished package is a 404, an unpublished version prints nothing.
	if out, _, err := npm(dir, "view", name+"@"+r.Version, "version"); err == nil && strings.TrimSpace(out) == r.Version {
		return false, nil
	}
	if os.Getenv("NODE_AUTH_TOKEN") != "" {
		if err := os.WriteFile(filepath.Join(dir, ".npmrc"), []byte("//registry.npmjs.org/:_authToken=${NODE_AUTH_TOKEN}\n"), 0o600); err != nil {
			return false, err
		}
	}
	args := []string{"publish", "--access", "public"}
	if strings.Contains(r.Version, "-") {
		// npm refuses a prerelease without a tag, so it never becomes latest.
		args = append(args, "--tag", cmp.Or(r.Tag, "next"))
	}
	if os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL") != "" && r.Repository != "" {
		args = append(args, "--provenance")
	}
	if _, stderr, err := npm(dir, args...); err != nil {
		return false, fmt.Errorf("npm publish %s: %w\n%s", name, err, strings.TrimSpace(stderr))
	}
	return true, nil
}
