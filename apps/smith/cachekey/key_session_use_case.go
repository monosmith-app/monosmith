package cachekey

import (
	"fmt"
	"io"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// DefaultClearAfter is how long the key stays on the clipboard.
const DefaultClearAfter = 60 * time.Second

// Session is one run of `smith cache key`: it makes a key, copies it and, on a
// terminal, answers one-key choices until the person is done.
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
	// ClearAfter is how long the key stays on the clipboard; zero or less
	// leaves it there.
	ClearAfter time.Duration
	// Env sets the key as a user environment variable.
	Env UserEnvironment
}

// Print makes a key and prints it: the way out when there is no clipboard.
//
// Errors: the random source failed or the output could not be written.
func (s Session) Print() error {
	key, err := Generate(s.Random)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(s.Out, key)
	return err
}

// Run makes a key, copies it, and then offers the one-key choices.
//
// Errors: the random source failed, or without a clipboard and without a
// terminal there is nowhere to put the key.
func (s Session) Run() error {
	key, err := Generate(s.Random)
	if err != nil {
		return err
	}
	if s.Clip == nil {
		return s.runWithoutClipboard(key)
	}
	if err := s.Clip.Write(key); err != nil {
		return problem.New("could not copy the key: "+err.Error(), "run `smith cache key --print` to print it instead")
	}
	s.say("Key copied to the clipboard (not shown). Paste it where you need it.")
	timer := s.arm()
	if s.Keys == nil {
		if s.ClearAfter > 0 {
			s.say("Without a terminal the clipboard is not cleared for you: clear it when you have pasted the key.")
		}
		return nil
	}
	for {
		s.say("  c copy again   g generate another   u set as a user environment variable   q done")
		select {
		case <-timer:
			s.clear(key, "The clipboard held the key for "+s.ClearAfter.String()+" and is now cleared.")
			return nil
		case r, open := <-s.Keys:
			if !open {
				s.clear(key, "")
				return nil
			}
			switch r {
			case 'c', 'C':
				if err := s.Clip.Write(key); err != nil {
					s.say("Could not copy the key again: " + err.Error())
					continue
				}
				s.say("Key copied again (not shown).")
				timer = s.arm()
			case 'g', 'G':
				if key, err = Generate(s.Random); err != nil {
					return err
				}
				if err := s.Clip.Write(key); err != nil {
					return problem.New("could not copy the new key: "+err.Error(), "run `smith cache key --print` to print it instead")
				}
				s.say("New key copied (not shown). This replaces the key shown before: use the newest one everywhere.")
				timer = s.arm()
			case 'u', 'U':
				s.setUserVariable(key)
			case 'q', 'Q', 3, 27:
				s.clear(key, "Done.")
				return nil
			}
		}
	}
}

// runWithoutClipboard handles a machine with no clipboard: the key is never
// printed unasked, so the person chooses.
func (s Session) runWithoutClipboard(key string) error {
	s.say("No clipboard found here (no clip, pbcopy, wl-copy, xclip or xsel, and no terminal that takes a copy request).")
	if s.Keys == nil {
		return problem.New("there is no clipboard to put the key on", "run `smith cache key --print` to print it")
	}
	s.say("  p print the key here   q quit")
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
	s.say(fmt.Sprintf("The clipboard clears in %s if it still holds the key. Clipboard-history tools may keep a copy.", s.ClearAfter))
	return s.After(s.ClearAfter)
}

// clear empties the clipboard only when it still holds the key, so something
// the person copied since is never lost.
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
	s.say("  " + s.Env.Preview())
	if !s.Env.CanSet() {
		s.say("smith does not edit shell profiles. Add that line to yours, with the key pasted in, then open a new terminal.")
		return
	}
	s.say("Run it? y yes, any other key no")
	if r, open := <-s.Keys; !open || (r != 'y' && r != 'Y') {
		s.say("Not run.")
		return
	}
	if err := s.Env.Set(key); err != nil {
		s.say("Could not set it: " + err.Error())
		return
	}
	s.say("SMITH_CACHE_KEY is set for your user. Open a new terminal to see it.")
}

func (s Session) say(line string) { fmt.Fprintln(s.Out, line) }
