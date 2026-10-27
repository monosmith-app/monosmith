package releasing

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/releaseassets"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// hostPlatform is the machine's "os/arch" in Go's spelling; tests replace it.
var hostPlatform = runtime.GOOS + "/" + runtime.GOARCH

// CargoBuild runs cargo build and returns what cargo printed on stdout (its
// JSON messages); tests replace it.
var CargoBuild = func(dir string, argv []string, env []string) (string, error) {
	return toolrun.Output(dir, argv, env...)
}

// InstalledTargets lists the Rust targets installed for the active toolchain.
// ok is false when rustup is not there to ask (a distribution's Rust has the
// targets it has), and then nothing is concluded. Tests replace it.
var InstalledTargets = func(dir string) (installed []string, ok bool) {
	out, err := toolrun.Output(dir, []string{"rustup", "target", "list", "--installed"})
	if err != nil {
		return nil, false
	}
	return strings.Fields(out), true
}

// BuildAssets builds an app for each platform of its release, one zip per
// platform holding the program with the README and licence beside it, and the
// checksums of them all. A Rust program is built with cargo's --target for the
// platform: the machine can build for its own system, for the other
// processor of it where the linker allows, and not for another operating
// system. An app that is native ([package.metadata.smith] native = true) is
// asked for this machine's platform only, by the native job of each runner.
// The checks run in a dry run too.
func BuildAssets(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	spec := *params.Assets
	p := params.Project
	native := slices.Contains(p.Tags, discovery.NativeTag)
	if len(spec.Platforms) == 0 {
		return protocol.PublishResult{}, fmt.Errorf("%s: no platforms to build assets for", p.Name)
	}
	crate := p.Name
	if len(p.Exports) > 0 {
		crate = p.Exports[0]
	}
	plans := make([]platformPlan, 0, len(spec.Platforms))
	for _, platform := range spec.Platforms {
		plan, err := planPlatform(workspaceRoot, p.Name, platform)
		if err != nil {
			return protocol.PublishResult{}, err
		}
		name, err := releaseassets.AssetName(spec, p.Name, params.Version, plan.os, plan.arch, "zip")
		if err != nil {
			return protocol.PublishResult{}, err
		}
		plan.archive = name
		plans = append(plans, plan)
	}
	if err := checkInstalled(workspaceRoot, p.Name, plans); err != nil {
		return protocol.PublishResult{}, err
	}

	result := protocol.PublishResult{Published: !params.DryRun, AssetPlatforms: map[string]string{}}
	outDir := filepath.Join(workspaceRoot, filepath.FromSlash(spec.OutDir))
	if !params.DryRun {
		if err := os.RemoveAll(outDir); err != nil {
			return protocol.PublishResult{}, err
		}
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return protocol.PublishResult{}, err
		}
	}
	var archives []string
	for _, plan := range plans {
		asset := path.Join(spec.OutDir, plan.archive)
		result.Assets = append(result.Assets, asset)
		result.AssetPlatforms[asset] = plan.platform
		if native {
			// Each runner attaches its own archive; one checksums file cannot
			// cover builds made on different runners, so each archive has its own.
			result.Assets = append(result.Assets, asset+".sha256")
		}
		if params.DryRun {
			continue
		}
		binary, err := buildArchive(workspaceRoot, p, crate, plan, filepath.Join(outDir, plan.archive))
		if err != nil {
			return protocol.PublishResult{}, err
		}
		if result.Binary == "" {
			result.Binary = binary
		}
		archive := filepath.Join(outDir, plan.archive)
		archives = append(archives, archive)
		if native {
			if err := releaseassets.WriteChecksums(archive+".sha256", archive); err != nil {
				return protocol.PublishResult{}, err
			}
		}
	}
	if result.Binary == "" {
		result.Binary = p.Name
	}
	if !native {
		result.Assets = append(result.Assets, path.Join(spec.OutDir, releaseassets.ChecksumsFile))
		if !params.DryRun {
			if err := releaseassets.WriteChecksums(filepath.Join(outDir, releaseassets.ChecksumsFile), archives...); err != nil {
				return protocol.PublishResult{}, err
			}
		}
	}
	return result, nil
}

// platformPlan is one platform of a release and what building it takes.
type platformPlan struct {
	platform, os, arch, triple, archive string
	reach                               Reach
}

