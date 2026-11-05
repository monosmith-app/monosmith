package pipeline

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/buildrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var (
	binaryDirEntry = regexp.MustCompile(`("binaryDir"\s*:\s*)"((?:[^"\\]|\\.)*)"`)
	cmakeVariable  = regexp.MustCompile(`\$(?:env|penv|vendor)?\{[^}]*\}`)
)

// CheckBuildDir reports where each CMake project builds, from what smith can
// read of it, and offers to move a project whose presets build outside the
// workspace's dist/<project dir>.
//
// Two things are read. The configure task's -B (a task smith made always
// builds into dist/<project dir>; a task written by hand in smith.toml may
// not), and the binaryDir of each configure preset in CMakePresets.json,
// which `cmake --preset` and an IDE use. A binaryDir that names a variable
// other than ${sourceDir}, ${sourceParentDir}, ${presetName},
// ${hostSystemName} or ${generator} is left unchecked. A CMakeLists.txt
// that sets CMAKE_BINARY_DIR itself, a preset file another file includes,
// and CMakeUserPresets.json are not read.
//
// The fix rewrites the binaryDir values of CMakePresets.json (one preset:
// dist/<project dir>, several: dist/<project dir>/${presetName}) and keeps
// every other byte. A wrong -B in a task of smith.toml is reported with the
// manual step: smith does not rewrite a command line.
//
// Parameters: workspaceRoot is the absolute workspace folder; projects the
// CMake projects. Returns one finding for each project that has something
// to say, and false when none has.
func CheckBuildDir(workspaceRoot string, projects []protocol.Project) ([]protocol.Finding, bool) {
	var findings []protocol.Finding
	for _, p := range projects {
		if p.Name == "" || p.Root == "" {
			continue
		}
		if f, ok := projectBuildDir(workspaceRoot, p, projects); ok {
			findings = append(findings, f)
		}
	}
	return findings, len(findings) > 0
}

func projectBuildDir(workspaceRoot string, p protocol.Project, projects []protocol.Project) (protocol.Finding, bool) {
	target := protocol.ProjectBuildDir(p.Root)
	check := p.Name + " builds into " + target
	var where []string // what builds outside dist/
	taskTree, taskKnown := taskBuildTree(p)
	taskWrong := taskKnown && !protocol.InBuildDir(taskTree)
	if taskWrong {
		where = append(where, "its configure task builds into "+taskTree+"/")
	}
	presetsPath := path.Join(p.Root, "CMakePresets.json")
	presets := ""
	if raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(presetsPath))); err == nil {
		presets = string(raw)
	}
	var dirs []string // the binaryDir values, as written
	var outside []string
	unreadable := false
	for _, m := range binaryDirEntry.FindAllStringSubmatch(presets, -1) {
		dirs = append(dirs, m[2])
		resolved, ok := resolveBinaryDir(p.Root, m[2])
		switch {
		case !ok:
			unreadable = true
		case !protocol.InBuildDir(resolved):
			outside = append(outside, resolved)
		}
	}
	if len(outside) > 0 {
		where = append(where, presetsPath+" builds into "+strings.Join(outside, ", ")+"/")
	}
	if len(where) == 0 {
		if taskKnown || (len(dirs) > 0 && !unreadable) {
			return protocol.Finding{Check: check, OK: true}, true
		}
		return protocol.Finding{}, false
	}
	f := protocol.Finding{
		Check:  check,
		Detail: strings.Join(where, "; ") + ", not into " + target + "/",
		Remedy: "set binaryDir to \"${sourceDir}/" + upTo(p.Root) + target + "\" in " + presetsPath + ", or run smith doctor --fix",
	}
	if taskWrong {
		f.Remedy = "pass -B " + target + " (from the workspace root) to the configure, build and test tasks in smith.toml, and set binaryDir in " + presetsPath + " to the same folder"
		return f, true // a command line written by hand is not rewritten
	}
	old := make([]string, 0, len(outside))
	for _, o := range outside {
		old = append(old, strings.TrimPrefix(o, p.Root+"/"))
	}
	manual := f.Remedy
	if mentioned := buildrefs.Mentions(workspaceRoot, p.Root, old...); len(mentioned) > 0 {
		return refused(f, presetsPath, strings.Join(mentioned, ", ")+" name"+plural(len(mentioned))+" "+strings.Join(outside, ", ")+"/ and would stop finding the build", manual), true
	}
	if reason := referencedBy(workspaceRoot, p, outside, projects); reason != "" {
		return refused(f, presetsPath, reason, manual), true
	}
	value := "${sourceDir}/" + upTo(p.Root) + target
	if len(dirs) > 1 {
		value += "/${presetName}"
	}
	edited := binaryDirEntry.ReplaceAllStringFunc(presets, func(entry string) string {
		m := binaryDirEntry.FindStringSubmatch(entry)
		if resolved, ok := resolveBinaryDir(p.Root, m[2]); ok && !protocol.InBuildDir(resolved) {
			return m[1] + `"` + value + `"`
		}
		return entry
	})
	if edited == presets {
		return refused(f, presetsPath, "the edit would change nothing", manual), true
	}
	move := protocol.Fix{Description: "Build " + p.Name + " into " + target, Files: []protocol.FileEdit{{Path: presetsPath, Content: edited}}}
	f.Fix = &protocol.Fix{
		Description: move.Description,
		AppliesTo:   []string{p.Name},
		Choices: []protocol.FixChoice{{
			Question: p.Name + " builds into " + strings.Join(outside, ", ") + "/ through its presets. smith builds every project into " + target + ". Moving it sets binaryDir in " + presetsPath + " and keeps every other line. The old folder stays on disk, unused (git ignores it), and nothing is committed. The next screen shows every changed line.",
			Options: []protocol.FixChoiceOption{
				{Label: "build into " + target + " (edits " + presetsPath + ")", Fix: move},
				{Label: "leave it as it is (nothing changes)"},
			},
		}},
	}
	return f, true
}

