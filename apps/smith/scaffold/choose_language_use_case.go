package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// chooseLanguage picks the language of a new project when --lang and --stack
// did not: a terminal is asked (the language the workspace already uses is the
// default), anywhere else the workspace's only language is used and the note
// says so, so a project called "site" is never silently a Go service.
func chooseLanguage(s *session.Session, opts AddOptions) (string, error) {
	used := map[string]bool{}
	for _, p := range s.Graph.Projects {
		used[p.Plugin] = true
	}
	names := make([]string, 0, len(used))
	for l := range used {
		names = append(names, l)
	}
	sort.Strings(names)
	var in io.Reader
	interactive := false
	if opts.Recipe != nil {
		in, interactive = opts.Recipe.In, opts.Recipe.Interactive
	}
	lang, note, err := pickLanguage(opts.Lang, opts.Name, names, s.Plugins.Languages(), in, opts.Out, interactive)
	if note != "" && opts.Out != nil {
		fmt.Fprintln(opts.Out, note)
	}
	return lang, err
}

// pickLanguage decides the language: the one asked for, else the answer of a
// person at a terminal, else the only language in use. note says what was
// chosen without being asked, and why.
func pickLanguage(requested, project string, used, available []string, in io.Reader, out io.Writer, interactive bool) (lang, note string, err error) {
	if requested != "" {
		return requested, "", nil
	}
	if interactive && in != nil && out != nil && len(available) > 0 {
		def := ""
		if len(used) == 1 {
			def = used[0]
		}
		return askLanguage(project, def, used, available, in, out)
	}
	if len(used) == 1 {
		return used[0], fmt.Sprintf("No --lang: using %s, the only language in this workspace (pass --lang with another to choose: %s).", used[0], otherLanguages(used[0], available)), nil
	}
	return "", "", fmt.Errorf("choose a language with --lang (%s) or a stack with --stack", strings.Join(available, ", "))
}

func askLanguage(project, def string, used, available []string, in io.Reader, out io.Writer) (string, string, error) {
	fmt.Fprintf(out, "Which language is %q?\n", project)
	for i, l := range available {
		mark := ""
		switch {
		case l == def:
			mark = "  (default: this workspace's only language)"
		case slices.Contains(used, l):
			mark = "  (already used here)"
		}
		fmt.Fprintf(out, "  %d) %s%s\n", i+1, l, mark)
	}
	prompt := "Language"
	if def != "" {
		prompt += " [" + def + "]"
	}
	reader := bufio.NewReader(in)
	for attempt := 0; attempt < 3; attempt++ {
		fmt.Fprintf(out, "%s: ", prompt)
		line, err := reader.ReadString('\n')
		answer := strings.TrimSpace(line)
		switch {
		case answer == "" && def != "":
			return def, "", nil
		case slices.Contains(available, answer):
			return answer, "", nil
		}
		if n, convErr := strconv.Atoi(answer); convErr == nil && n >= 1 && n <= len(available) {
			return available[n-1], "", nil
		}
		if err != nil {
			break
		}
		fmt.Fprintf(out, "%q is not one of: %s\n", answer, strings.Join(available, ", "))
	}
	return "", "", fmt.Errorf("choose a language with --lang (%s) or a stack with --stack", strings.Join(available, ", "))
}

// otherLanguages names the languages a person is likely to want instead of
// the one chosen (go, python, rust, node first, when installed), for the
// note; the plugins' own order would start with bicep.
func otherLanguages(chosen string, available []string) string {
	var likely []string
	for _, l := range []string{"go", "python", "rust", "node"} {
		if l != chosen && slices.Contains(available, l) {
			likely = append(likely, l)
		}
	}
	if len(likely) == 0 {
		return "`smith kinds` lists them"
	}
	return strings.Join(likely, ", ") + ", ...; `smith kinds` lists them"
}
