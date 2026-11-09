package pipeline

import (
	"os"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tomltext"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// lintProgram is the program the lint task runs.
const lintProgram = "golangci-lint"

// lookAll lists every file a program name matches on PATH, in the order the
// system would pick them: the first is the one that runs. Tests replace it.
var lookAll = func(name string) []string { return pathMatches(name, os.Getenv("PATH"), goruntime.GOOS) }

var (
	versionAfterWord = regexp.MustCompile(`version v?(\d+\.\d+\.\d+)`)
	anyVersion       = regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)
)

// lintVersion reads the version a golangci-lint printed: "golangci-lint has
// version 2.15.0 built with go1.26" and "golangci-lint version v1.64.8 (commit
// ...)" both work. It returns "" when no version is in the text.
func lintVersion(output string) string {
	if m := versionAfterWord.FindStringSubmatch(output); m != nil {
		return m[1]
	}
	if m := anyVersion.FindStringSubmatch(output); m != nil {
		return m[1]
	}
	return ""
}

// lintBinary is one golangci-lint file and what it says about itself.
type lintBinary struct {
	path    string
	line    string
	version string
}

// probeLint runs a golangci-lint file for its version.
func probeLint(root, path string) lintBinary {
	out, _ := probe(root, []string{path, "--version"})
	return lintBinary{path: path, line: firstLine(out, ""), version: lintVersion(out)}
}

// words says the file and its version line: "C:\x\golangci-lint.exe (golangci-lint has version 1.64.8 ...)".
func (b lintBinary) words() string {
	if b.line == "" {
		return b.path + " (it printed no version)"
	}
	return b.path + " (" + b.line + ")"
}

// goBinFolder is where `go install` writes programs: GOBIN, else the bin folder
// of the first GOPATH entry. It returns "" when go does not say.
func goBinFolder(root string) string {
	out, err := probe(root, []string{"go", "env", "GOBIN", "GOPATH"})
	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(out), "\r", ""), "\n")
	if err != nil || len(lines) < 2 {
		return ""
	}
	if bin := strings.TrimSpace(lines[0]); bin != "" {
		return bin
	}
	if gopath := filepath.SplitList(strings.TrimSpace(lines[1])); len(gopath) > 0 && gopath[0] != "" {
		return filepath.Join(gopath[0], "bin")
	}
	return ""
}

// acceptedLintVersion is the version smith.toml says the workspace keeps on
// purpose: `golangci-lint = "2.15.0"` under [toolchains]. It is "" when none.
func acceptedLintVersion(root string) string {
	src, err := os.ReadFile(filepath.Join(root, "smith.toml"))
	if err != nil {
		return ""
	}
	raw, ok := tomltext.Raw(string(src), "toolchains", lintProgram)
	if !ok {
		return ""
	}
	return strings.Trim(strings.TrimSpace(strings.SplitN(raw, "#", 2)[0]), `"'`)
}

