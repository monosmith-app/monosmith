package pipeline

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/buildrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tomltext"
)

// cargoTargetDir is the folder smith builds every crate into: one target
// folder for the whole workspace, so crates share their compiled dependencies.
var cargoTargetDir = protocol.BuildDir + "/cargo"

// cargoConfigFiles are the files cargo reads its configuration from, in each
// folder, newest name first.
var cargoConfigFiles = []string{".cargo/config.toml", ".cargo/config"}

// buildHome is one place where cargo is configured: the workspace root, or
// the folder of a crate that is the root of its own workspace. The crates
// that share a home share one target folder.
type buildHome struct {
	dir    string
	crates []string
	// config is the file that sets target-dir ("" when none does), and value
	// is what it says as written in the file.
	config, value string
}

// CheckBuildDir reports where the Rust crates build, from what cargo reads:
// the target-dir of the nearest .cargo/config.toml (cargo reads the files from
// the crate's folder up to the root), else the default target/ folder. The
// CARGO_TARGET_DIR of the environment wins over both. A crate whose build
// goes outside the workspace's dist/ is a finding with a fix that asks first.
//
// Cargo.toml has no target-dir key (cargo reads it only from its
// configuration), so the manifest is not read for it. A configuration that
// states target-dir in a form smith cannot read as a plain key (an inline
// table, a dotted key) is left unchecked.
//
// Parameters: workspaceRoot is the absolute workspace folder; projects the
// Rust crates. Returns one finding for each home, and false when there is no crate.
func CheckBuildDir(workspaceRoot string, projects []protocol.Project) ([]protocol.Finding, bool) {
	homes := map[string]*buildHome{}
	var order []string
	for _, p := range projects {
		home := homeOf(workspaceRoot, p)
		h, ok := homes[home]
		if !ok {
			h = &buildHome{dir: home}
			h.config, h.value = readTargetDir(workspaceRoot, p.Root, home)
			homes[home] = h
			order = append(order, home)
		}
		h.crates = append(h.crates, p.Name)
	}
	sort.Strings(order)
	var findings []protocol.Finding
	for _, key := range order {
		if f, ok := homeFinding(workspaceRoot, *homes[key], projects); ok {
			findings = append(findings, f)
		}
	}
	return findings, len(findings) > 0
}

// homeOf is the folder whose target folder the crate uses: its own folder when
// its Cargo.toml has a [workspace] table, else the workspace root.
func homeOf(workspaceRoot string, p protocol.Project) string {
	if src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "Cargo.toml")); err == nil {
		if crate, err := discovery.ParseCargo(src); err == nil && crate.HasWorkspace {
			return p.Root
		}
	}
	return "."
}

// readTargetDir finds the nearest configuration file that sets target-dir,
// walking up from the crate's folder to the workspace root.
func readTargetDir(workspaceRoot, crateRoot, home string) (file, value string) {
	for dir := crateRoot; ; dir = path.Dir(dir) {
		for _, name := range cargoConfigFiles {
			rel := path.Join(dir, name)
			src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
			if err != nil {
				continue
			}
			if raw, ok := tomltext.Raw(string(src), "build", "target-dir"); ok {
				var v struct{ V string }
				if toml.Unmarshal([]byte("v = "+raw), &v) == nil && v.V != "" {
					return rel, v.V
				}
				return rel, "\x00" // written in a form that is not a plain string
			}
			if strings.Contains(string(src), "target-dir") {
				return rel, "\x00"
			}
		}
		if dir == "." || dir == home {
			return "", ""
		}
	}
}

