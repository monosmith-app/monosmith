package kindchange

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const pushGuardComment = "# Released by its tag, never pushed to a gem server."

var (
	kindMetadata   = regexp.MustCompile(`(\.metadata\[['"]smith_kind['"]\]\s*=\s*['"])([^'"]*)(['"])`)
	kindLine       = regexp.MustCompile(`(?m)^([ \t]*)\S*\.metadata\[['"]smith_kind['"]\][^\n]*\n`)
	pushGuard      = regexp.MustCompile(`(?m)^[ \t]*\S*\.metadata\[['"]allowed_push_host['"]\]\s*=\s*['"]none['"][^\n]*\r?\n`)
	pushGuardNote  = regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(pushGuardComment) + `[ \t]*\r?\n`)
	relativeString = regexp.MustCompile(`(['"])(\.{1,2}/[^'"]*)(['"])`)
)

func isGemFile(name string) bool { return name == "Gemfile" || strings.HasSuffix(name, ".gemspec") }

// Change answers what making the gem params.Project a params.To changes: the
// smith_kind metadata in its gemspec, the guard that keeps a lib or an app off
// the gem server, and the Gemfile entries and paths that name its directory.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	specs := edits.List(p.Root, func(name string) bool { return strings.HasSuffix(name, ".gemspec") })
	if len(specs) == 0 {
		return protocol.ChangeKindResult{}, fmt.Errorf("no gemspec in %s", p.Root)
	}
	manifest := specs[0]
	src, _ := edits.Read(manifest)
	var result protocol.ChangeKindResult

	src = setKind(src, string(params.To))
	src = setPushGuard(src, params.To != protocol.ProjectPackage)
	if params.To == protocol.ProjectPackage {
		for _, field := range []string{"description", "homepage"} {
			if !regexp.MustCompile(`\.` + field + `\s*=`).MatchString(src) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s has no %s: the page rubygems.org shows for a published gem uses it", manifest, field))
			}
		}
	}
	edits.Write(manifest, src)
	result.Warnings = append(result.Warnings, programWarnings(src, manifest, params.To)...)

	if params.NewRoot != p.Root {
		moveOwnFiles(edits, params)
		retargetDependents(edits, params)
		retargetWorkspace(edits, params)
	}
	result.Files = edits.Edits()
	return result, nil
}

// setKind sets the smith_kind metadata, adding it before the gemspec's last
// line when it has none.
func setKind(src, kind string) string {
	if kindMetadata.MatchString(src) {
		return kindMetadata.ReplaceAllString(src, "${1}"+kind+"${3}")
	}
	i := strings.LastIndex(src, "end")
	if i < 0 {
		return src
	}
	return src[:i] + "  spec.metadata['smith_kind'] = '" + kind + "'\n" + src[i:]
}

// setPushGuard keeps a gem that is not released to a gem server from being
// pushed to one, and lifts the guard from one that is.
func setPushGuard(src string, guarded bool) string {
	if !guarded {
		src = pushGuardNote.ReplaceAllString(src, "")
		return pushGuard.ReplaceAllString(src, "")
	}
	if pushGuard.MatchString(src) {
		return src
	}
	loc := kindLine.FindStringSubmatchIndex(src)
	if loc == nil {
		return src
	}
	indent := src[loc[2]:loc[3]]
	return src[:loc[0]] + indent + pushGuardComment + "\n" + indent + "spec.metadata['allowed_push_host'] = 'none'\n" + src[loc[0]:]
}

// programWarnings say what the kind change makes of the gem's executables.
func programWarnings(gemspec, manifest string, to protocol.ProjectKind) []string {
	hasProgram := regexp.MustCompile(`\.executables\s*=`).MatchString(gemspec)
	switch {
	case to == protocol.ProjectApp && !hasProgram:
		return []string{fmt.Sprintf("an app runs a program, and %s lists no executables: add them (and a bindir) before running it", manifest)}
	case to != protocol.ProjectApp && hasProgram:
		return []string{fmt.Sprintf("this gem has executables; as a %s it is installed by other gems and the programs are kept but not deployed", to)}
	}
	return nil
}

// moveOwnFiles keeps the gem's own relative paths (a Gemfile that reaches the
// workspace's, a require_relative) pointing at the right places once its
// directory moves.
func moveOwnFiles(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, file := range edits.List(p.Root, isGemFile) {
		text, _ := edits.Read(file)
		edits.Write(file, relativeString.ReplaceAllStringFunc(text, func(m string) string {
			parts := relativeString.FindStringSubmatch(m)
			return parts[1] + pathrefs.Rebase(parts[2], p.Root, params.NewRoot) + parts[3]
		}))
	}
}

// retargetDependents points every other gem's Gemfile and gemspec at the new
// directory.
func retargetDependents(edits *fileedits.Set, params protocol.ChangeKindParams) {
	for _, other := range params.Projects {
		if other.Language != "ruby" || other.Root == params.Project.Root {
			continue
		}
		for _, file := range edits.List(other.Root, isGemFile) {
			text, _ := edits.Read(file)
			updated, _ := pathrefs.Retarget(text, other.Root, params.Project.Root, params.NewRoot)
			edits.Write(file, updated)
		}
	}
}

// retargetWorkspace updates the root Gemfile (gemspec path: entries) and the
// lockfile, which records where each path gem lives.
func retargetWorkspace(edits *fileedits.Set, params protocol.ChangeKindParams) {
	for _, file := range []string{"Gemfile", "Gemfile.lock"} {
		if text, ok := edits.Read(file); ok {
			updated, _ := pathrefs.Retarget(text, "", params.Project.Root, params.NewRoot)
			edits.Write(file, updated)
		}
	}
}
