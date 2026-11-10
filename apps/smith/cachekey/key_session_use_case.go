package cachekey

import (
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// DefaultClearAfter is how long the key stays on the clipboard.
const DefaultClearAfter = 60 * time.Second

// Session is one run of `smith cache key`: it makes a secret, copies it and,
// on a terminal, answers one-key choices until the person is done.
type Session struct {
	// Random is the source of random bytes (crypto/rand.Reader).
	Random io.Reader
	// Clip is the clipboard; nil when there is none.
	Clip Clipboard
	// Out is where messages go.
	Out io.Writer
	// Keys delivers the keys the person presses; nil without a terminal, and
	// then the session copies once and ends. A closed channel means done.
	Keys <-chan rune
	// After returns a channel that fires after a delay (time.After).
	After func(time.Duration) <-chan time.Time
	// ClearAfter is how long the secret stays on the clipboard; zero or less
	// leaves it there.
	ClearAfter time.Duration
	// Env sets the key as a user environment variable.
	Env UserEnvironment
	// Noun names the secret in messages; "Key" when empty.
	Noun string
	// Produce makes a secret; Generate(Random) when nil. A caller that makes
	// another kind of secret, such as a SAS token, passes its own.
	Produce func() (string, error)
	// Actions are more one-key choices after the built-in ones, such as
	// sending the key to a CI provider's secret store.
	Actions []KeyAction
	// NoPrompt leaves out the line that lists the one-key choices, for a
	// screen that draws them as buttons.
	NoPrompt bool
}

// KeyAction is one more choice of the key prompt.
type KeyAction struct {
	// Key is the key to press.
	Key rune
	// Label is the choice's text in the prompt.
	Label string
	// Run does the work. confirm shows the exact command line and asks; it
	// returns true only on a yes. secret must reach a program only through
	// its standard input.
	Run func(secret string, confirm func(command string) bool)
}

func (s Session) noun() string {
	if s.Noun == "" {
		return "Key"
	}
	return s.Noun
}

func (s Session) make() (string, error) {
	if s.Produce != nil {
		return s.Produce()
	}
	return Generate(s.Random)
}

// confirm shows a command and asks; only y runs it.
func (s Session) confirm(command string) bool {
	s.say("  " + command)
	s.say("Run it? y yes, any other key no")
	if r, open := <-s.Keys; !open || (r != 'y' && r != 'Y') {
		s.say("Not run.")
		return false
	}
	return true
}

// Print makes a key and prints it: the way out when there is no clipboard.
//
// Errors: the random source failed or the output could not be written.
func (s Session) Print() error {
	key, err := s.make()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(s.Out, key)
	return err
}

// Run makes a secret, copies it, and then offers the one-key choices.
//
// Errors: the random source failed, or without a clipboard and without a
// terminal there is nowhere to put the secret.
func (s Session) Run() error {
	key, err := s.make()
	if err != nil {
		return err
	}
	if s.Clip == nil {
		return s.runWithoutClipboard(key)
	}
	if err := s.Clip.Write(key); err != nil {
		return problem.New("could not copy the "+strings.ToLower(s.noun())+": "+err.Error(), "run `smith cache key --print` to print a key instead")
	}
	s.say(s.noun() + " copied to the clipboard (not shown). Paste it where you need it.")
	timer := s.arm()
	if s.Keys == nil {
		if s.ClearAfter > 0 {
			s.say("Without a terminal the clipboard is not cleared for you: clear it when you have pasted the key.")
		}
		return nil
	}
	for {
		if !s.NoPrompt {
			s.say(s.promptLine())
		}
		select {
		case <-timer:
			s.clear(key, "The clipboard held the "+strings.ToLower(s.noun())+" for "+s.ClearAfter.String()+" and is now cleared.")
			return nil
		case r, open := <-s.Keys:
			if !open {
				s.clear(key, "")
				return nil
			}
			switch r {
			case 'c', 'C':
				if err := s.Clip.Write(key); err != nil {
					s.say("Could not copy it again: " + err.Error())
					continue
				}
				s.say(s.noun() + " copied again (not shown).")
				timer = s.arm()
			case 'g', 'G':
				if key, err = s.make(); err != nil {
					return err
				}
				if err := s.Clip.Write(key); err != nil {
					return problem.New("could not copy the new "+strings.ToLower(s.noun())+": "+err.Error(), "run `smith cache key --print` to print a key instead")
				}
				s.say("New " + strings.ToLower(s.noun()) + " copied (not shown). This replaces the " + strings.ToLower(s.noun()) + " shown before: use the newest one everywhere.")
				timer = s.arm()
			case 'u', 'U':
				if s.noun() == "Key" {
					s.setUserVariable(key)
				}
			case 'q', 'Q', 3, 27:
				s.clear(key, "Done.")
				return nil
			default:
				for _, a := range s.Actions {
					if unicode.ToLower(a.Key) == unicode.ToLower(r) {
						a.Run(key, s.confirm)
					}
				}
			}
		}
	}
}

// runWithoutClipboard handles a machine with no clipboard: the secret is never
// printed unasked, so the person chooses.
func (s Session) runWithoutClipboard(key string) error {
	s.say("No clipboard found here (no clip, pbcopy, wl-copy, xclip or xsel, and no terminal that takes a copy request).")
	if s.Keys == nil {
		return problem.New("there is no clipboard to put it on", "run `smith cache key --print` to print a key")
	}
	s.say("  p print it here   q quit")
	for r := range s.Keys {
		switch r {
		case 'p', 'P':
			s.say(key)
			return nil
		case 'q', 'Q', 3, 27:
			return nil
		}
	}
	return nil
}

// arm starts the timer that clears the clipboard; a nil channel never fires.
func (s Session) arm() <-chan time.Time {
	if s.ClearAfter <= 0 || !s.Clip.CanRead() || s.After == nil {
		return nil
	}
	s.say(fmt.Sprintf("The clipboard clears in %s if it still holds the %s. Clipboard-history tools may keep a copy.", s.ClearAfter, strings.ToLower(s.noun())))
	return s.After(s.ClearAfter)
}

// clear empties the clipboard only when it still holds the secret, so
// something the person copied since is never lost.
func (s Session) clear(key, message string) {
	if s.ClearAfter > 0 && s.Clip.CanRead() {
		if now, err := s.Clip.Read(); err == nil && now == key {
			if err := s.Clip.Write(""); err == nil {
				switch message {
				case "":
					message = "Clipboard cleared."
				case "Done.":
					message = "Done. The clipboard is cleared."
				}
			}
		}
	}
	if message != "" {
		s.say(message)
	}
}

// setUserVariable shows the exact command, asks, and only then runs it.
func (s Session) setUserVariable(key string) {
	s.say("This sets SMITH_CACHE_KEY for your user on this machine:")
	if !s.Env.CanSet() {
		s.say("  " + s.Env.Preview())
		s.say("smith does not edit shell profiles. Add that line to yours, with the key pasted in, then open a new terminal.")
		return
	}
	if !s.confirm(s.Env.Preview()) {
		return
	}
	if err := s.Env.Set(key); err != nil {
		s.say("Could not set it: " + err.Error())
		return
	}
	s.say("SMITH_CACHE_KEY is set for your user. Open a new terminal to see it.")
}

func (s Session) say(line string) { fmt.Fprintln(s.Out, line) }

// promptLine lists the choices.
func (s Session) promptLine() string {
	line := "  c copy again   g generate another"
	if s.noun() == "Key" {
		line += "   u set as a user environment variable"
	}
	for _, a := range s.Actions {
		line += "   " + string(a.Key) + " " + a.Label
	}
	return line + "   q done"
}
