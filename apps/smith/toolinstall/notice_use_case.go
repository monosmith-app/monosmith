package toolinstall

import (
	"context"
	"fmt"
	"io"
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
		what := "is not installed"
		if st.State == Outdated {
			what = "is too old (" + st.Version + " is installed, " + st.Wanted.String() + " is wanted)"
		}
		fmt.Fprintf(&b, "! %s %s: %s.\n  Run smith tools install %s to choose how to get it.\n", st.Title, what, st.Consequence, st.Name)
	}
	return b.String()
}

// Notice prints a warning for each tool the workspace needs that is missing,
// for `smith add` and `smith install`. extraLanguages and extraChoices add the
// tools of a language or choice that no project uses yet.
func Notice(ctx context.Context, s *session.Session, out io.Writer, extraLanguages, extraChoices []string) {
	needs := WorkspaceNeeds(s, extraLanguages, extraChoices)
	fmt.Fprint(out, NoticeText(DetectAll(ctx, ExecDetector{}, needs)))
}

// NoticeForLanguage prints the warnings for the tools of a language by name,
// without a session: `smith new` runs before the workspace can be opened.
// Git is always checked. A language with no entry gives only the git check.
func NoticeForLanguage(ctx context.Context, out io.Writer, language string) {
	names := append([]string{"git"}, languageTools[strings.ToLower(language)]...)
	var needs []Need
	for _, t := range toolcatalog.For(names...) {
		needs = append(needs, Need{ToolNeed: t})
	}
	fmt.Fprint(out, NoticeText(DetectAll(ctx, ExecDetector{}, needs)))
}
