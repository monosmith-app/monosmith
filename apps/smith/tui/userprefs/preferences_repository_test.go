package userprefs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPreferencesRoundTripThroughTheSmithHome(t *testing.T) {
	home := filepath.Join(t.TempDir(), "not-yet")
	if p, err := Load(home); err != nil || !reflect.DeepEqual(p, Preferences{}) {
		t.Fatalf("no file is no preferences: %+v, %v", p, err)
	}
	want := Preferences{Theme: "light", Icons: "ascii", SettingsTipSeen: true}
	if err := Save(home, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(home)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("%+v, %v", got, err)
	}
	if entries, _ := os.ReadDir(home); len(entries) != 1 {
		t.Fatalf("a temporary file was left behind: %v", entries)
	}
}

func TestAFileThatIsNotJSONGivesDefaultsAndTheReason(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(Path(home), []byte("{nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := Load(home)
	if err == nil || !reflect.DeepEqual(p, Preferences{}) {
		t.Fatalf("%+v, %v", p, err)
	}
}

func TestWithoutAHomeNothingIsKept(t *testing.T) {
	if p, err := Load(""); err != nil || !reflect.DeepEqual(p, Preferences{}) {
		t.Fatal(p, err)
	}
	if Save("", Preferences{}) == nil {
		t.Fatal("saving with no home must say so")
	}
}
