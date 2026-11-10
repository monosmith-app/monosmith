package completioncommand

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/userprefs"
)

func TestTheAnswerIsKeptInThePreferencesFileOfTheSmithHome(t *testing.T) {
	home := filepath.Join(t.TempDir(), "smith-home")
	if err := userprefs.Save(home, userprefs.Preferences{Theme: "light"}); err != nil {
		t.Fatal(err)
	}
	m := PreferencesMemory{Home: home}
	if got := m.Load(); got.Choice != "" {
		t.Fatalf("%+v", got)
	}
	when := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	if err := m.Save(Answer{Choice: ChoiceLater, AskAfter: when}); err != nil {
		t.Fatal(err)
	}
	got := m.Load()
	if got.Choice != ChoiceLater || !got.AskAfter.Equal(when) {
		t.Fatalf("%+v", got)
	}
	if p, _ := userprefs.Load(home); p.Theme != "light" {
		t.Fatalf("saving the answer lost another preference: %+v", p)
	}
	raw, _ := os.ReadFile(userprefs.Path(home))
	if !strings.Contains(string(raw), `"completionOffer": "later"`) {
		t.Fatalf("the file does not name the answer:\n%s", raw)
	}
}
