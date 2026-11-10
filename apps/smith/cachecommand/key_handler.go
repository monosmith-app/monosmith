package cachecommand

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// rawWriter turns each newline into a carriage return and a newline, because
// a terminal in raw mode does not.
type rawWriter struct{ w io.Writer }

func (r rawWriter) Write(p []byte) (int, error) {
	if _, err := r.w.Write(bytes.ReplaceAll(p, []byte("\n"), []byte("\r\n"))); err != nil {
		return 0, err
	}
	return len(p), nil
}

// onTerminal is true when both standard input and output are a terminal.
func onTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// runSecretSession copies a secret and runs its one-key choices.
//
// Parameters: session has everything but the clipboard, the output and the
// key source; out is the output (a rawWriter when the terminal is raw); keys
// is the shared input of a raw terminal, nil when there is no terminal.
func runSecretSession(session cachekey.Session, out io.Writer, keys *stdinFeed) error {
	session.Out = out
	var terminal io.Writer
	if keys != nil {
		terminal = os.Stdout
		session.Keys = keys.Keys()
	}
	session.Clip = cachekey.FindSystem(terminal)
	return session.Run()
}

// newSession is a Session with the system's random source, timer and
// environment.
func newSession(out io.Writer, clearAfter time.Duration) cachekey.Session {
	return cachekey.Session{
		Random: rand.Reader, Out: out, After: time.After, ClearAfter: clearAfter,
		Env: systemUserEnvironment(),
	}
}

// systemUserEnvironment is the user environment of this machine.
func systemUserEnvironment() cachekey.UserEnvironment {
	return cachekey.UserEnvironment{GOOS: runtime.GOOS, Run: cachekey.ExecRunner, Machine: shellenv.OSEnvironment{}}
}

func keyCommand() *cobra.Command {
	var printKey, remove bool
	var clearAfter time.Duration
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Make a key (SMITH_CACHE_KEY) for the remote cache and copy it, unseen",
		Long: `Make the shared secret the remote cache is signed with, and put it on the
clipboard without showing it.

SMITH_CACHE_KEY is a random value you generate once and give to everyone and
every pipeline that uses the remote cache. smith signs every cached object with
it (HMAC-SHA256) and checks the signature before it uses the object. Without the
key the remote cache is off. smith reads it from the environment only, never
from smith.toml or a .env file, and this command never writes it to a file.

The key is 32 bytes from the operating system's secure random source, as 64 hex
characters. It goes to the clipboard through clip, pbcopy, wl-copy, xclip or
xsel, or through the terminal itself (OSC 52). On a terminal you then choose:
c copies it again, g makes another key (the old one is then wrong: use the newest
everywhere), u keeps it for your user on this machine, q is done. On Windows u
shows the command and asks first, then sets a user environment variable. On
macOS and Linux u shows the profile file of your shell (zsh, bash or fish) and
the lines with the key hidden, and writes a marked block there only when you
press Enter. "smith cache key --remove" undoes it. The clipboard is cleared after
--clear-after, or when you press q, only if it still holds the key. Clipboard
history tools may keep a copy.

Without a terminal the command copies the key once and ends.`,
		Example: `  smith cache key
  smith cache key --clear-after 0
  smith cache key --print`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if remove {
				return cachekey.RemoveKey(systemUserEnvironment(), out)
			}
			session := newSession(out, clearAfter)
			if printKey {
				return session.Print()
			}
			if out == os.Stdout && onTerminal() {
				if restore, ok := rawTerminal(); ok {
					defer restore()
					return runSecretSession(session, rawWriter{out}, newStdinFeed(os.Stdin))
				}
			}
			return runSecretSession(session, out, nil)
		},
	}
	cmd.Flags().BoolVar(&remove, "remove", false, "remove the key that the u choice kept for your user (the profile block, or the Windows user variable)")
	cmd.Flags().BoolVar(&printKey, "print", false, "print the key instead of copying it")
	cmd.Flags().DurationVar(&clearAfter, "clear-after", cachekey.DefaultClearAfter, "clear the clipboard after this long if it still holds the key; 0 leaves it")
	return cmd
}