func homeFinding(workspaceRoot string, h buildHome, projects []protocol.Project) (protocol.Finding, bool) {
	names := strings.Join(h.crates, ", ")
	check := "Rust builds into " + cargoTargetDir
	if h.dir != "." {
		check = names + " builds into " + cargoTargetDir
	}
	if h.value == "\x00" {
		return protocol.Finding{}, false // not read: said in the manual
	}
	where, source := "target", "the default, because no .cargo/config.toml sets target-dir"
	if h.config != "" {
		where, source = h.value, h.config+" sets target-dir to "+h.value
	}
	if env := os.Getenv("CARGO_TARGET_DIR"); env != "" {
		where, source = env, "CARGO_TARGET_DIR in the environment"
	}
	resolved := where
	if !filepath.IsAbs(where) {
		base := h.dir
		if h.config != "" && os.Getenv("CARGO_TARGET_DIR") == "" {
			base = path.Dir(path.Dir(h.config))
		}
		resolved = path.Clean(path.Join(base, filepath.ToSlash(where)))
	}
	if protocol.InBuildDir(resolved) {
		return filed(protocol.Finding{Check: check, OK: true}, h.crates), true
	}
	f := protocol.Finding{
		Check:  check,
		Detail: names + verb(len(h.crates)) + " into " + resolved + "/ (" + source + "), not into " + cargoTargetDir + "/",
		Remedy: "set [build] target-dir = \"" + relativeTarget(h.dir) + "\" in " + configFile(h) + ", or run smith doctor --fix",
	}
	f.Problem = f.Detail
	if len(h.crates) == 1 {
		f.Problem = strings.TrimPrefix(f.Detail, names+" ")
	}
	f = filed(f, h.crates)
	if os.Getenv("CARGO_TARGET_DIR") != "" {
		f.Warning = true
		f.Remedy = "unset CARGO_TARGET_DIR (the environment overrides .cargo/config.toml), or point it into " + cargoTargetDir + "; smith does not edit your environment"
		return f, true
	}
	if filepath.IsAbs(where) {
		f.Warning = true
		f.Remedy = "smith does not move a target folder outside the workspace (" + where + "); set [build] target-dir = \"" + relativeTarget(h.dir) + "\" in " + configFile(h) + " by hand"
		return f, true
	}
	if mentioned := buildrefs.Mentions(workspaceRoot, h.dir, resolvedFromHome(h.dir, resolved)); len(mentioned) > 0 {
		f.Warning = true
		f.Remedy = "smith does not move this build for you, because " + strings.Join(mentioned, ", ") + " name" + plural(len(mentioned)) + " " + resolved + "/ and would stop finding the build. To move it by hand: set [build] target-dir = \"" + relativeTarget(h.dir) + "\" in " + configFile(h) + ", and update " + strings.Join(mentioned, ", ")
		return f, true
	}
	fix, refusal := targetDirFix(workspaceRoot, h, names, resolved)
	f.Fix = fix
	if refusal != "" {
		f.Warning = true
		f.Remedy = "smith does not edit " + configFile(h) + " for you, because " + refusal + ". To move it by hand: set [build] target-dir = \"" + relativeTarget(h.dir) + "\" in it"
	}
	return f, true
}

func plural(n int) string {
	if n == 1 {
		return "s"
	}
	return ""
}

// resolvedFromHome is a workspace-relative folder as the home names it.
func resolvedFromHome(home, resolved string) string {
	if home == "." {
		return resolved
	}
	return strings.TrimPrefix(resolved, home+"/")
}

// configFile is the file the fix writes: the one that sets target-dir, else
// .cargo/config.toml of the home.
func configFile(h buildHome) string {
	if h.config != "" {
		return h.config
	}
	return path.Join(h.dir, ".cargo/config.toml")
}

// relativeTarget is dist/cargo as the .cargo folder of the home names it: from
// the folder that holds .cargo, so the workspace root writes dist/cargo.
func relativeTarget(home string) string {
	if home == "." {
		return cargoTargetDir
	}
	return strings.Repeat("../", strings.Count(home, "/")+1) + cargoTargetDir
}

// targetDirFix writes target-dir into the configuration, keeping its other
// lines. It asks first: move the build, or leave the crates as they are. The
// old target folder is not touched.
func targetDirFix(workspaceRoot string, h buildHome, names, oldDir string) (*protocol.Fix, string) {
	file := configFile(h)
	src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(file)))
	var edited string
	if err != nil {
		edited = "# Every build goes to the workspace's dist/ (smith).\n[build]\ntarget-dir = \"" + relativeTarget(h.dir) + "\"\n"
	} else {
		edited = tomltext.SetString(string(src), "build", "target-dir", relativeTarget(h.dir))
	}
	var check map[string]any
	if toml.Unmarshal([]byte(edited), &check) != nil {
		return nil, file + " is not valid TOML, so smith cannot edit it safely"
	}
	move := protocol.Fix{Description: "Build the Rust crates into " + cargoTargetDir, Files: []protocol.FileEdit{{Path: file, Content: edited}}}
	return &protocol.Fix{
		Description: "Build " + names + " into " + cargoTargetDir,
		AppliesTo:   h.crates,
		Choices: []protocol.FixChoice{{
			Question: names + verb(len(h.crates)) + " into " + oldDir + "/. smith builds every Rust crate into " + cargoTargetDir + ", one target folder for the workspace. Moving it sets target-dir in " + file + " and keeps every other line. The old folder stays on disk, unused (git ignores it), and nothing is committed. The next screen shows every changed line.",
			Options: []protocol.FixChoiceOption{
				{Label: "build into " + cargoTargetDir + " (writes " + file + ")", Fix: move},
				{Label: "leave it as it is (nothing changes)"},
			},
		}},
	}, ""
}

// verb is "builds" for one project and "build" for several, after a name list.
func verb(n int) string {
	if n == 1 {
		return " builds"
	}
	return " build"
}

// filed files a finding under its project when one crate builds in the home,
// and under "config" when several share it or it is the whole workspace's.
func filed(f protocol.Finding, crates []string) protocol.Finding {
	if len(crates) == 1 {
		f.Project = crates[0]
	} else {
		f.Area = "config"
	}
	return f
}
