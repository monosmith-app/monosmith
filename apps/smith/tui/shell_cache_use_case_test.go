package tui

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/apps/smith/cachesetup"
	"github.com/monosmith-app/monosmith/apps/smith/tui/cachewizard"
)

// fakeClip is a clipboard in memory.
type fakeClip struct{ text string }

func (c *fakeClip) Name() string          { return "fake" }
func (c *fakeClip) Write(t string) error  { c.text = t; return nil }
func (c *fakeClip) Read() (string, error) { return c.text, nil }
func (c *fakeClip) CanRead() bool         { return true }

// cacheLab is a shell whose remote cache setup talks to fakes: no cloud, no
// clipboard program and no smith.toml is touched.
type cacheLab struct {
	*lab
	clip    *fakeClip
	ran     []string
	stdin   []string
	written []string
	commits []bool
}

func newCacheLab(t *testing.T) *cacheLab {
	t.Helper()
	c := &cacheLab{lab: newLab(t, true, 110, 34), clip: &fakeClip{}}
	sh := c.sh()
	sh.cacheSecrets = func() cachewizard.Secrets {
		return cachewizard.Secrets{Clip: c.clip, ClearAfter: time.Minute, Env: cachekey.UserEnvironment{GOOS: "linux"}, After: func(time.Duration) <-chan time.Time { return nil }}
	}
	sh.sources.CacheSetup = func() (cachesetup.Facts, cachesetup.Ports, error) {
		return cachesetup.Facts{WorkspaceName: "acme", Providers: []string{"github"}}, cachesetup.Ports{
			Run: func(stdin string, argv ...string) (string, string, error) {
				c.ran, c.stdin = append(c.ran, cachesetup.CommandLine(argv)), append(c.stdin, stdin)
				if argv[0] == "az" {
					time.Sleep(30 * time.Millisecond)
					return `[{"name":"Prod","id":"sub-1","isDefault":true}]`, "", nil
				}
				return "", "", nil
			},
			LookPath: func(string) (string, error) { return "x", nil },
			Now:      func() time.Time { return time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC) },
			WriteRemote: func(r string, commit bool) (string, error) {
				c.written, c.commits = append(c.written, r), append(c.commits, commit)
				return "", nil
			},
		}, nil
	}
	c.m = sh
	return c
}

// open starts the setup the way the palette does, and shows what the engine
// says. The shell's own wait for news is not driven (it would race the test).
func (c *cacheLab) open() {
	sh := c.sh()
	sh.openCacheSetup()
	c.m = sh
	c.pump()
}

// pump hands the shell the engine's news until it has none for a moment.
func (c *cacheLab) pump() {
	b := c.sh().cacheRun
	for b != nil {
		select {
		case ev := <-b.Events:
			c.m, _ = c.m.Update(cacheMsg{run: b, ev: ev})
		case <-time.After(150 * time.Millisecond):
			return
		}
	}
}

// press sends keys to the open setup, then lets the engine answer.
func (c *cacheLab) press(keys ...string) {
	for _, k := range keys {
		c.m, _ = c.m.Update(keyMsg(k))
		c.pump()
	}
}

