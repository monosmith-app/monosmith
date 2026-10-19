package tui

import (
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/navigator"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func init() { tickInterval = time.Microsecond }

var smithCommands = []clicontract.CommandDescription{
	{Name: "affected", Description: "List projects a change touches, before a release"},
	{Name: "doctor", Description: "Check the workspace"},
	{Name: "up", Description: "Show dependencies with newer releases and update the ones you pick"},
	{Name: "release", Description: "Version, tag and publish", Options: []clicontract.OptionDescription{{Name: "dry-run"}}},
}

var sampleProjects = []navigator.Project{
	{Name: "web", Kind: "app", Language: "node", Root: "apps/web", Stack: "node-vitest", Deps: []string{"kit"}, Affected: true,
		Tasks: []navigator.Task{{Name: "build", Command: "npm run build", Cache: true}, {Name: "test", Command: "npm run test", DependsOn: []string{"^build"}, Cache: true}}},
	{Name: "orders", Kind: "app", Language: "go", Root: "apps/orders",
		Tasks: []navigator.Task{{Name: "build", Command: "go build ./..."}, {Name: "test", Command: "go test ./..."}}},
	{Name: "kit", Kind: "lib", Language: "go", Root: "libs/kit", Dependents: []string{"web"},
		Tasks: []navigator.Task{{Name: "build", Command: "go build ./..."}, {Name: "test", Command: "go test ./..."}}},
	{Name: "sdk", Kind: "package", Language: "python", Root: "packages/sdk", Tasks: []navigator.Task{{Name: "test", Command: "pytest"}}},
}

// scripted is a process that prints its lines, then, unless hold is set,
// exits; with hold set it stays running until stopped.
type scripted struct {
	mu      sync.Mutex
	lines   []string
	exit    runner.Exit
	hold    bool
	sent    bool
	stopped bool
	release chan struct{}
}

func (p *scripted) Next() ([]string, *runner.Exit) {
	p.mu.Lock()
	if !p.sent {
		p.sent = true
		lines := p.lines
		p.mu.Unlock()
		return lines, nil
	}
	hold := p.hold
	p.mu.Unlock()
	if hold {
		<-p.release
		p.mu.Lock()
		p.exit.Cancelled = p.stopped
		p.exit.Code = map[bool]int{true: 1, false: p.exit.Code}[p.stopped]
		p.mu.Unlock()
	}
	e := p.exit
	return nil, &e
}

func (p *scripted) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.stopped {
		p.stopped = true
		if p.hold {
			close(p.release)
		}
	}
}

// lab holds a shell and the process the next command will start.
type lab struct {
	t       *testing.T
	m       tea.Model
	started [][]string
	next    *scripted
	clock   time.Time
	quit    bool
}

func (l *lab) sh() Shell { return l.m.(Shell) }

func newLab(t *testing.T, inWorkspace bool, w, h int) *lab {
	t.Helper()
	l := &lab{t: t, clock: time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)}
	report := doctor.Report{Passed: 1, Failed: 1, Findings: []doctor.Finding{
		{Finding: protocol.Finding{Check: "go.work lists every module", Detail: "libs/x is missing", Remedy: "go work use ./libs/x"}, Source: "go"},
		{Finding: protocol.Finding{Check: "git is installed", OK: true}, Source: "smith"},
	}}
	plan := release.Plan{Items: []release.Item{{Project: "web", Current: "1.0.0", Next: "1.1.0", Reason: "feat: dark mode"}}, Unchanged: []string{"sdk"}}
	cfg := Config{
		Version: "1.2.3", Workspace: "acme", Theme: theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}),
		Commands: smithCommands,
		Actions:  guide.Actions([]string{"go", "node"}, []string{"go", "node", "python"}, []string{"go", "node"}),
	}
	if inWorkspace {
		cfg.Sources = &Sources{
			Doctor:   func() (doctor.Report, error) { return report, nil },
			Release:  func() releaseLoaded { return releaseLoaded{plan: plan} },
			Projects: func() projectsLoaded { return projectsLoaded{projects: sampleProjects, branch: "main"} },
			Settings: func() settingsLoaded {
				fields, _ := configedit.Fields()
				var out settingsLoaded
				for _, f := range fields {
					switch f.Path {
					case "ci.devcontainer":
						out.settings = append(out.settings, setting{field: f, value: "true", set: true})
					case "ci.providers":
						out.settings = append(out.settings, setting{field: f, value: `["github"]`, set: true})
					case "release.branch":
						out.settings = append(out.settings, setting{field: f})
					}
				}
				return out
			},
			Plan: func(argv []string) []string {
				return []string{"kit:build", "orders:build", "web:build", "sdk:test"}
			},
			Start: func(argv []string) (runner.Handle, error) {
				l.started = append(l.started, argv)
				if l.next == nil {
					return &scripted{exit: runner.Exit{}}, nil
				}
				p := l.next
				l.next = nil
				return p, nil
			},
		}
	}
	sh := NewShell(cfg)
	sh.now = func() time.Time { return l.clock }
	l.m = sh
	l.drive(sh.Init())
	l.send(tea.WindowSizeMsg{Width: w, Height: h})
	return l
}

