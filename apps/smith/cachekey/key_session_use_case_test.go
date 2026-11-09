package cachekey

import (
	"bytes"
	"crypto/rand"
	"strings"
	"testing"
	"time"
)

type fakeClipboard struct {
	text     string
	readable bool
	writes   int
}

func (f *fakeClipboard) Name() string          { return "fake" }
func (f *fakeClipboard) Write(s string) error  { f.text = s; f.writes++; return nil }
func (f *fakeClipboard) Read() (string, error) { return f.text, nil }
func (f *fakeClipboard) CanRead() bool         { return f.readable }

func keysOf(runes ...rune) <-chan rune {
	ch := make(chan rune, len(runes))
	for _, r := range runes {
		ch <- r
	}
	close(ch)
	return ch
}

func newSession(clip Clipboard, keys <-chan rune, out *bytes.Buffer) Session {
	return Session{
		Random: rand.Reader, Clip: clip, Out: out, Keys: keys,
		After:      func(time.Duration) <-chan time.Time { return nil },
		ClearAfter: DefaultClearAfter,
		Env:        UserEnvironment{GOOS: "linux"},
	}
}

func TestRunCopiesTheKeyAndNeverPrintsIt(t *testing.T) {
	clip, out := &fakeClipboard{readable: true}, &bytes.Buffer{}
	if err := newSession(clip, nil, out).Run(); err != nil {
		t.Fatal(err)
	}
	if len(clip.text) != 64 || strings.Contains(out.String(), clip.text) {
		t.Fatalf("clipboard %q, output %q", clip.text, out)
	}
	if !strings.Contains(out.String(), "Key copied to the clipboard (not shown). Paste it where you need it.") {
		t.Fatalf("output %q", out)
	}
}

func TestGenerateAnotherReplacesTheKeyAndSaysSo(t *testing.T) {
	clip, out := &fakeClipboard{readable: true}, &bytes.Buffer{}
	first := ""
	s := newSession(clip, nil, out)
	ch := make(chan rune, 2)
	s.Keys = ch
	go func() {
		for clip.writes < 1 {
			time.Sleep(time.Millisecond)
		}
		first = clip.text
		ch <- 'g'
		for clip.writes < 2 {
			time.Sleep(time.Millisecond)
		}
		ch <- 'q'
	}()
	if err := s.Run(); err != nil {
		t.Fatal(err)
	}
	if first == "" || strings.Contains(out.String(), first) {
		t.Fatalf("first %q, output %q", first, out)
	}
	if !strings.Contains(out.String(), "This replaces the key shown before: use the newest one everywhere.") {
		t.Fatalf("output %q", out)
	}
}

func TestDoneClearsTheClipboardOnlyWhileItHoldsTheKey(t *testing.T) {
	for name, tc := range map[string]struct {
		change bool
		want   string
	}{"still the key": {false, ""}, "copied something else": {true, "other"}} {
		t.Run(name, func(t *testing.T) {
			clip, out := &fakeClipboard{readable: true}, &bytes.Buffer{}
			s := newSession(clip, nil, out)
			ch := make(chan rune, 1)
			s.Keys = ch
			go func() {
				for clip.writes < 1 {
					time.Sleep(time.Millisecond)
				}
				if tc.change {
					clip.text = "other"
				}
				ch <- 'q'
			}()
			if err := s.Run(); err != nil {
				t.Fatal(err)
			}
			if clip.text != tc.want {
				t.Fatalf("clipboard %q, want %q", clip.text, tc.want)
			}
		})
	}
}

func TestTheTimerClearsTheClipboard(t *testing.T) {
	clip, out := &fakeClipboard{readable: true}, &bytes.Buffer{}
	s := newSession(clip, make(chan rune), out)
	fire := make(chan time.Time, 1)
	fire <- time.Now()
	s.After = func(time.Duration) <-chan time.Time { return fire }
	if err := s.Run(); err != nil {
		t.Fatal(err)
	}
	if clip.text != "" || !strings.Contains(out.String(), "now cleared") {
		t.Fatalf("clipboard %q, output %q", clip.text, out)
	}
}

