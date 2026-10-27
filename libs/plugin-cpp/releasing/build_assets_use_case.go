package releasing

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-cpp/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/releaseassets"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// hostPlatform is the machine's "os/arch" in Go's spelling; tests replace it.
var hostPlatform = runtime.GOOS + "/" + runtime.GOARCH

// CMakeRunner runs cmake; tests replace it.
var CMakeRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

// platforms are the systems and processors a release names.
var (
	systems    = []string{"windows", "linux", "darwin"}
	processors = []string{"amd64", "arm64"}
)

// CheckPlatform says why a platform ("os/arch") cannot be released, or "".
func CheckPlatform(platform string) string {
	goos, goarch, ok := strings.Cut(platform, "/")
	if !ok || !slices.Contains(systems, goos) || !slices.Contains(processors, goarch) {
		return fmt.Sprintf("%q is not one of %s with %s", platform, strings.Join(systems, ", "), strings.Join(processors, " or "))
	}
	return ""
}

// BuildAssets builds a C++ app in Release mode with CMake and Ninja and zips
// what it installs, or its program when it has no install rules, with the
// README and licence files, as the asset of this machine's platform. C++ is
// compiled and linked for the system it runs on, so an app is built on each
// operating system: a native app (set(SMITH_NATIVE ON) in its
// CMakeLists.txt), whose native jobs each ask for their own platform. The
// checks run in a dry run too. The version is the one the release wrote into
// the project() call.
func BuildAssets(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	spec := *params.Assets
	p := params.Project
	native := slices.Contains(p.Tags, discovery.NativeTag)
	if len(spec.Platforms) == 0 {
		return protocol.PublishResult{}, fmt.Errorf("%s: no platforms to build assets for", p.Name)
	}
	for _, platform := range spec.Platforms {
		if why := CheckPlatform(platform); why != "" {
			return protocol.PublishResult{}, fmt.Errorf("%s: %s", p.Name, why)
		}
		if platform != hostPlatform {
			return protocol.PublishResult{}, fmt.Errorf("%s: a C++ program is built on the system it is for, so this machine (%s) cannot build %s. Mark the app native (set(SMITH_NATIVE ON) in %s/CMakeLists.txt) and each runner of `smith ci native` builds its own archive", p.Name, hostPlatform, platform, p.Root)
		}
	}
	goos, goarch, _ := strings.Cut(hostPlatform, "/")
	name, err := releaseassets.AssetName(spec, p.Name, params.Version, goos, goarch, "zip")
	if err != nil {
		return protocol.PublishResult{}, err
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	lists := readLists(dir)
	binary := lists.Executable
	if binary == "" {
		binary = p.Name
	}
	asset := path.Join(spec.OutDir, name)
	res := protocol.PublishResult{Published: !params.DryRun, Assets: []string{asset}, AssetPlatforms: map[string]string{asset: hostPlatform}, Binary: binary}
	if native {
		// Each runner attaches its own archive; one checksums file cannot
		// cover builds made on different runners, so the archive has its own.
		res.Assets = append(res.Assets, asset+".sha256")
	} else {
		res.Assets = append(res.Assets, path.Join(spec.OutDir, releaseassets.ChecksumsFile))
	}
	if params.DryRun {
		return res, nil
	}

	outDir := filepath.Join(workspaceRoot, filepath.FromSlash(spec.OutDir))
	if err := os.RemoveAll(outDir); err != nil {
		return protocol.PublishResult{}, err
	}
	tree, stage := filepath.Join(outDir, "build"), filepath.Join(outDir, "stage")
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return protocol.PublishResult{}, err
	}
	toolrun.Step("Building " + p.Name + " in Release mode")
	for _, argv := range [][]string{
		{"cmake", "-S", ".", "-B", tree, "-G", "Ninja", "-DCMAKE_BUILD_TYPE=Release"},
		{"cmake", "--build", tree},
		{"cmake", "--install", tree, "--prefix", stage},
	} {
		if err := CMakeRunner(dir, argv); err != nil {
			return protocol.PublishResult{}, fmt.Errorf("%s: %w", p.Name, err)
		}
	}
	entries, err := programEntries(stage, tree, binary)
	if err != nil {
		return protocol.PublishResult{}, fmt.Errorf("%s: %w", p.Name, err)
	}
	archive := filepath.Join(outDir, name)
	if err := releaseassets.WriteZip(archive, addCompanions(entries, releaseassets.Companions(workspaceRoot, p.Root))); err != nil {
		return protocol.PublishResult{}, fmt.Errorf("%s: %w", p.Name, err)
	}
	for _, d := range []string{tree, stage} {
		if err := os.RemoveAll(d); err != nil {
			return protocol.PublishResult{}, err
		}
	}
	checksums := filepath.Join(outDir, releaseassets.ChecksumsFile)
	if native {
		checksums = archive + ".sha256"
	}
	if err := releaseassets.WriteChecksums(checksums, archive); err != nil {
		return protocol.PublishResult{}, err
	}
	return res, nil
}

// readLists reads the project's CMakeLists.txt.
func readLists(dir string) discovery.Lists {
	src, err := os.ReadFile(filepath.Join(dir, "CMakeLists.txt"))
	if err != nil {
		return discovery.Lists{}
	}
	return discovery.ParseLists(string(src))
}

// programEntries lists what the archive holds: the files the project's
// install rules put in stage, else the program from the build tree.
func programEntries(stage, tree, binary string) ([]releaseassets.Entry, error) {
	var entries []releaseassets.Entry
	err := filepath.WalkDir(stage, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(stage, p)
		if err != nil {
			return err
		}
		entries = append(entries, releaseassets.Entry{Path: p, Name: filepath.ToSlash(rel), Executable: isProgram(d.Name(), binary)})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(entries) > 0 {
		return entries, nil
	}
	for _, name := range []string{binary, binary + ".exe"} {
		program := filepath.Join(tree, name)
		if info, err := os.Stat(program); err == nil && !info.IsDir() {
			return []releaseassets.Entry{{Path: program, Name: name, Executable: true}}, nil
		}
	}
	return nil, fmt.Errorf("the build has no program named %s and the project installs nothing: add an install(TARGETS %s) rule, or name the app's add_executable target like the project", binary, binary)
}

func isProgram(file, binary string) bool { return file == binary || file == binary+".exe" }

// addCompanions adds the README and licence files the install does not
// already have.
func addCompanions(entries, companions []releaseassets.Entry) []releaseassets.Entry {
	for _, c := range companions {
		have := false
		for _, e := range entries {
			if strings.EqualFold(e.Name, c.Name) {
				have = true
			}
		}
		if !have {
			entries = append(entries, c)
		}
	}
	return entries
}
