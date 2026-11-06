package pipeline

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/buildrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/xmltext"
)

// CheckBuildDir reports where each JVM module builds, from what its build tool
// reads, and offers to move a module that builds outside the workspace's
// dist/ (dist/maven/<artifactId> for Maven, dist/gradle/<artifactId> for
// Gradle).
//
// Maven: the <build><directory> of the module's pom, else of the parent pom
// the module names by the default relativePath (../pom.xml), up to the
// workspace root, else target/. A directory written with a property smith
// cannot evaluate, or a parent with another relativePath, is left unchecked.
//
// Gradle: layout.buildDirectory or buildDir in the module's script, else
// build/. A workspace whose root script sets the build directory for all
// modules is left unchecked, and so is a setting that names no path.
//
// Parameters: workspaceRoot is the absolute workspace folder; projects the
// JVM modules. Returns the findings, and false when there is none.
func CheckBuildDir(workspaceRoot string, projects []protocol.Project) ([]protocol.Finding, bool) {
	var findings []protocol.Finding
	for _, p := range projects {
		if p.Name == "" {
			continue
		}
		var f protocol.Finding
		var ok bool
		if script := gradleScriptOf(workspaceRoot, p.Root); script != "" {
			f, ok = gradleFinding(workspaceRoot, p, script, projects)
		} else {
			f, ok = mavenFinding(workspaceRoot, p, projects)
		}
		if ok {
			findings = append(findings, f)
		}
	}
	return findings, len(findings) > 0
}

func readFile(workspaceRoot, rel string) (string, bool) {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
	return string(raw), err == nil
}

func gradleScriptOf(workspaceRoot, root string) string {
	for _, name := range []string{"build.gradle.kts", "build.gradle"} {
		rel := path.Join(root, name)
		if _, ok := readFile(workspaceRoot, rel); ok {
			return rel
		}
	}
	return ""
}

// refusedFinding turns a failed finding into the warning of a move smith will
// not make, with the reason and the manual step.
func refusedFinding(f protocol.Finding, what, reason, manual string) protocol.Finding {
	f.Warning = true
	f.Remedy = "smith does not edit " + what + " for you, because " + reason + ". To move it by hand: " + strings.TrimSuffix(manual, ", or run smith doctor --fix")
	return f
}

// moveFix is the question of a build move: do it, or leave the module as it is.
func moveFix(name, target, oldDir, file, content, what string) *protocol.Fix {
	move := protocol.Fix{Description: "Build " + name + " into " + target, Files: []protocol.FileEdit{{Path: file, Content: content}}}
	return &protocol.Fix{
		Description: move.Description,
		AppliesTo:   []string{name},
		Choices: []protocol.FixChoice{{
			Question: name + " builds into " + oldDir + "/. smith builds every project into " + target + ". Moving it " + what + " in " + file + " and keeps every other line. The old folder stays on disk, unused (git ignores it), and nothing is committed. The next screen shows every changed line.",
			Options: []protocol.FixChoiceOption{
				{Label: "build into " + target + " (edits " + file + ")", Fix: move},
				{Label: "leave it as it is (nothing changes)"},
			},
		}},
	}
}

// ---- Maven ----

var (
	propertyRef   = regexp.MustCompile(`\$\{[^}]*\}`)
	noLocalParent = regexp.MustCompile(`<relativePath\s*/>`)
)