func refused(f protocol.Finding, what, reason, manual string) protocol.Finding {
	f.Warning = true
	f.Remedy = "smith does not edit " + what + " for you, because " + reason + ". To move it by hand: " + strings.TrimSuffix(manual, ", or run smith doctor --fix")
	return f
}

func plural(n int) string {
	if n == 1 {
		return "s"
	}
	return ""
}

func upTo(root string) string { return strings.Repeat("../", strings.Count(root, "/")+1) }

// taskBuildTree is the build tree the configure task passes with -B, from the
// workspace root; known is false when the task has none or builds it from a
// variable.
func taskBuildTree(p protocol.Project) (tree string, known bool) {
	t, ok := p.Tasks["configure"]
	if !ok {
		return "", false
	}
	for i, arg := range t.Command {
		value := ""
		switch {
		case arg == "-B" && i+1 < len(t.Command):
			value = t.Command[i+1]
		case strings.HasPrefix(arg, "-B") && len(arg) > 2:
			value = arg[2:]
		default:
			continue
		}
		if strings.Contains(value, "${") {
			return "", false
		}
		if rest, ok := strings.CutPrefix(value, "{workspaceRoot}"); ok {
			return path.Clean(strings.TrimPrefix(rest, "/")), true
		}
		if filepath.IsAbs(value) || strings.Contains(value, "{") {
			return "", false
		}
		base := t.Cwd
		if base == "" {
			base = p.Root
		}
		return path.Clean(path.Join(base, value)), true
	}
	return "", false
}

// resolveBinaryDir is a preset's binaryDir from the workspace root; ok is
// false when it holds a variable smith does not evaluate.
func resolveBinaryDir(root, value string) (string, bool) {
	fromWorkspace := strings.Contains(value, "${sourceDir}") || strings.Contains(value, "${sourceParentDir}")
	value = strings.NewReplacer(
		"${sourceDir}", root, "${sourceParentDir}", path.Dir(root),
		"${presetName}", "preset", "${hostSystemName}", "host", "${generator}", "generator",
	).Replace(value)
	if cmakeVariable.MatchString(value) || filepath.IsAbs(value) || strings.HasPrefix(value, "/") {
		return "", false
	}
	if !fromWorkspace {
		value = path.Join(root, value) // relative: from the source folder
	}
	return path.Clean(value), true
}

// referencedBy names a project whose CMakeLists.txt reads this project's build
// folder by path.
func referencedBy(workspaceRoot string, p protocol.Project, outside []string, projects []protocol.Project) string {
	for _, q := range projects {
		if q.Name == p.Name {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(q.Root), "CMakeLists.txt"))
		if err != nil {
			continue
		}
		for _, o := range outside {
			if strings.Contains(string(raw), "../"+path.Base(p.Root)+"/"+strings.TrimPrefix(o, p.Root+"/")) {
				return path.Join(q.Root, "CMakeLists.txt") + " reads a file of " + p.Name + " from " + o + "/, which would stop existing"
			}
		}
	}
	return ""
}
