package directlink

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lab is a smith installed by a package manager, with the entries it put on PATH.
type lab struct {
	home, bin, exe string
}

func newLab(t *testing.T) lab {
	t.Helper()
	root := t.TempDir()
	l := lab{home: filepath.Join(root, ".smith"), bin: filepath.Join(root, "npm"), exe: filepath.Join(root, "npm", "node_modules", "monosmith", "native", "smith.exe")}
	put(t, l.exe, "the native program")
	return l
}

func put(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o755); err != nil {
		t.Fatal(err)
	}
}

func (l lab) options(goos string) Options {
	return Options{Home: l.home, Exe: l.exe, Version: "2.41.0", GOOS: goos}
}

const (
	npmCmd = "@ECHO off\r\n\"%_prog%\"  \"%dp0%\\node_modules\\monosmith\\bin\\smith.js\" %*\r\n"
	npmPs1 = "#!/usr/bin/env pwsh\r\n& \"node\"  \"$basedir/node_modules/monosmith/bin/smith.js\" $args\r\n"
	npmSh  = "#!/bin/sh\nexec node \"$basedir/node_modules/monosmith/bin/smith.js\" \"$@\"\n"
)

func read(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestFindEntriesClassifiesTheEntriesOfTheInstallation(t *testing.T) {
	l := newLab(t)
	put(t, filepath.Join(l.bin, "smith.cmd"), npmCmd)
	put(t, filepath.Join(l.bin, "smith.ps1"), npmPs1)
	put(t, filepath.Join(l.bin, "smith"), npmSh)
	got := FindEntries(l.exe, "windows", RunDir(l.home), "")
	kinds := map[string]Kind{}
	for _, e := range got {
		kinds[filepath.Base(e.Path)] = e.Kind
	}
	want := map[string]Kind{"smith.cmd": CmdShim, "smith.ps1": Ps1Shim, "smith": ShShim}
	if len(kinds) != 3 || kinds["smith.cmd"] != want["smith.cmd"] || kinds["smith.ps1"] != want["smith.ps1"] || kinds["smith"] != want["smith"] {
		t.Errorf("entries = %v, want %v", kinds, want)
	}
}

func TestLinkPointsEveryShimAtAHardLinkOfTheProgramAndUndoGivesThemBack(t *testing.T) {
	l := newLab(t)
	put(t, filepath.Join(l.bin, "smith.cmd"), npmCmd)
	put(t, filepath.Join(l.bin, "smith.ps1"), npmPs1)
	put(t, filepath.Join(l.bin, "smith"), npmSh)
	o := l.options("windows")

	report, err := Link(o)
	if err != nil || len(report.Changed) != 3 || len(report.Skipped) != 0 {
		t.Fatalf("Link = %+v, %v", report, err)
	}
	target := filepath.Join(RunDir(l.home), "smith-2.41.0.exe")
	if report.Target != target || !sameFile(target, l.exe) {
		t.Fatalf("target = %q, is it a link of the installed program? %v", report.Target, sameFile(target, l.exe))
	}
	if got := read(t, filepath.Join(l.bin, "smith.cmd")); got != shimText(CmdShim, target) || strings.Contains(got, "node") || !strings.Contains(got, "goto #_undefined_#") {
		t.Errorf("smith.cmd = %q", got)
	}
	if got := read(t, filepath.Join(l.bin, "smith")); !strings.Contains(got, "exec \"") || strings.Contains(got, "node") {
		t.Errorf("smith (sh) = %q", got)
	}

	// The installed file is free: it can be deleted and written again while the link runs.
	if err := os.Remove(l.exe); err != nil {
		t.Fatal(err)
	}
	put(t, l.exe, "the next version")
	if read(t, target) != "the native program" {
		t.Error("the link did not keep the old program")
	}

	again, err := Link(o)
	if err != nil || len(again.Changed) != 0 || len(again.Already) != 3 {
		t.Fatalf("a second Link changed things: %+v, %v", again, err)
	}

	restored, skipped, err := Undo(o)
	if err != nil || len(restored) != 3 || len(skipped) != 0 {
		t.Fatalf("Undo = %v, %v, %v", restored, skipped, err)
	}
	if read(t, filepath.Join(l.bin, "smith.cmd")) != npmCmd || read(t, filepath.Join(l.bin, "smith.ps1")) != npmPs1 || read(t, filepath.Join(l.bin, "smith")) != npmSh {
		t.Error("an entry was not given back as it was")
	}
	if _, ok := LoadState(l.home); ok {
		t.Error("the link is remembered after Undo")
	}
}

func TestLinkOnAnotherSystemCallsTheInstalledProgramAndMakesNoLink(t *testing.T) {
	l := newLab(t)
	put(t, filepath.Join(l.bin, "smith"), npmSh)
	report, err := Link(l.options("linux"))
	if err != nil || report.Target != l.exe || len(report.Changed) != 1 {
		t.Fatalf("Link = %+v, %v", report, err)
	}
	if _, err := os.Stat(filepath.Join(RunDir(l.home), "smith-2.41.0.exe")); err == nil {
		t.Error("a link was made where links are not needed")
	}
	if !strings.Contains(read(t, filepath.Join(l.bin, "smith")), filepath.ToSlash(l.exe)) {
		t.Error("the entry does not call the installed program")
	}
}

func TestLinkRepointsASymlinkAndUndoPointsItBack(t *testing.T) {
	l := newLab(t)
	script := filepath.Join(l.bin, "node_modules", "monosmith", "bin", "smith.js")
	put(t, script, "#!/usr/bin/env node\n")
	entry := filepath.Join(l.bin, "smith")
	if err := os.Symlink(script, entry); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	o := l.options("linux")
	if report, err := Link(o); err != nil || len(report.Changed) != 1 {
		t.Fatalf("Link = %+v, %v", report, err)
	}
	if got, _ := os.Readlink(entry); got != l.exe {
		t.Errorf("the symlink points at %q", got)
	}
	if _, _, err := Undo(o); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.Readlink(entry); got != script {
		t.Errorf("the symlink was not pointed back: %q", got)
	}
}

func TestLinkReplacesALauncherProgramAndKeepsItAside(t *testing.T) {
	l := newLab(t)
	launcher := filepath.Join(l.bin, "smith.exe")
	put(t, launcher, "MZ launcher of monosmith")
	o := l.options("windows")
	report, err := Link(o)
	if err != nil || len(report.Changed) != 1 || report.Changed[0].Kind != Launcher {
		t.Fatalf("Link = %+v, %v", report, err)
	}
	if !sameFile(launcher, l.exe) {
		t.Error("the launcher is not a link of the program")
	}
	if read(t, launcher+".smith-old") != "MZ launcher of monosmith" {
		t.Error("the launcher was not kept aside")
	}
	if _, _, err := Undo(o); err != nil {
		t.Fatal(err)
	}
	if read(t, launcher) != "MZ launcher of monosmith" {
		t.Error("the launcher was not given back")
	}
}

func TestNeededIsTrueUntilTheVersionIsLinkedAndNeverFromALink(t *testing.T) {
	l := newLab(t)
	o := l.options("windows")
	if Needed(o) {
		t.Fatal("needed with no entry on PATH")
	}
	put(t, filepath.Join(l.bin, "smith.cmd"), npmCmd)
	if !Needed(o) {
		t.Fatal("not needed before the first link")
	}
	report, err := Link(o)
	if err != nil {
		t.Fatal(err)
	}
	if Needed(o) {
		t.Error("needed after the version was linked")
	}
	fromLink := o
	fromLink.Exe = report.Target
	if Needed(fromLink) || !RunningFromLink(l.home, report.Target) {
		t.Error("a smith that runs from its link wants to link again")
	}
	if got := InstalledExe(l.home, report.Target); got != l.exe {
		t.Errorf("the installed program is %q, want %q (update reads the install source from it)", got, l.exe)
	}
	next := o
	next.Version = "2.42.0"
	if !Needed(next) {
		t.Error("a new version does not link again")
	}
}

func TestCheckSaysWhichEntriesStartDirectly(t *testing.T) {
	l := newLab(t)
	put(t, filepath.Join(l.bin, "smith.cmd"), npmCmd)
	o := l.options("windows")
	if st := Check(o); len(st.Entries) != 1 || st.Entries[0].Kind != CmdShim || st.Direct || st.Linked != "" {
		t.Fatalf("before: %+v", st)
	}
	if _, err := Link(o); err != nil {
		t.Fatal(err)
	}
	if st := Check(o); len(st.Entries) != 1 || st.Entries[0].Kind != Direct || st.Linked != "2.41.0" {
		t.Errorf("after: %+v", st)
	}
}

func TestAnotherInstallationOfSmithIsNeverTouched(t *testing.T) {
	l := newLab(t)
	put(t, filepath.Join(l.bin, "smith.cmd"), npmCmd)
	// A second smith with an install of its own: its entries name monosmith too, and one
	// is a whole smith program, as a pip Scripts folder holds.
	other := filepath.Join(t.TempDir(), "other-npm")
	put(t, filepath.Join(other, "node_modules", "monosmith", "native", "smith.exe"), "other program")
	put(t, filepath.Join(other, "smith.cmd"), npmCmd)
	put(t, filepath.Join(other, "smith.exe"), "MZ"+strings.Repeat("monosmith ", 200000))
	put(t, filepath.Join(l.bin, "smith.exe"), "MZ"+strings.Repeat("monosmith ", 200000))

	report, err := Link(l.options("windows"))
	if err != nil || len(report.Changed) != 1 || filepath.Base(report.Changed[0].Path) != "smith.cmd" {
		t.Fatalf("Link = %+v, %v (only the entry of this installation may change; a whole program is no launcher)", report, err)
	}
	if read(t, filepath.Join(other, "smith.cmd")) != npmCmd {
		t.Error("an entry of another installation was changed")
	}
	if !strings.HasPrefix(read(t, filepath.Join(other, "smith.exe")), "MZ") {
		t.Error("the program of another installation was replaced")
	}
}

func TestOwnedDirsFollowTheLayoutOfThePackageManager(t *testing.T) {
	cases := map[string][]string{
		filepath.FromSlash("/home/a/.npm-global/lib/node_modules/monosmith/node_modules/@monosmith/cli-linux-x64/bin/smith"):            {filepath.FromSlash("/home/a/.npm-global/lib"), filepath.FromSlash("/home/a/.npm-global/bin")},
		filepath.FromSlash("C:/Users/a/AppData/Roaming/npm/node_modules/monosmith/node_modules/@monosmith/cli-win32-x64/bin/smith.exe"): {filepath.FromSlash("C:/Users/a/AppData/Roaming/npm")},
	}
	for exe, want := range cases {
		got := ownedDirs(exe)
		for _, w := range want {
			found := false
			for _, g := range got {
				found = found || g == w
			}
			if !found {
				t.Errorf("ownedDirs(%q) lacks %q: %v", exe, w, got)
			}
		}
	}
	if got := ownedDirs(filepath.FromSlash("/opt/smith/bin/smith")); len(got) != 0 {
		t.Errorf("a program outside a package manager owns entries: %v", got)
	}
}