// mavenBuildDirectory is where Maven builds a module, from the workspace
// root, as its pom chain says. ok is false when it cannot be read reliably.
func mavenBuildDirectory(workspaceRoot, root, artifactID string) (dir string, set bool, ok bool) {
	local := path.Join(root, "target")
	for file := path.Join(root, "pom.xml"); ; {
		src, found := readFile(workspaceRoot, file)
		if !found {
			return local, false, true
		}
		if v, has := xmltext.Path(src, "project", "build", "directory"); has && v != "" {
			value := strings.NewReplacer(
				"${project.basedir}", "\x00base", "${basedir}", "\x00base",
				"${maven.multiModuleProjectDirectory}", "\x00root", "${project.artifactId}", artifactID,
			).Replace(v)
			if propertyRef.MatchString(value) {
				return "", false, false
			}
			switch {
			case filepath.IsAbs(value) || strings.HasPrefix(value, "/"):
				return value, true, true
			case strings.Contains(value, "\x00"):
				value = strings.ReplaceAll(strings.ReplaceAll(value, "\x00base", root), "\x00root", ".")
				return path.Clean(value), true, true
			}
			return path.Clean(path.Join(root, value)), true, true // relative: from the module
		}
		if _, parent := xmltext.Path(src, "project", "parent", "artifactId"); !parent || noLocalParent.MatchString(src) {
			return local, false, true
		}
		rel, _ := xmltext.Path(src, "project", "parent", "relativePath")
		if rel == "" {
			rel = "../pom.xml"
		}
		if !strings.HasSuffix(rel, ".xml") {
			rel = path.Join(rel, "pom.xml")
		}
		file = path.Join(path.Dir(file), rel)
		if strings.HasPrefix(file, "..") {
			return local, false, true // the parent is outside the workspace
		}
	}
}

func mavenFinding(workspaceRoot string, p protocol.Project, projects []protocol.Project) (protocol.Finding, bool) {
	pom := path.Join(p.Root, "pom.xml")
	src, ok := readFile(workspaceRoot, pom)
	if !ok {
		return protocol.Finding{}, false
	}
	artifactID, _ := xmltext.Path(src, "project", "artifactId")
	if packaging, _ := xmltext.Path(src, "project", "packaging"); artifactID == "" || packaging == "pom" {
		return protocol.Finding{}, false
	}
	target := protocol.BuildDir + "/maven/" + artifactID
	check := p.Name + " builds into " + target
	dir, set, readable := mavenBuildDirectory(workspaceRoot, p.Root, artifactID)
	if !readable {
		return protocol.Finding{}, false
	}
	if protocol.InBuildDir(dir) {
		return protocol.Finding{Check: check, Project: p.Name, OK: true}, true
	}
	source := "no <build><directory> in " + pom + " or its parent pom"
	if set {
		source = "its pom sets <directory> to " + dir
	}
	f := protocol.Finding{
		Check:   check,
		Project: p.Name,
		Detail:  pom + " builds into " + dir + "/ (" + source + "), not into " + target + "/",
		Remedy:  "set <build><directory>${project.basedir}/" + upTo(p.Root) + target + "</directory></build> in " + pom + ", or run smith doctor --fix",
	}
	f.Problem = f.Detail
	if outputs := p.Tasks["build"].Outputs; len(outputs) > 0 {
		f.Detail += ". The build task caches " + target + " instead, which that build never writes, so a cache hit restores nothing"
	}
	manual := f.Remedy
	if filepath.IsAbs(dir) || strings.HasPrefix(dir, "/") {
		return refusedFinding(f, pom, "the build goes to a folder outside the workspace ("+dir+")", manual), true
	}
	if reason := mavenRefusal(workspaceRoot, p, src, dir, projects); reason != "" {
		return refusedFinding(f, pom, reason, manual), true
	}
	value := "${project.basedir}/" + upTo(p.Root) + protocol.BuildDir + "/maven/${project.artifactId}"
	edited, edit := xmltext.SetPath(src, value, "project", "build", "directory")
	if !edit {
		return refusedFinding(f, pom, "its XML is written in a form smith cannot edit safely", manual), true
	}
	if check, set, readable := mavenAfterEdit(workspaceRoot, p, edited, artifactID); !readable || !set || !protocol.InBuildDir(check) {
		return refusedFinding(f, pom, "the edit would not give the build directory smith expects", manual), true
	}
	f.Fix = moveFix(p.Name, target, dir, pom, edited, "sets <build><directory>")
	return f, true
}