func TestAClipboardThatCannotBeReadIsNeverCleared(t *testing.T) {
	clip, out := &fakeClipboard{readable: false}, &bytes.Buffer{}
	if err := newSession(clip, keysOf('q'), out).Run(); err != nil {
		t.Fatal(err)
	}
	if len(clip.text) != 64 || clip.writes != 1 {
		t.Fatalf("clipboard %q after %d writes", clip.text, clip.writes)
	}
}

func TestNoClipboardOffersPrintingAsAChoice(t *testing.T) {
	out := &bytes.Buffer{}
	if err := newSession(nil, keysOf('p'), out).Run(); err != nil || !strings.Contains(out.String(), "No clipboard found") {
		t.Fatalf("%v %q", err, out)
	}
	if lines := strings.Split(strings.TrimSpace(out.String()), "\n"); len(lines[len(lines)-1]) != 64 {
		t.Fatalf("the key was not printed: %q", out)
	}
	if err := newSession(nil, nil, &bytes.Buffer{}).Run(); err == nil || !strings.Contains(err.Error(), "--print") {
		t.Fatalf("got %v", err)
	}
}

func TestUserVariableIsShownFirstAndAskedFirst(t *testing.T) {
	var ran []string
	env := UserEnvironment{GOOS: "windows", Run: func(stdin string, argv ...string) (string, error) {
		ran = argv
		if len(stdin) != 65 {
			t.Errorf("the key must arrive on standard input, got %q", stdin)
		}
		return "", nil
	}}
	for name, tc := range map[string]struct {
		answer rune
		runs   bool
	}{"yes": {'y', true}, "no": {'n', false}} {
		t.Run(name, func(t *testing.T) {
			ran = nil
			out := &bytes.Buffer{}
			s := newSession(&fakeClipboard{readable: true}, keysOf('u', tc.answer, 'q'), out)
			s.Env = env
			if err := s.Run(); err != nil {
				t.Fatal(err)
			}
			if (ran != nil) != tc.runs || !strings.Contains(out.String(), "SetEnvironmentVariable('SMITH_CACHE_KEY', '<the key>', 'User')") {
				t.Fatalf("ran %v, output %q", ran, out)
			}
			for _, a := range ran {
				if len(a) == 64 {
					t.Fatal("the key is in the arguments")
				}
			}
		})
	}
}

func TestOtherSystemsGetAProfileLineAndNothingRuns(t *testing.T) {
	out := &bytes.Buffer{}
	s := newSession(&fakeClipboard{readable: true}, keysOf('u', 'q'), out)
	if err := s.Run(); err != nil || !strings.Contains(out.String(), `export SMITH_CACHE_KEY="<the key>"`) {
		t.Fatalf("%v %q", err, out)
	}
}

func TestFindPrefersAnInstalledToolThenTheTerminal(t *testing.T) {
	has := func(names ...string) func(string) (string, error) {
		return func(n string) (string, error) {
			for _, h := range names {
				if h == n {
					return n, nil
				}
			}
			return "", errNone
		}
	}
	env := func(k string) string {
		if k == "WAYLAND_DISPLAY" {
			return "wayland-0"
		}
		return ""
	}
	if c := Find("linux", env, has("wl-copy", "xclip"), nil, nil); c == nil || c.Name() != "wl-copy" {
		t.Fatalf("got %v", c)
	}
	if c := Find("linux", func(string) string { return "" }, has("xclip"), nil, nil); c == nil || c.Name() != "xclip" {
		t.Fatalf("got %v", c)
	}
	if c := Find("windows", env, has("clip"), nil, nil); c == nil || c.Name() != "clip" {
		t.Fatalf("got %v", c)
	}
	if c := Find("linux", env, has(), nil, nil); c != nil {
		t.Fatalf("got %v", c)
	}
	out := &bytes.Buffer{}
	c := Find("linux", env, has(), nil, out)
	if c == nil || c.CanRead() {
		t.Fatalf("got %v", c)
	}
	_ = c.Write("abc")
	if !strings.HasPrefix(out.String(), "\x1b]52;c;") {
		t.Fatalf("got %q", out)
	}
}

type noneError struct{}

func (noneError) Error() string { return "not found" }

var errNone = noneError{}