// drive runs the commands an update returned, as Bubble Tea would, and feeds
// the messages back. Commands that block (a tick, a process that is still
// running) are given a moment and then left.
func (l *lab) drive(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		switch msg := msg.(type) {
		case nil:
		case tea.BatchMsg:
			for _, c := range msg {
				l.drive(c)
			}
		case tickMsg:
			// Frames are sent by the tests that look at animation.
		case tea.QuitMsg:
			l.quit = true
		default:
			l.send(msg)
		}
	case <-time.After(150 * time.Millisecond):
	}
}

func (l *lab) send(msg tea.Msg) *lab {
	var cmd tea.Cmd
	l.m, cmd = l.m.Update(msg)
	l.drive(cmd)
	return l
}

// keys sends key names: "ctrl+r", "enter", "down", or text.
func (l *lab) keys(names ...string) *lab {
	for _, n := range names {
		l.send(keyMsg(n))
	}
	return l
}

func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "f1":
		return tea.KeyMsg{Type: tea.KeyF1}
	case "f2":
		return tea.KeyMsg{Type: tea.KeyF2}
	case "f3":
		return tea.KeyMsg{Type: tea.KeyF3}
	case "f10":
		return tea.KeyMsg{Type: tea.KeyF10}
	case "ctrl+o":
		return tea.KeyMsg{Type: tea.KeyCtrlO}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+g":
		return tea.KeyMsg{Type: tea.KeyCtrlG}
	case "ctrl+p":
		return tea.KeyMsg{Type: tea.KeyCtrlP}
	case "ctrl+q":
		return tea.KeyMsg{Type: tea.KeyCtrlQ}
	case "ctrl+r":
		return tea.KeyMsg{Type: tea.KeyCtrlR}
	case "ctrl+b":
		return tea.KeyMsg{Type: tea.KeyCtrlB}
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	case "ctrl+x":
		return tea.KeyMsg{Type: tea.KeyCtrlX}
	case "ctrl+n":
		return tea.KeyMsg{Type: tea.KeyCtrlN}
	case "ctrl+t":
		return tea.KeyMsg{Type: tea.KeyCtrlT}
	case "ctrl+a":
		return tea.KeyMsg{Type: tea.KeyCtrlA}
	}
	if rest, ok := strings.CutPrefix(k, "alt+"); ok {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(rest), Alt: true}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

// typeText sends each rune as a key.
func (l *lab) typeText(text string) *lab {
	for _, r := range text {
		l.send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return l
}

func (l *lab) view() string { return pane.Strip(l.sh().View()) }

func (l *lab) lines() []string { return strings.Split(l.view(), "\n") }

// has fails the test unless every text is on screen.
func (l *lab) has(texts ...string) {
	l.t.Helper()
	view := l.view()
	for _, want := range texts {
		if !strings.Contains(view, want) {
			l.t.Fatalf("missing %q on screen:\n%s", want, view)
		}
	}
}

func (l *lab) hasNot(texts ...string) {
	l.t.Helper()
	view := l.view()
	for _, bad := range texts {
		if strings.Contains(view, bad) {
			l.t.Fatalf("%q should not be on screen:\n%s", bad, view)
		}
	}
}

// find returns the screen cell where text starts on a row.
func (l *lab) find(text string) (x, y int) {
	l.t.Helper()
	for row, line := range l.lines() {
		if i := strings.Index(line, text); i >= 0 {
			return len([]rune(line[:i])), row
		}
	}
	l.t.Fatalf("%q not on screen:\n%s", text, l.view())
	return 0, 0
}

func click(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
}

func rightClick(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonRight}
}

func motion(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionMotion, Button: tea.MouseButtonNone}
}

func wheel(x, y, dir int) tea.MouseMsg {
	b := tea.MouseButtonWheelDown
	if dir < 0 {
		b = tea.MouseButtonWheelUp
	}
	return tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: b}
}

func (l *lab) clickText(text string) *lab {
	x, y := l.find(text)
	return l.send(click(x+1, y))
}

func (l *lab) advance(d time.Duration) { l.clock = l.clock.Add(d) }

// argv is the last command line the shell started.
func (l *lab) lastStarted() string {
	if len(l.started) == 0 {
		return ""
	}
	return strings.Join(l.started[len(l.started)-1], " ")
}