// lintCandidates are the files that can run as golangci-lint, in the order they
// win: the copy smith installed for the workspace, then what PATH finds.
func lintCandidates(root string) (found []lintBinary, onPath bool) {
	var paths []string
	if toolrun.ManagedBinExists(root) {
		managed := filepath.Join(toolrun.ManagedBinDir(root), lintProgram+exeSuffix(goruntime.GOOS))
		if _, err := os.Stat(managed); err == nil {
			paths = append(paths, managed)
		}
	}
	system := lookAll(lintProgram)
	onPath = len(system) > 0
	for _, p := range system {
		if len(paths) == 0 || !sameFile(paths[0], p) {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		paths = []string{lintProgram}
	}
	for _, p := range paths {
		found = append(found, probeLint(root, p))
	}
	return found, onPath
}

// lintFinding checks the golangci-lint that will run (the copy smith manages
// for the workspace, else the first on PATH) against the version CI pins. The
// detail always says which file it found and the version line it printed.
//
// When the version differs, the fix is a question with the ways on, none of
// them silent: install the pinned version into smith's own folder
// (.smith/bin), which the lint task and this check then use first, whatever
// PATH says; or keep the one found, which writes the accepted version under
// [toolchains] in smith.toml so the check stops warning. When the pinned
// version is already in the folder `go install` writes to but another copy
// comes first on PATH, or that folder is not on PATH, the remedy names both
// folders. smith never edits PATH and never deletes or replaces another copy.
func lintFinding(root, pin string) protocol.Finding {
	found, onPath := lintCandidates(root)
	first := found[0]
	accepted := acceptedLintVersion(root)
	f := protocol.Finding{Check: "golangci-lint is " + pin + ", as in CI", Area: "toolchain", Warning: true}
	switch {
	case !onPath && first.line == "":
		f.Detail = "golangci-lint is not on PATH"
	default:
		f.Detail = "found " + first.words()
	}
	if len(found) > 1 {
		var others []string
		for _, b := range found[1:] {
			others = append(others, b.words())
		}
		f.Detail += "; also on PATH, after it: " + strings.Join(others, ", ")
	}
	switch {
	case first.version == pin:
		f.OK = true
		return f
	case first.version != "" && accepted != "" && versionAccepted(accepted, first.version):
		f.OK, f.Detail = true, f.Detail+"; kept on purpose by [toolchains] in smith.toml (CI pins "+pin+", so lint results may differ from CI)"
		return f
	}
	switch {
	case !onPath && first.line == "":
		f.Problem = "golangci-lint is not on PATH, CI uses " + pin
	case first.version == "":
		f.Problem = "the golangci-lint on PATH reports no version, CI uses " + pin
	default:
		f.Problem = "golangci-lint is " + first.version + ", CI uses " + pin
	}
	switch order, ok := toolrun.PinOrder(first.version, pin); {
	case ok && order > 0:
		f.Detail += "; it is newer than the " + pin + " CI pins"
	case ok && order < 0:
		f.Detail += "; it is older than the " + pin + " CI pins"
	}

	managed := toolrun.ManagedBinDir(root)
	bin := goBinFolder(root)
	if bin != "" {
		here := filepath.Join(bin, lintProgram+exeSuffix(goruntime.GOOS))
		if _, err := os.Stat(here); err == nil && (!onPath || !sameFile(first.path, here)) {
			if installed := probeLint(root, here); installed.version == pin {
				f.Detail += "; the pinned " + pin + " is installed in " + bin
				if first.line != "" {
					f.Remedy = "PATH finds " + first.words() + " before " + bin + " (" + pin + "): move " + bin + " earlier in PATH, or remove the old copy. smith does not edit PATH or delete files. Or let smith use its own copy of " + pin + " (smith doctor --fix)"
				} else {
					f.Remedy = bin + " (" + pin + ") is where go install writes, and it is not on PATH: add it to PATH. smith does not edit PATH. Or let smith use its own copy of " + pin + " (smith doctor --fix)"
				}
			}
		}
	}
	if f.Remedy == "" {
		f.Remedy = "smith doctor --fix asks how: install " + pin + " into " + managed + " (smith then uses it first), or keep the one found"
	}
	f.Fix = lintFix(root, pin, managed, first)
	return f
}

// lintFix is the choice that clears a golangci-lint that differs from the pin.
// Its first answer is the recommended one.
func lintFix(root, pin, managed string, found lintBinary) *protocol.Fix {
	install := LintInstallCommand()
	managedInstall := protocol.Fix{
		Description: "Install golangci-lint " + pin + " into " + managed,
		Steps: []protocol.FixStep{{
			Description: "Install golangci-lint " + pin + " into " + managed + ", where smith's lint task and doctor look first (uses the network)",
			Argv:        install,
			Env:         []string{"GOBIN=" + managed},
		}},
	}
	options := []protocol.FixChoiceOption{{
		Label: "Install " + pin + " into smith's own folder " + managed + " and use it for lint (recommended). It leaves your other golangci-lint untouched and PATH unchanged",
		Fix:   managedInstall,
	}}
	if found.version != "" {
		if keep, ok := keepFix(root, found); ok {
			options = append(options, protocol.FixChoiceOption{
				Label: "Keep " + found.path + " (" + found.version + "): smith.toml records that you chose it, and doctor stops warning. Lint results may differ from CI, which uses " + pin,
				Fix:   keep,
			})
		}
		if update, ok := updateOption(root, pin, managed, found); ok {
			options = append(options, update)
		}
	} else {
		options = append(options, protocol.FixChoiceOption{
			Label: "Install " + pin + " with go install (" + strings.Join(install, " ") + "). It lands in the Go bin folder, which must come first on PATH",
			Fix: protocol.Fix{Description: "Install golangci-lint " + pin + " with go install", Steps: []protocol.FixStep{{
				Description: strings.Join(install, " ") + " (uses the network)", Argv: install,
			}}},
		})
	}
	return &protocol.Fix{
		Description: "Get the golangci-lint that CI pins",
		Choices: []protocol.FixChoice{{
			Question: "How do you want to deal with golangci-lint? CI pins " + pin + "; smith found " + found.words() + ".",
			Options:  options,
		}},
	}
}

// updateOption is the answer that updates the copy found, in place, to the
// pinned version. It is greyed with a reason when no route knows how the copy
// was installed. It is left out (ok = false) for smith's own copy, which the
// first answer already replaces, and when the tool has no route.
func updateOption(root, pin, managed string, found lintBinary) (protocol.FixChoiceOption, bool) {
	route, has := updateRoutes[lintProgram]
	if !has || sameFile(filepath.Dir(found.path), managed) {
		return protocol.FixChoiceOption{}, false
	}
	label := "Update " + found.path + " to " + pin + " (every project that uses this copy gets it)"
	argv, env, ok := route(root, found.path)
	if !ok {
		return protocol.FixChoiceOption{Label: label, Disabled: unknownOrigin}, true
	}
	command := strings.Join(env, " ") + " " + strings.Join(argv, " ")
	return protocol.FixChoiceOption{
		Label: label + ". Runs " + command + " (uses the network). It changes every project that uses this copy; PATH stays as it is",
		Fix: protocol.Fix{
			Description: "Update golangci-lint at " + found.path + " to " + pin,
			Steps: []protocol.FixStep{{
				Description: "Update " + found.path + " to " + pin + ": " + command + " (uses the network)",
				Argv:        argv,
				Env:         env,
			}},
		},
	}, true
}

// keepFix is the fix that records, in smith.toml, that the workspace keeps the
// golangci-lint version found: `golangci-lint = "<version>"` under [toolchains].
func keepFix(root string, found lintBinary) (protocol.Fix, bool) {
	src, err := os.ReadFile(filepath.Join(root, "smith.toml"))
	if err != nil {
		return protocol.Fix{}, false
	}
	return protocol.Fix{
		Description: "Keep golangci-lint " + found.version + " (recorded in smith.toml)",
		Files:       []protocol.FileEdit{{Path: "smith.toml", Content: tomltext.SetString(string(src), "toolchains", lintProgram, found.version)}},
	}, true
}

// versionAccepted reports whether a version meets an accepted one: the accepted
// version is a prefix on dot boundaries ("2.15" accepts 2.15.0).
func versionAccepted(accepted, have string) bool {
	return have == accepted || strings.HasPrefix(have, accepted+".")
}

// exeSuffix is the extension of a program file on a system.
func exeSuffix(goos string) string {
	if goos == "windows" {
		return ".exe"
	}
	return ""
}

// sameFile reports whether two paths name the same file, comparing case-blind
// on Windows, where the file system does.
func sameFile(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if goruntime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
