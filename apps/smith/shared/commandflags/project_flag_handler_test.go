package commandflags

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func parse(t *testing.T, build func(cmd *cobra.Command), args ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "x", RunE: func(*cobra.Command, []string) error { return nil }}
	build(cmd)
	cmd.SetArgs(args)
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	if err := cmd.Execute(); err != nil {
		t.Fatalf("%v: %v", args, err)
	}
	return cmd
}

// Every spelling of the project flag gives the same projects, repeated or comma
// separated, and mixed spellings add up. The synonyms are not in --help.
func TestProjectFlagSpellings(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"canonical short", []string{"-p", "web"}, []string{"web"}},
		{"canonical long", []string{"--project", "web"}, []string{"web"}},
		{"comma separated", []string{"-p", "web,api"}, []string{"web", "api"}},
		{"repeated", []string{"-p", "web", "-p", "api"}, []string{"web", "api"}},
		{"synonym -w", []string{"-w", "web"}, []string{"web"}},
		{"synonym --workspace", []string{"--workspace", "web", "--workspace", "api"}, []string{"web", "api"}},
		{"synonym --projects", []string{"--projects", "web,api"}, []string{"web", "api"}},
		{"inline value", []string{"--project=web,api"}, []string{"web", "api"}},
		{"mixed spellings", []string{"-p", "web", "-w", "api", "--projects", "core"}, []string{"web", "api", "core"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got []string
			cmd := parse(t, func(cmd *cobra.Command) { AddProjectFlag(cmd, &got, "projects to run") }, c.args...)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
			if !cmd.Flags().Changed("project") {
				t.Fatal("Changed(project) is false although a spelling was given")
			}
		})
	}
}

func TestProjectFlagHelpShowsOneFlagAndNamesTheSynonyms(t *testing.T) {
	var got []string
	cmd := &cobra.Command{Use: "x"}
	AddProjectFlag(cmd, &got, "projects to run")
	help := cmd.Flags().FlagUsages()
	if !strings.Contains(help, "-p, --project") || !strings.Contains(help, "(also -w, --workspace, --projects)") {
		t.Fatalf("help should show -p, --project with the note:\n%s", help)
	}
	if n := strings.Count(strings.TrimSpace(help), "\n"); n != 0 {
		t.Fatalf("a synonym has its own help line:\n%s", help)
	}
}

func TestOneProjectFlagSpellings(t *testing.T) {
	for _, args := range [][]string{{"-p", "web"}, {"--project", "web"}, {"-w", "web"}, {"--workspace", "web"}, {"--projects", "web"}} {
		var got string
		parse(t, func(cmd *cobra.Command) { AddOneProjectFlag(cmd, &got, "the app") }, args...)
		if got != "web" {
			t.Fatalf("%v gave %q", args, got)
		}
	}
}

// The other shared flags: canonical and synonym both work and the synonym is
// out of --help.
func TestSharedFlagSynonyms(t *testing.T) {
	cases := []struct {
		name    string
		build   func(cmd *cobra.Command, s *string, b *bool)
		args    [][]string
		wantStr string
		wantBin bool
		hidden  []string
	}{
		{"lang", func(cmd *cobra.Command, s *string, _ *bool) { AddLangFlag(cmd, s, "", "language") },
			[][]string{{"-l", "go"}, {"--lang", "go"}, {"--language", "go"}}, "go", false, []string{"language"}},
		{"output", func(cmd *cobra.Command, s *string, _ *bool) { AddOutputFlag(cmd, s, "", "file") },
			[][]string{{"-o", "a.txt"}, {"--output", "a.txt"}, {"--out", "a.txt"}}, "a.txt", false, []string{"out"}},
		{"no-install", func(cmd *cobra.Command, _ *string, b *bool) { AddNoInstallFlag(cmd, b, "skip") },
			[][]string{{"--no-install"}, {"--skip-install"}}, "", true, []string{"skip-install"}},
		{"yes", func(cmd *cobra.Command, _ *string, b *bool) { AddYesFlag(cmd, b, "no ask") },
			[][]string{{"-y"}, {"--yes"}}, "", true, nil},
	}
	for _, c := range cases {
		for _, args := range c.args {
			t.Run(c.name+" "+strings.Join(args, " "), func(t *testing.T) {
				var s string
				var b bool
				parse(t, func(cmd *cobra.Command) { c.build(cmd, &s, &b) }, args...)
				if s != c.wantStr || b != c.wantBin {
					t.Fatalf("got %q %v", s, b)
				}
			})
		}
		t.Run(c.name+" help", func(t *testing.T) {
			var s string
			var b bool
			cmd := &cobra.Command{Use: "x"}
			c.build(cmd, &s, &b)
			for _, h := range c.hidden {
				if strings.Contains(cmd.Flags().FlagUsages(), "--"+h+" ") && !strings.Contains(cmd.Flags().FlagUsages(), "(also") {
					t.Fatalf("synonym --%s is not explained in the help", h)
				}
				if f := cmd.Flags().Lookup(h); f == nil || !f.Hidden {
					t.Fatalf("synonym --%s is not a hidden flag", h)
				}
			}
		})
	}
}