func (c *cacheLab) text(s string) {
	for _, r := range s {
		c.m, _ = c.m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestCacheSetupWalksTheHTTPSPathAndTheKeyIsNeverShown(t *testing.T) {
	c := newCacheLab(t)
	c.open()
	c.has("Set up the remote cache", "What kind of remote storage?", "Azure Blob", "HTTPS server")
	c.press("3")
	c.has("Address of the cache server")
	c.text("https://cache.acme.dev/smith")
	c.press("enter")
	c.has("This changes smith.toml:", `remote = "https://cache.acme.dev/smith"`, "[Yes, write it (y)]")
	if len(c.written) != 0 {
		t.Fatal("nothing is written before the yes")
	}
	c.press("y")
	if len(c.written) != 1 || c.written[0] != "https://cache.acme.dev/smith" {
		t.Fatalf("written %v", c.written)
	}
	c.has("Make the shared key")
	c.press("y")

	key := c.clip.text
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(key) {
		t.Fatalf("the key reaches the clipboard: %q", key)
	}
	c.has("Key ready", "[Copy again (c)]", "[Generate another (g)]", "[Send to GitHub Actions (gh) (h)]", "not shown here")
	c.hasNot(key)

	// Copy again copies the same key; Generate another replaces it.
	c.clip.text = "something else"
	c.press("c")
	if c.clip.text != key {
		t.Fatal("copy again puts the same key back")
	}
	c.press("g")
	if c.clip.text == key || len(c.clip.text) != 64 {
		t.Fatalf("generate another makes a new key: %q", c.clip.text)
	}
	key = c.clip.text
	c.hasNot(key)

	// The command is shown first; only y runs it, with the key on stdin.
	c.press("h")
	c.has("gh secret set SMITH_CACHE_KEY", "Run it?")
	if len(c.ran) != 0 {
		t.Fatalf("nothing runs before the y: %v", c.ran)
	}
	c.press("y")
	if len(c.ran) != 1 || c.stdin[0] != key+"\n" {
		t.Fatalf("ran %v stdin %q", c.ran, c.stdin)
	}
	c.hasNot(key)

	// Done clears the clipboard because it still holds the key, then the steps come.
	c.press("q")
	if c.clip.text != "" {
		t.Fatalf("the clipboard still holds %q", c.clip.text)
	}
	c.has("GitHub Actions", "Settings, Secrets and variables", "[Next (Enter)]")
	c.press("enter")
	c.has("Developer machines")
	c.press("enter")
	if c.sh().cache != nil {
		t.Fatal("the setup closes when the steps end")
	}
}

func TestCacheSetupShowsASpinnerWhileACloudListingRunsAndTheCommandFirst(t *testing.T) {
	c := newCacheLab(t)
	c.open()
	c.press("enter") // Azure Blob
	c.has("smith will run this command. It only reads:", "az account list", "[Run it (y)]")
	if len(c.ran) != 0 {
		t.Fatal("az does not run before the y")
	}
	c.m, _ = c.m.Update(keyMsg("y"))
	// The engine sits in az for 30 ms: the screen already says so.
	b := c.sh().cacheRun
	ev := <-b.Events
	c.m, _ = c.m.Update(cacheMsg{run: b, ev: ev})
	c.has("Listing with az")
	c.pump()
	c.has("Which Azure subscription?", "Prod")
}

func TestCacheSetupClickOnAButtonDoesWhatTheKeyDoes(t *testing.T) {
	c := newCacheLab(t)
	c.open()
	c.press("3")
	c.text("https://cache.acme.dev/x")
	c.press("enter", "n") // do not write
	c.press("y")          // make the key
	c.clip.text = ""
	c.clickText("[Copy again (c)]")
	c.pump()
	if len(c.clip.text) != 64 {
		t.Fatalf("a click on Copy again copies: %q", c.clip.text)
	}
}

func TestCacheSetupEscapeClosesAndTheEngineStops(t *testing.T) {
	c := newCacheLab(t)
	c.open()
	c.press("esc")
	if c.sh().cache != nil || len(c.written) != 0 {
		t.Fatal("esc closes the setup and writes nothing")
	}
}

func TestACacheSetupThatCannotStartClosesAndTellsWhy(t *testing.T) {
	c := newCacheLab(t)
	sh := c.sh()
	sh.sources.CacheSetup = func() (cachesetup.Facts, cachesetup.Ports, error) {
		return cachesetup.Facts{}, cachesetup.Ports{}, errors.New("no smith.toml here")
	}
	c.m = sh
	c.open()
	if c.sh().cache != nil {
		t.Fatal("an engine that cannot start closes the screen")
	}
	if !strings.Contains(c.sh().toast.Text, "no smith.toml here") {
		t.Fatalf("toast %q", c.sh().toast.Text)
	}
}

func TestTheWriteOffersAnUntickedCommitBoxAndTheTickReachesTheWrite(t *testing.T) {
	for _, tick := range []bool{false, true} {
		c := newCacheLab(t)
		c.open()
		c.press("3")
		c.text("https://cache.acme.dev/smith")
		c.press("enter")
		c.has("[ ] Commit this change so it can be undone")
		if tick {
			c.press("c")
			c.has("[x] Commit this change so it can be undone")
		}
		c.press("y")
		if len(c.written) != 1 || c.commits[0] != tick {
			t.Fatalf("tick %v: written %v commit %v", tick, c.written, c.commits)
		}
	}
}

func TestClickingTheCommitBoxTicksIt(t *testing.T) {
	c := newCacheLab(t)
	c.open()
	c.press("3")
	c.text("https://cache.acme.dev/smith")
	c.press("enter")
	c.clickText("[ ] Commit this change so it can be undone")
	c.has("[x] Commit this change")
	c.press("y")
	if len(c.commits) != 1 || !c.commits[0] {
		t.Fatalf("commit %v", c.commits)
	}
}
