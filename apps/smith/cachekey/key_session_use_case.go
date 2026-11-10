package cachekey

import (
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/libs/words/words"
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
	// Variable is the environment variable that holds the secret, such as
	// SMITH_CACHE_KEY. When it is set, the u choice keeps the secret for the
	// user under that name, and the messages name it. Empty offers no u choice.
	Variable string
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

// InSentence is the noun as it reads inside a sentence: lower case, except an
// acronym such as "SAS token", which keeps its capitals.
func InSentence(noun string) string {
	if len(noun) > 1 && noun[1] >= 'A' && noun[1] <= 'Z' {
		return noun
	}
	return strings.ToLower(noun)
}

func (s Session) inSentence() string { return InSentence(s.noun()) }

func (s Session) make() (string, error) {
	if s.Produce != nil {
		return s.Produce()
	}
	return Generate(s.Random)
}

// copiedLine tells the person where the secret goes next.
func (s Session) copiedLine() string {
	if s.Variable == "" {
		return words.Say("cachekey.copied.plain", "noun", s.noun())
	}
	return words.Say("cachekey.copied", "noun", s.noun(), "variable", s.Variable)
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
		return problem.New("could not copy the "+s.inSentence()+": "+err.Error(), "run `smith cache key --print` to print a key instead")
	}
	s.say(s.copiedLine())
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
			s.clear(key, "The clipboard held the "+s.inSentence()+" for "+s.ClearAfter.String()+" and is now cleared.")
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
					return problem.New("could not copy the new "+s.inSentence()+": "+err.Error(), "run `smith cache key --print` to print a key instead")
				}
				s.say("New " + s.inSentence() + " copied (not shown). This replaces the " + s.inSentence() + " shown before: use the newest one everywhere.")
				timer = s.arm()
			case 'u', 'U':
				if s.Variable != "" {
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
	s.say(fmt.Sprintf("The clipboard clears in %s if it still holds the %s. Clipboard-history tools may keep a copy.", s.ClearAfter, s.inSentence()))
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

// setUserVariable keeps the key for the user, only after the person confirmed.
// On Windows it shows the exact command and asks. Elsewhere it shows the
// profile file and the lines with the key hidden, and the person chooses to
// write them, to cancel, or to do it by hand.
func (s Session) setUserVariable(key string) {
	if s.Env.CanSet() {
		s.say(words.Say("cachekey.env.windows.intro", "variable", s.Variable))
		if !s.confirm(s.Env.Preview(s.Variable)) {
			return
		}
		if err := s.Env.Set(s.Variable, key); err != nil {
			s.say(words.Say("cachekey.env.windows.failed", "reason", err.Error()))
			return
		}
		s.say(words.Say("cachekey.env.windows.done", "variable", s.Variable))
		return
	}
	t, known := s.Env.Profile()
	if !known {
		s.sayByHand(t, known)
		return
	}
	s.say(words.Say("cachekey.profile.intro", "path", t.Path, "variable", s.Variable))
	s.say("  " + t.Line(s.Variable, MaskedKey))
	s.say(words.Say("cachekey.profile.warning"))
	s.say(words.Say("cachekey.profile.choices"))
	switch s.profileChoice() {
	case "write":
		if err := WriteProfileKey(s.Env.Machine, t, s.Variable, key); err != nil {
			s.say(err.Error())
			return
		}
		s.say(words.Say("cachekey.profile.written", "path", t.Path, "variable", s.Variable))
	case "hand":
		s.sayByHand(t, known)
	default:
		s.say(words.Say("cachekey.profile.cancelled"))
	}
}

// profileChoice waits for the answer to the profile question: "write" on
// Enter, "hand" on h, and "cancel" on Esc, Ctrl+C, q or n, or when the keys
// end. Any other key is ignored, so no stray key writes the file.
func (s Session) profileChoice() string {
	for r := range s.Keys {
		switch unicode.ToLower(r) {
		case '\r', '\n':
			return "write"
		case 'h':
			return "hand"
		case 27, 3, 'q', 'n':
			return "cancel"
		}
	}
	return "cancel"
}

// sayByHand shows the line to add, with a placeholder where the key goes, for
// the person who writes the profile without smith.
func (s Session) sayByHand(t ProfileTarget, known bool) {
	if known {
		s.say(words.Say("cachekey.byhand.file", "path", t.Path))
	} else {
		s.say(words.Say("cachekey.byhand.other"))
	}
	s.say("  " + s.Env.Preview(s.Variable))
	s.say(words.Say("cachekey.byhand.next"))
}

func (s Session) say(line string) { fmt.Fprintln(s.Out, line) }

// promptLine lists the choices.
func (s Session) promptLine() string {
	line := "  c copy again   g generate another"
	if s.Variable != "" {
		line += "   u set as a user environment variable"
	}
	for _, a := range s.Actions {
		line += "   " + string(a.Key) + " " + a.Label
	}
	return line + "   q done"
}
