package cachewizard

import (
	"crypto/rand"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/apps/smith/cachesetup"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
)

// closed is what a bridge panics with when the person closed the wizard, to
// unwind the engine out of a call that waits for an answer.
type closed struct{}

// Bridge runs the setup engine off the interface's goroutine and is its
// Dialogue: each question becomes an event for the screen, and the call waits
// for the answer. The interface and the CLI therefore show the same engine.
type Bridge struct {
	// Events are the engine's news, in order. The host reads them one by one.
	Events  chan Event
	secrets Secrets

	replies chan Reply
	cancel  chan struct{}
	once    sync.Once

	mu      sync.Mutex
	presses chan rune
}

// Secrets are the connections of the secret screen to the machine; tests pass
// fakes, so no test touches a real clipboard or environment.
type Secrets struct {
	// Clip is the clipboard; nil when there is none.
	Clip cachekey.Clipboard
	// Env sets a user environment variable.
	Env cachekey.UserEnvironment
	// ClearAfter is how long a secret stays on the clipboard.
	ClearAfter time.Duration
	// After is the timer (time.After when nil).
	After func(time.Duration) <-chan time.Time
}

// SystemSecrets are the real connections of this machine.
//
// Parameters: terminal is where the OSC 52 copy goes when no clipboard program
// exists (nil for none).
func SystemSecrets(terminal io.Writer) Secrets {
	return Secrets{
		Clip: cachekey.FindSystem(terminal), ClearAfter: cachekey.DefaultClearAfter,
		Env: cachekey.UserEnvironment{GOOS: runtime.GOOS, Run: cachekey.ExecRunner, Machine: shellenv.OSEnvironment{}},
	}
}

// NewBridge makes a bridge.
//
// Parameters: secrets are the clipboard, the environment and the clearing time
// of the secret screen.
func NewBridge(secrets Secrets) *Bridge {
	if secrets.After == nil {
		secrets.After = time.After
	}
	return &Bridge{Events: make(chan Event, 64), secrets: secrets, replies: make(chan Reply, 1), cancel: make(chan struct{})}
}

// Start runs the engine in the background. run gets the bridge as its Dialogue
// and returns what cachesetup.Run returns; the last event is an EndEvent.
func (b *Bridge) Start(run func(cachesetup.Dialogue) (string, error)) {
	go func() {
		var end EndEvent
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(closed); !ok {
					panic(r)
				}
				end = EndEvent{Cancelled: true}
			}
			select {
			case b.Events <- end:
			case <-b.cancel:
			}
		}()
		end.Remote, end.Err = run(b)
	}()
}

// Done is closed when the wizard was closed.
func (b *Bridge) Done() <-chan struct{} { return b.cancel }

// Reply gives the engine the answer it waits for.
func (b *Bridge) Reply(r Reply) {
	select {
	case b.replies <- r:
	default:
	}
}

// Press gives the secret session a key press.
func (b *Bridge) Press(r rune) {
	b.mu.Lock()
	ch := b.presses
	b.mu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- r:
	default:
	}
}

// Cancel closes the wizard: the engine stops at its next question, and a
// secret session ends and clears the clipboard if it still holds the secret.
func (b *Bridge) Cancel() {
	b.once.Do(func() {
		close(b.cancel)
		b.mu.Lock()
		defer b.mu.Unlock()
		if b.presses != nil {
			close(b.presses)
			b.presses = nil
		}
	})
}

func (b *Bridge) emit(ev Event) {
	select {
	case b.Events <- ev:
	case <-b.cancel:
		panic(closed{})
	}
}

func (b *Bridge) ask(s Screen) Reply {
	b.emit(AskEvent{Screen: s})
	select {
	case r := <-b.replies:
		return r
	case <-b.cancel:
		panic(closed{})
	}
}

// Say implements cachesetup.Dialogue.
func (b *Bridge) Say(text string) { b.emit(SayEvent{Text: text}) }

// Choose implements cachesetup.Dialogue.
func (b *Bridge) Choose(prompt string, options []cachesetup.Option, def string) string {
	return b.ask(Screen{Kind: Choose, Prompt: prompt, Options: options, Default: def}).Value
}

// Text implements cachesetup.Dialogue.
func (b *Bridge) Text(prompt string) string {
	return b.ask(Screen{Kind: Text, Prompt: prompt}).Value
}

// Confirm implements cachesetup.Dialogue.
func (b *Bridge) Confirm(c cachesetup.Confirmation) (bool, bool) {
	r := b.ask(Screen{Kind: Confirm, Confirmation: c})
	return r.Yes, r.Commit
}

// Working implements cachesetup.Dialogue: the screen shows a spinner and the
// words until the returned function runs.
func (b *Bridge) Working(text string) func() {
	b.emit(AskEvent{Screen: Screen{Kind: Working, Prompt: text}})
	return func() { b.emit(IdleEvent{}) }
}

// Show implements cachesetup.Dialogue.
func (b *Bridge) Show(title string, lines []string) {
	b.ask(Screen{Kind: Page, Prompt: title, Lines: lines})
}

// Secret implements cachesetup.Dialogue: the secret is made, copied without
// being shown, and held in the session's memory only while the screen is open.
// The session is the one `smith cache key` uses, so the clipboard, the
// clearing after a delay (only when the clipboard still holds the secret) and
// the confirmed commands are the same.
func (b *Bridge) Secret(noun, variable string, produce func() (string, error), actions []cachekey.KeyAction) error {
	presses := make(chan rune, 16)
	b.mu.Lock()
	b.presses = presses
	b.mu.Unlock()
	b.emit(AskEvent{Screen: Screen{Kind: Secret, Noun: noun, Buttons: buttonsFor(variable != "", actions)}})
	err := cachekey.Session{
		Random: rand.Reader, Clip: b.secrets.Clip, Out: lineWriter{b}, Keys: presses,
		After: b.secrets.After, ClearAfter: b.secrets.ClearAfter, Env: b.secrets.Env,
		Noun: noun, Variable: variable, Produce: produce, Actions: actions, NoPrompt: true,
	}.Run()
	b.emit(IdleEvent{})
	return err
}

// buttonsFor are the choices of the secret screen, the ones the session reads.
func buttonsFor(hasVariable bool, actions []cachekey.KeyAction) []Button {
	buttons := []Button{{'c', "Copy again"}, {'g', "Generate another"}}
	if hasVariable {
		buttons = append(buttons, Button{'u', "Set as a user environment variable"})
	}
	for _, a := range actions {
		buttons = append(buttons, Button{a.Key, strings.ToUpper(a.Label[:1]) + a.Label[1:]})
	}
	return append(buttons, Button{'q', "Done"})
}

// lineWriter turns what the session prints into lines the screen shows. The
// session never prints a secret unasked.
type lineWriter struct{ b *Bridge }

func (w lineWriter) Write(p []byte) (int, error) {
	for _, l := range strings.Split(string(p), "\n") {
		if l = strings.TrimRight(l, "\r"); strings.TrimSpace(l) != "" {
			w.b.Say(strings.TrimSpace(l))
		}
	}
	return len(p), nil
}

var _ cachesetup.Dialogue = (*Bridge)(nil)
