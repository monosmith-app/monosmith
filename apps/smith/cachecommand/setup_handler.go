package cachecommand

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/apps/smith/cachesetup"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

func setupCommand() *cobra.Command {
	var questions, asJSON bool
	var clearAfter time.Duration
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up the remote cache step by step: where, the key, and each CI provider's secrets",
		Long: `Set up the remote cache with a guided list of choices.

smith asks which storage (Azure Blob, S3 or an HTTPS server; Google Cloud Storage
is not supported), then helps you find the place: with the az or aws program it
lists subscriptions, accounts, containers or buckets and folders, and proposes a
sub-folder named after this workspace. It then shows the exact smith.toml change
and writes [cache] remote only after you say yes, with the comments kept. It can
make the shared key (see smith cache key), make the Azure SAS token, and prints
the steps for each provider of [ci] providers, with the secrets they need.

Every command smith runs for you is shown first and runs only after you
confirm. Nothing is committed unless you pass --commit. The key and the token
stay in memory and go to the clipboard, never into a file.

--questions --json prints the questions, for an editor to ask.`,
		Example: `  smith cache setup
  smith cache setup --commit
  smith cache setup --questions --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if questions {
				if !asJSON {
					return problem.New("--questions prints the questions as data", "add --json")
				}
				return encode(out, cachesetup.Questions())
			}
			if !onTerminal() {
				return problem.New("smith cache setup asks questions, so it needs a terminal", "run it in a terminal, or see the questions with `smith cache setup --questions --json`", "or set the address yourself with `smith config set cache.remote <address>`")
			}
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			cfg, err := workspace.Load(root)
			if err != nil {
				return err
			}
			name := cfg.Workspace.Name
			if name == "" {
				name = filepath.Base(root)
			}
			restore, raw := rawTerminal()
			if !raw {
				return problem.New("could not switch the terminal to key-press mode", "set the address yourself with `smith config set cache.remote <address>`")
			}
			defer restore()
			out = rawWriter{out}
			keys := newStdinFeed(os.Stdin)
			editor := &lineEditor{feed: keys, echo: out, interrupt: func() {
				restore()
				fmt.Fprintln(os.Stdout, "\nStopped. Nothing more was changed.")
				os.Exit(130)
			}}
			_, err = cachesetup.Run(
				cachesetup.Facts{WorkspaceName: name, Remote: cfg.Cache.Remote, Providers: cfg.CI.Providers},
				cachesetup.Ports{
					Dialogue: &cachesetup.Terminal{In: bufio.NewReader(editor), Out: out, SecretFn: func(noun, variable string, produce func() (string, error), actions []cachekey.KeyAction) error {
						s := newSession(out, clearAfter)
						s.Noun, s.Variable, s.Produce, s.Actions = noun, variable, produce, actions
						return runSecretSession(s, out, keys)
					}},
					Run: cachesetup.ExecRunner, LookPath: exec.LookPath, Now: time.Now,
					// The command's own --commit flag, through operation.Wrap, commits the change.
					WriteRemote: func(remote string, _ bool) (string, error) {
						return "", configedit.Set(root, "cache.remote", []string{remote})
					},
				})
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the questions as JSON (with --questions)")
	cmd.Flags().BoolVar(&questions, "questions", false, "print the questions smith cache setup asks, with their answers, and change nothing (needs --json)")
	cmd.Flags().DurationVar(&clearAfter, "clear-after", cachekey.DefaultClearAfter, "clear the clipboard after this long if it still holds the key; 0 leaves it")
	operation.Wrap(cmd, operation.Spec{Describe: func(*cobra.Command, []string) string { return "set up the remote cache" }})
	return cmd
}