// planPlatform checks that this machine can build the platform's program.
func planPlatform(workspaceRoot, app, platform string) (platformPlan, error) {
	goos, goarch, ok := strings.Cut(platform, "/")
	if !ok || goos == "" || goarch == "" {
		return platformPlan{}, fmt.Errorf("%s: platform %q is not os/arch", app, platform)
	}
	triple, err := Target(platform)
	if err != nil {
		return platformPlan{}, fmt.Errorf("%s: %w", app, err)
	}
	plan := platformPlan{platform: platform, os: goos, arch: goarch, triple: triple, reach: CrossReach(hostPlatform, platform)}
	switch plan.reach {
	case OnItsOwnSystem:
		return plan, fmt.Errorf("%s: this machine (%s) cannot build %s: a Rust program is built on the system it is for. Mark the app native (native = true under [package.metadata.smith] in its Cargo.toml) and each runner of `smith ci native` builds its own archive", app, hostPlatform, platform)
	case WithLinker:
		if !LinkerAvailable(workspaceRoot, platform, triple) {
			program, pkg := Linker(platform)
			return plan, fmt.Errorf("%s: building %s on this machine (%s) needs a linker for it: install %s (apt install %s), or set %s, or build it on a %s runner by marking the app native", app, platform, hostPlatform, program, pkg, LinkerEnv(triple), platform)
		}
	}
	return plan, nil
}

// checkInstalled fails when a platform's Rust target is not installed.
func checkInstalled(workspaceRoot, app string, plans []platformPlan) error {
	installed, ok := InstalledTargets(workspaceRoot)
	if !ok {
		return nil
	}
	var missing []string
	for _, plan := range plans {
		if !slices.Contains(installed, plan.triple) {
			missing = append(missing, plan.triple)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("%s: the Rust target %s is not installed: run rustup target add %s", app, strings.Join(missing, ", "), strings.Join(missing, " "))
}

// buildArchive builds the app for one platform and zips its programs with the
// README and licence. The version is the one the release wrote into the
// manifest, which cargo builds into the program (CARGO_PKG_VERSION). It
// returns the name of the app's program, without ".exe".
func buildArchive(workspaceRoot string, p protocol.Project, crate string, plan platformPlan, archive string) (string, error) {
	argv := []string{"cargo", "build", "--release", "-p", crate, "--target", plan.triple, "--message-format=json-render-diagnostics"}
	var env []string
	if plan.reach == WithLinker && os.Getenv(LinkerEnv(plan.triple)) == "" && !configured(workspaceRoot, plan.triple) {
		program, _ := Linker(plan.platform)
		env = append(env, LinkerEnv(plan.triple)+"="+program)
	}
	toolrun.Step("Building " + p.Name + " for " + plan.platform)
	out, err := CargoBuild(workspaceRoot, argv, env)
	if err != nil {
		return "", fmt.Errorf("building %s: %w", plan.platform, err)
	}
	programs := builtPrograms(out)
	if len(programs) == 0 {
		return "", fmt.Errorf("%s: cargo built no program for %s: the crate has no binary target (src/main.rs or [[bin]])", p.Name, plan.platform)
	}
	var entries []releaseassets.Entry
	program := ""
	for _, exe := range programs {
		name := filepath.Base(exe)
		entries = append(entries, releaseassets.Entry{Path: exe, Name: name, Executable: true})
		stem := strings.TrimSuffix(name, ".exe")
		if program == "" || stem == crate || stem == p.Name {
			program = stem
		}
	}
	entries = append(entries, releaseassets.Companions(workspaceRoot, p.Root)...)
	if err := releaseassets.WriteZip(archive, entries); err != nil {
		return "", fmt.Errorf("%s: %w", plan.platform, err)
	}
	return program, nil
}

// builtPrograms reads cargo's JSON messages and returns the programs it
// built, sorted: the paths of the artifacts that are executables.
func builtPrograms(messages string) []string {
	var programs []string
	scanner := bufio.NewScanner(strings.NewReader(messages))
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		var m struct {
			Reason     string `json:"reason"`
			Executable string `json:"executable"`
		}
		if json.Unmarshal(scanner.Bytes(), &m) != nil || m.Reason != "compiler-artifact" || m.Executable == "" {
			continue
		}
		if !slices.Contains(programs, m.Executable) {
			programs = append(programs, m.Executable)
		}
	}
	slices.Sort(programs)
	return programs
}
