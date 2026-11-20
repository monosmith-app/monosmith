package releasing

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var placeholder = regexp.MustCompile(`\{(\w+)\}`)

// hostPlatform is the machine's "os/arch"; tests replace it.
var hostPlatform = runtime.GOOS + "/" + runtime.GOARCH

// BuildAssets zips a Node app's build (dist/<project dir>) as its release
// asset for this machine's platform. A Node app cannot be cross-built, since
// what differs per platform is native code npm installed for this machine,
// so it is built on each OS: a native app ("smith": {"native": true}), whose
// native jobs each ask for their own platform.
//
// A dry run names the asset of every platform and adds a note for each one
// this machine cannot build, so the plan stays readable on any machine. A real
// run fails before anything is built when a platform is not this machine's.
//
// Parameters: workspaceRoot is the workspace folder; params carry the project,
// the version, the asset spec and whether this is a dry run.
// Returns the assets with their platforms, or an error when a platform is not
// this machine's (real run), no platform is set, the name pattern is wrong or
// the build is missing.
func BuildAssets(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	spec := *params.Assets
	p := params.Project
	if len(spec.Platforms) == 0 {
		return protocol.PublishResult{}, fmt.Errorf("%s: no platforms to build assets for", p.Name)
	}
	res := protocol.PublishResult{AssetPlatforms: map[string]string{}}
	for _, platform := range spec.Platforms {
		if platform != hostPlatform && !params.DryRun {
			return protocol.PublishResult{}, fmt.Errorf("%s: this machine (%s) cannot build the release asset for %s, because a Node app's asset holds the native code npm installed for this machine. Nothing was tagged. Run the release on %s, or mark the app native (\"smith\": {\"native\": true} in package.json) so a job on each OS builds its own", p.Name, hostPlatform, platform, platform)
		}
		goos, goarch, _ := strings.Cut(platform, "/")
		name, err := assetName(spec, p.Name, params.Version, goos, goarch, "zip")
		if err != nil {
			return protocol.PublishResult{}, err
		}
		asset := path.Join(spec.OutDir, name)
		res.Assets = append(res.Assets, asset)
		res.AssetPlatforms[asset] = platform
		if platform != hostPlatform {
			res.Notes = append(res.Notes, fmt.Sprintf("this machine (%s) cannot build the %s asset, so a release here stops before it tags. Run the release on %s, or mark the app native (\"smith\": {\"native\": true} in package.json)", hostPlatform, platform, platform))
		}
	}
	if params.DryRun {
		return res, nil
	}
	asset := res.Assets[0]
	build := filepath.Join(workspaceRoot, filepath.FromSlash(protocol.BuildOutputDir(p)))
	if info, err := os.Stat(build); err != nil || !info.IsDir() {
		return protocol.PublishResult{}, fmt.Errorf("%s: no build in %s to make the release asset from; build it first", p.Name, protocol.BuildOutputDir(p))
	}
	out := filepath.Join(workspaceRoot, filepath.FromSlash(asset))
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return protocol.PublishResult{}, err
	}
	if err := zipDir(build, out); err != nil {
		return protocol.PublishResult{}, fmt.Errorf("%s: %w", asset, err)
	}
	return res, nil
}

// assetName fills the asset name pattern; os and arch are Go's names, which
// the spec's aliases rename in the file name only.
func assetName(spec protocol.AssetSpec, project, version, goos, goarch, ext string) (string, error) {
	product := spec.Product
	if product == "" {
		product = project
	}
	if a, ok := spec.OSAlias[goos]; ok {
		goos = a
	}
	if a, ok := spec.ArchAlias[goarch]; ok {
		goarch = a
	}
	values := map[string]string{"product": product, "version": version, "os": goos, "arch": goarch, "ext": ext}
	pattern := spec.Name
	if pattern == "" {
		pattern = "{product}_{version}_{os}_{arch}.{ext}"
	}
	var unknown []string
	name := placeholder.ReplaceAllStringFunc(pattern, func(m string) string {
		v, ok := values[m[1:len(m)-1]]
		if !ok {
			unknown = append(unknown, m)
		}
		return v
	})
	if len(unknown) > 0 {
		return "", fmt.Errorf("asset name %q has unknown placeholder %s; use {product}, {version}, {os}, {arch} or {ext}", pattern, strings.Join(unknown, ", "))
	}
	return name, nil
}

// zipDir writes dir's files into a zip at out, with forward-slash paths
// relative to dir and their modes kept, so executables stay executable.
func zipDir(dir, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	w := zip.NewWriter(f)
	walkErr := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name, header.Method = filepath.ToSlash(rel), zip.Deflate
		dst, err := w.CreateHeader(header)
		if err != nil {
			return err
		}
		src, err := os.Open(p)
		if err != nil {
			return err
		}
		_, err = io.Copy(dst, src)
		_ = src.Close()
		return err
	})
	closeErr := w.Close()
	fileErr := f.Close()
	for _, e := range []error{walkErr, closeErr, fileErr} {
		if e != nil {
			return e
		}
	}
	return nil
}
