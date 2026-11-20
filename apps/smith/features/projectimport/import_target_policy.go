package projectimport

import (
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/features/adoptsplit"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var gitURL = regexp.MustCompile(`^(https?|ssh|git|file)://|^[\w.-]+@[\w.-]+:`)

// IsGitURL says whether the source is a git URL (https://, ssh://, git://,
// file:// or the scp form git@host:owner/repo) and not a local folder.
//
// Parameters: arg is the source as typed.
// Returns: true for a git URL.
func IsGitURL(arg string) bool { return gitURL.MatchString(strings.TrimSpace(arg)) }

// DefaultKind is the kind the identification suggests: an app for a program,
// a package for a project a registry publishes, a lib for the rest.
//
// Parameters: c is the identified project.
// Returns: the suggested kind.
func DefaultKind(c adoptmove.Candidate) protocol.ProjectKind {
	switch {
	case c.App:
		return protocol.ProjectApp
	case c.Published:
		return protocol.ProjectPackage
	}
	return protocol.ProjectLib
}

// ParseKind reads --as.
//
// Parameters: as is the flag's value.
// Returns: the kind, or a problem for a word that is not app, lib or package.
func ParseKind(as string) (protocol.ProjectKind, error) {
	switch k := protocol.ProjectKind(strings.TrimSpace(as)); k {
	case protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage:
		return k, nil
	}
	return "", problem.New("--as "+as+" is not a kind smith knows", "use --as app, --as lib or --as package")
}

// Folder settles the project's folder. --into wins and is checked; without
// it the folder is the kind's folder and the name (apps/<name>). A kind
// given with --as that the folder contradicts (--as app --into packages/x)
// is refused, because the folder decides the kind.
//
// Parameters: kind is the settled kind ("" when only --into says where);
// name is the project's name; into is --into or "".
// Returns: the folder with forward slashes and the kind it gives, or a problem.
func Folder(kind protocol.ProjectKind, name, into string) (string, protocol.ProjectKind, error) {
	if into == "" {
		dir, err := adoptmove.Destination(projectkind.Directory(kind), name)
		return dir, kind, err
	}
	dir, err := adoptmove.CleanFolder(into)
	if err != nil {
		return "", "", err
	}
	byFolder := projectkind.Resolve("", dir, "")
	switch {
	case kind == "" && byFolder == "":
		return "", "", problem.New("--into "+dir+" is outside apps/, libs/ and packages/, so smith cannot tell the project's kind", "add --as app, --as lib or --as package")
	case kind == "":
		return dir, byFolder, nil
	case byFolder != "" && byFolder != kind:
		return "", "", problem.New("--as "+string(kind)+" and --into "+dir+" disagree: "+projectkind.Directory(byFolder)+"/ holds the "+string(byFolder)+" projects",
			"use --into "+projectkind.Directory(kind)+"/"+path.Base(dir)+", or drop --as")
	}
	return dir, kind, nil
}

// History settles --history for a source.
//
// Parameters: flag is --history ("" when not given); git says whether the
// source is a git repository.
// Returns: the mode (keep for a git repository, copy for a plain folder,
// when the flag is empty), or a problem for a mode the source cannot give.
func History(flag string, git bool) (string, error) {
	switch flag {
	case "":
		if git {
			return HistoryKeep, nil
		}
		return HistoryCopy, nil
	case HistoryCopy:
		return flag, nil
	case HistoryKeep, HistorySquash:
		if !git {
			return "", problem.New("--history "+flag+" needs a git repository, and the source is a plain folder", "drop --history: a plain folder is copied", "or make the folder a git repository with a commit, then run smith import again")
		}
		return flag, nil
	}
	return "", problem.New("--history "+flag+" is not a mode smith knows", "use --history keep (every commit), squash (one commit) or copy (the files only)")
}

// Lockfiles are the lockfile names a project of the language drops when it
// joins the workspace, because the workspace has one lockfile for every
// project: npm's, yarn's, pnpm's and bun's for Node, uv's for Python.
//
// Parameters: ecosystem is the identified language.
// Returns: the names, as path.Match patterns in lower case.
func Lockfiles(ecosystem string) []string {
	if ecosystem == adoptsplit.Python {
		return []string{"uv.lock"}
	}
	return adoptsplit.For(ecosystem).Lockfiles
}

// ManifestNames lists every manifest the language plugins know, for the
// message that says no project was found.
//
// Parameters: claims are the plugins' claims.
// Returns: the names, joined with commas.
func ManifestNames(claims []adoptmove.Claim) string {
	var names []string
	seen := map[string]bool{}
	for _, c := range claims {
		for _, m := range c.Manifests {
			if !seen[m] {
				seen[m] = true
				names = append(names, m)
			}
		}
	}
	return strings.Join(names, ", ")
}
