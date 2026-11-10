package commandflags

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func tree(build func(sub *cobra.Command)) *cobra.Command {
	root := &cobra.Command{Use: "smith"}
	sub := &cobra.Command{Use: "thing"}
	build(sub)
	root.AddCommand(sub)
	return root
}

// The guard must catch the mistakes it exists for.
func TestViolationsCatchesInventedSpellings(t *testing.T) {
	cases := []struct {
		name  string
		build func(sub *cobra.Command)
		want  string
	}{
		{"another word for the project", func(s *cobra.Command) { s.Flags().String("app", "", "") }, "--app is another name"},
		{"workspace in plain view", func(s *cobra.Command) { s.Flags().StringP("workspace", "w", "", "") }, "--workspace is another name"},
		{"language in plain view", func(s *cobra.Command) { s.Flags().String("language", "", "") }, "--language is another name"},
		{"confirm for yes", func(s *cobra.Command) { s.Flags().Bool("confirm", false, "") }, "--confirm is another name"},
		{"plan for dry-run", func(s *cobra.Command) { s.Flags().Bool("plan", false, "") }, "--plan is another name"},
		{"format for json", func(s *cobra.Command) { s.Flags().String("format", "", "") }, "--format is another name"},
		{"yes without -y", func(s *cobra.Command) { s.Flags().Bool("yes", false, "") }, `--yes has shorthand ""`},
		{"lang without its synonym", func(s *cobra.Command) { s.Flags().StringP("lang", "l", "", "") }, "missing its synonym --language"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := strings.Join(Violations(tree(c.build)), "\n")
			if !strings.Contains(got, c.want) {
				t.Fatalf("want a violation containing %q, got:\n%s", c.want, got)
			}
		})
	}
}

func TestViolationsCatchesOneShorthandForTwoFlags(t *testing.T) {
	root := &cobra.Command{Use: "smith"}
	a, b := &cobra.Command{Use: "a"}, &cobra.Command{Use: "b"}
	a.Flags().StringP("stack", "s", "", "")
	b.Flags().StringP("scope", "s", "", "")
	root.AddCommand(a, b)
	if got := strings.Join(Violations(root), "\n"); !strings.Contains(got, "-s stands for more than one flag") {
		t.Fatalf("got:\n%s", got)
	}
}

func TestViolationsAcceptsTheHelpers(t *testing.T) {
	root := tree(func(s *cobra.Command) {
		var p []string
		var l, o string
		var y, n bool
		AddProjectFlag(s, &p, "x")
		AddLangFlag(s, &l, "", "x")
		AddOutputFlag(s, &o, "", "x")
		AddYesFlag(s, &y, "x")
		AddNoInstallFlag(s, &n, "x")
	})
	for _, v := range Violations(root) {
		// the synthetic command is not in the lists of commands of the file
		if !strings.Contains(v, "flag_vocabulary.json") {
			t.Fatalf("a helper-made flag was rejected: %s", v)
		}
	}
}
