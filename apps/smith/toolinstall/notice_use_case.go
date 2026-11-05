package toolinstall

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
)

// languageTools names the catalogue tools of a language, for a command that
// runs before any plugin is loaded (smith new). Plugins stay the source of
// truth everywhere else; TestLanguageToolsAreInTheCatalogue keeps this table
// honest.
var languageTools = map[string][]string{
	"node": {"node"}, "go": {"go"}, "python": {"python"}, "py": {"python"}, "dotnet": {"dotnet"},
	"rust": {"rust"}, "jvm": {"java"}, "dart": {"dart"}, "ruby": {"ruby"}, "php": {"php"}, "cpp": {"cmake"},
}

// NoticeText words the tools that are missing or outdated as warnings, each
// with the command that offers the ways to get it. It is empty when every tool
// is there. It asks nothing and installs nothing.
func NoticeText(list []Status) string {
	var b strings.Builder
	for _, st := range list {
		if !st.Lacking() {
			continue
		}
		fmt.Fprintf(&b, "! %s %s: %s.\n  Run smith tools install %s to choose how to get it.\n", st.Title, st.Problem(), st.Consequence, st.Name)
	}
	return b.String()
}

// Notice prints a warning for each tool that is missing, for `smith add` and
// `smith install`. A command that touches some projects warns only about the
// tools those projects need, so a Python change never warns about the .NET SDK.
// extraLanguages and extraChoices add the tools of a language or choice that no
// project uses yet.
//
// Parameters: projects are the projects the command touches; nil means all of
// them (a plain install restores every project).
func Notice(ctx context.Context, s *session.Session, out io.Writer, projects, extraLanguages, extraChoices []string) {
	needs := OnlyProjects(WorkspaceNeeds(s, extraLanguages, extraChoices), projects)
	fmt.Fprint(out, NoticeText(DetectAll(ctx, ExecDetector{}, needs)))
}

// OnlyProjects keeps the tools that the named projects use, and the tools no
// project owns (git, the GitHub CLI, a tool named on the command line).
//
// Parameters: needs are the tools of the workspace; projects are the projects a
// command touches; nil keeps every tool.
// Returns: the tools the command can need.
func OnlyProjects(needs []Need, projects []string) []Need {
	if projects == nil {
		return needs
	}
	var out []Need
	for _, n := range needs {
		if len(n.Projects) == 0 || slices.ContainsFunc(n.Projects, func(p string) bool { return slices.Contains(projects, p) }) {
			out = append(out, n)
		}
	}
	return out
}

// OfferForLanguage deals with the tools of a language by name, without a
// session: `smith new` runs before the workspace can be opened. It offers as
// Offer does. Git is always checked. A language with no entry gives only the
// git check.
func OfferForLanguage(ctx context.Context, o OfferOptions, language string) {
	names := append([]string{"git"}, languageTools[strings.ToLower(language)]...)
	var needs []Need
	for _, t := range toolcatalog.For(names...) {
		needs = append(needs, Need{ToolNeed: t})
	}
	offerFor(ctx, DetectAll(ctx, ExecDetector{}, needs), o, ExecRunner{}, ExecDetector{})
}