// mavenAfterEdit reads the build directory a pom would give once edited, to
// prove the edit says what it should before it is offered.
func mavenAfterEdit(workspaceRoot string, p protocol.Project, edited, artifactID string) (string, bool, bool) {
	v, _ := xmltext.Path(edited, "project", "build", "directory")
	value := strings.NewReplacer("${project.basedir}", p.Root, "${project.artifactId}", artifactID).Replace(v)
	return path.Clean(value), v != "", !propertyRef.MatchString(value)
}

func upTo(root string) string { return strings.Repeat("../", strings.Count(root, "/")+1) }

func mavenRefusal(workspaceRoot string, p protocol.Project, pomSrc, dir string, projects []protocol.Project) string {
	old := path.Base(dir)
	if literal := regexp.MustCompile(`(?:^|[^\w./-])` + regexp.QuoteMeta(old) + `[/\\]`); literal.MatchString(pomSrc) {
		return "its pom names " + old + "/ as a path, which would stop existing"
	}
	if mentioned := buildrefs.Mentions(workspaceRoot, p.Root, strings.TrimPrefix(dir, p.Root+"/")); len(mentioned) > 0 {
		return strings.Join(mentioned, ", ") + " name" + plural(len(mentioned)) + " " + dir + "/ and would stop finding the build"
	}
	ref := regexp.MustCompile(`(?:^|[^\w./-])` + regexp.QuoteMeta(dir) + `[/\\]|\.\./` + regexp.QuoteMeta(path.Base(p.Root)) + `/` + regexp.QuoteMeta(old) + `[/\\]`)
	for _, q := range projects {
		if q.Name == p.Name {
			continue
		}
		if other, ok := readFile(workspaceRoot, path.Join(q.Root, "pom.xml")); ok && ref.MatchString(other) {
			return path.Join(q.Root, "pom.xml") + " reads a file of " + p.Name + " by its path in " + old + "/, which would stop existing"
		}
	}
	return ""
}

func plural(n int) string {
	if n == 1 {
		return "s"
	}
	return ""
}

// ---- Gradle ----

var (
	// buildDirSetting is a one-line assignment of the build directory.
	buildDirSetting = regexp.MustCompile(`(?m)^([ \t]*)(?:layout\.)?(?:buildDirectory|buildDir)\b[ \t]*(?:=|\.set\()[^\r\n]*$`)
	buildDirAny     = regexp.MustCompile(`\b(?:layout\.buildDirectory|buildDirectory|buildDir)\b\s*(?:=|\.set\(|\.fileValue)`)
	quoted          = regexp.MustCompile(`["']([^"']+)["']`)
	centralDir      = regexp.MustCompile(`["'](?:[^"']*/)?` + regexp.QuoteMeta(protocol.BuildDir) + `/gradle/[^"']+["']`)
	plugins         = regexp.MustCompile(`(?m)^plugins\s*\{`)
	buildLiteral    = regexp.MustCompile(`["'](?:\./)?build(?:/|["'])`)
)

