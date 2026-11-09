package cachecommand

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
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

// pressedKeys reads single key presses from in until it ends.
func pressedKeys(in io.Reader) <-chan rune {
	ch := make(chan rune)
	go func() {
		defer close(ch)
		buf := make([]byte, 1)
		for {
			if n, err := in.Read(buf); err != nil || n == 0 {
				return
			}
			ch <- rune(buf[0])
		}
	}()
	return ch
}

func keyCommand() *cobra.Command {
	var print bool
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
everywhere), u sets it as a user environment variable on this machine (the
command is shown and asked first), q is done. The clipboard is cleared after
--clear-after, or when you press q, only if it still holds the key. Clipboard
history tools may keep a copy.

Without a terminal the command copies the key once and ends.`,
		Example: `  smith cache key
  smith cache key --clear-after 0
  smith cache key --print`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			session := cachekey.Session{
				Random: rand.Reader, Out: out, After: time.After, ClearAfter: clearAfter,
				Env: cachekey.UserEnvironment{GOOS: runtime.GOOS, Run: cachekey.ExecRunner},
			}
			if print {
				return session.Print()
			}
			stdin, stdout := int(os.Stdin.Fd()), int(os.Stdout.Fd())
			interactive := cmd.OutOrStdout() == os.Stdout && term.IsTerminal(stdin) && term.IsTerminal(stdout)
			var terminal io.Writer
			if interactive {
				terminal = os.Stdout
			}
			session.Clip = cachekey.FindSystem(terminal)
			if !interactive {
				return session.Run()
			}
			state, err := term.MakeRaw(stdin)
			if err != nil {
				return session.Run()
			}
			defer func() { _ = term.Restore(stdin, state) }()
			session.Out = rawWriter{out}
			session.Keys = pressedKeys(os.Stdin)
			return session.Run()
		},
	}
	cmd.Flags().BoolVar(&print, "print", false, "print the key instead of copying it")
	cmd.Flags().DurationVar(&clearAfter, "clear-after", cachekey.DefaultClearAfter, "clear the clipboard after this long if it still holds the key; 0 leaves it")
	return cmd
}