func gradleFinding(workspaceRoot string, p protocol.Project, script string, projects []protocol.Project) (protocol.Finding, bool) {
	src, _ := readFile(workspaceRoot, script)
	id := p.Name
	target := protocol.BuildDir + "/gradle/" + id
	if m := regexp.MustCompile(`(?m)^\s*artifactId\s*=\s*["']([^"']+)["']`).FindStringSubmatch(src); m != nil {
		id = m[1]
		target = protocol.BuildDir + "/gradle/" + id
	}
	check := p.Name + " builds into " + target
	line := buildDirSetting.FindStringSubmatchIndex(src)
	where := path.Join(p.Root, "build")
	source := "no layout.buildDirectory in " + script
	if line != nil {
		text := src[line[0]:line[1]]
		if centralDir.MatchString(text) {
			return protocol.Finding{Check: check, Project: p.Name, OK: true}, true
		}
		lit := quoted.FindStringSubmatch(text)
		if lit == nil || strings.Count(src, "buildDir") > 1 {
			return protocol.Finding{}, false // names no path, or set more than once: not read
		}
		where, source = path.Clean(path.Join(p.Root, lit[1])), script+" sets the build directory to "+lit[1]
	} else {
		if buildDirAny.MatchString(src) {
			return protocol.Finding{}, false // set in a form smith does not read
		}
		for _, name := range []string{"build.gradle.kts", "build.gradle"} {
			if rootScript, ok := readFile(workspaceRoot, name); ok && buildDirAny.MatchString(rootScript) {
				return protocol.Finding{}, false // the root script sets it for the modules
			}
		}
	}
	f := protocol.Finding{
		Check:   check,
		Project: p.Name,
		Detail:  script + " builds into " + where + "/ (" + source + "), not into " + target + "/",
		Remedy:  "add layout.buildDirectory.set(rootProject.layout.projectDirectory.dir(\"" + target + "\")) to " + script + ", or run smith doctor --fix",
	}
	f.Problem = f.Detail
	manual := f.Remedy
	if reason := gradleRefusal(workspaceRoot, p, src, script, where, projects); reason != "" {
		return refusedFinding(f, script, reason, manual), true
	}
	setting := "layout.buildDirectory.set(rootProject.layout.projectDirectory.dir(\"" + target + "\"))"
	var edited string
	if line != nil {
		edited = src[:line[0]] + src[line[2]:line[3]] + setting + src[line[1]:]
	} else if edited = insertAfterPlugins(src, setting); edited == "" {
		eol := "\n"
		if strings.Contains(src, "\r\n") {
			eol = "\r\n"
		}
		edited = strings.TrimRight(src, "\r\n") + eol + eol + setting + eol
	}
	if !centralDir.MatchString(edited) {
		return refusedFinding(f, script, "the edit would not give the build directory smith expects", manual), true
	}
	f.Fix = moveFix(p.Name, target, where, script, edited, "sets the build directory")
	return f, true
}

// insertAfterPlugins puts a line after the plugins { } block, where a script
// sets its project-wide values; "" when the script has none or its braces do
// not balance.
func insertAfterPlugins(src, line string) string {
	loc := plugins.FindStringIndex(src)
	if loc == nil {
		return ""
	}
	depth := 0
	for i := loc[1] - 1; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end := i + 1
				if nl := strings.IndexByte(src[end:], '\n'); nl >= 0 {
					end += nl + 1
				} else {
					end = len(src)
				}
				eol := "\n"
				if strings.Contains(src, "\r\n") {
					eol = "\r\n"
				}
				prefix := ""
				if !strings.HasSuffix(src[:end], "\n") {
					prefix = eol
				}
				return src[:end] + prefix + eol + line + eol + src[end:]
			}
		}
	}
	return ""
}

func gradleRefusal(workspaceRoot string, p protocol.Project, src, script, where string, projects []protocol.Project) string {
	if buildLiteral.MatchString(src) {
		return script + " names the build/ folder as a path, which would stop existing"
	}
	if mentioned := buildrefs.Mentions(workspaceRoot, p.Root, strings.TrimPrefix(where, p.Root+"/")); len(mentioned) > 0 {
		return strings.Join(mentioned, ", ") + " name" + plural(len(mentioned)) + " " + where + "/ and would stop finding the build"
	}
	ref := regexp.MustCompile(`(?:^|[^\w./-])` + regexp.QuoteMeta(where) + `[/\\]|\.\./` + regexp.QuoteMeta(path.Base(p.Root)) + `/build[/\\]`)
	for _, q := range projects {
		if q.Name == p.Name {
			continue
		}
		if other := gradleScriptOf(workspaceRoot, q.Root); other != "" {
			if text, _ := readFile(workspaceRoot, other); ref.MatchString(text) {
				return other + " reads a file of " + p.Name + " by its path in build/, which would stop existing"
			}
		}
	}
	return ""
}
