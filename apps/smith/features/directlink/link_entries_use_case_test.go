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
	got := FindEntries(l.exe, "windows", RunDir(l.home), "", nil)
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

func TestLinkOnAnotherSystemCallsALinkInTheRunFolder(t *testing.T) {
	l := newLab(t)
	put(t, filepath.Join(l.bin, "smith"), npmSh)
	report, err := Link(l.options("linux"))
	target := filepath.Join(RunDir(l.home), "smith-2.41.0")
	if err != nil || report.Target != target || len(report.Changed) != 1 {
		t.Fatalf("Link = %+v, %v", report, err)
	}
	if !sameFile(target, l.exe) {
		t.Error("the run folder holds no link of the installed program")
	}
	if !strings.Contains(read(t, filepath.Join(l.bin, "smith")), filepath.ToSlash(target)) {
		t.Error("the entry does not call the link")
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
	if got, _ := os.Readlink(entry); got != filepath.Join(RunDir(l.home), "smith-2.41.0") {
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

func TestOwnedDirsKnowThePnpmLayoutAndANativeEntryNeedsNoLink(t *testing.T) {
	pnpm := filepath.FromSlash("C:/Users/a/AppData/Local/pnpm/store/v11/links/@monosmith/cli-win32-x64/2.45.0/abc/node_modules/@monosmith/cli-win32-x64/bin/smith.exe")
	got := ownedDirs(pnpm)
	home := filepath.FromSlash("C:/Users/a/AppData/Local/pnpm")
	want := []string{filepath.Join(home, "bin"), home}
	for _, w := range want {
		found := false
		for _, g := range got {
			found = found || g == w
		}
		if !found {
			t.Errorf("ownedDirs(pnpm) lacks %q: %v", w, got)
		}
	}
	if !isNativeEntry("", filepath.FromSlash("C:/Python/Scripts/smith.exe")) || !isNativeEntry("", filepath.FromSlash("/home/a/.local/bin/smith")) {
		t.Error("a program in a Scripts or bin folder is its own entry")
	}
	if isNativeEntry("", filepath.FromSlash("/home/a/.npm-global/lib/node_modules/monosmith/node_modules/@monosmith/cli-linux-x64/bin/smith")) {
		t.Error("an npm install has a launcher in front")
	}
	l := newLab(t)
	native := filepath.Join(l.bin, "Scripts", "smith.exe")
	put(t, native, "the whole program")
	o := Options{Home: l.home, Exe: native, Version: "2.41.0", GOOS: "windows"}
	if report, err := Link(o); err != nil || len(report.Changed) != 0 || len(report.Already) != 1 {
		t.Fatalf("Link of a native entry = %+v, %v", report, err)
	}
	if st := Check(o); !st.Direct || len(st.Entries) != 1 || st.Entries[0].Kind != Direct {
		t.Errorf("Check of a native entry = %+v", st)
	}
	if Needed(o) {
		t.Error("a native entry wants a link")
	}
}

// launcherLab is smith installed by the npm launcher package (ADR 0101): the program
// in <home>/versions/<version>/, and the npm entries in a folder on PATH that start
// node_modules/monosmith/src/main.mjs.
type launcherLab struct {
	home, npm, exe string
}

func newLauncherLab(t *testing.T, goos string) launcherLab {
	t.Helper()
	root := t.TempDir()
	name := "smith"
	if goos == "windows" {
		name = "smith.exe"
	}
	l := launcherLab{home: filepath.Join(root, ".smith"), npm: filepath.Join(root, "npm"), exe: filepath.Join(root, ".smith", "versions", "2.50.0", name)}
	put(t, l.exe, "the program of 2.50.0")
	return l
}

func (l launcherLab) options(goos string, path ...string) Options {
	return Options{Home: l.home, Exe: l.exe, Version: "2.50.0", GOOS: goos, PathDirs: path}
}

const (
	launcherCmd = "@ECHO off\r\n\"%_prog%\"  \"%dp0%\\node_modules\\monosmith\\src\\main.mjs\" %*\r\n"
	launcherPs1 = "#!/usr/bin/env pwsh\r\n& \"node\"  \"$basedir/node_modules/monosmith/src/main.mjs\" $args\r\n"
	launcherSh  = "#!/bin/sh\nexec node  \"$basedir/node_modules/monosmith/src/main.mjs\" \"$@\"\n"
)

func TestLinkFindsTheNpmLauncherEntriesOnPathOnWindows(t *testing.T) {
	l := newLauncherLab(t, "windows")
	put(t, filepath.Join(l.npm, "smith.cmd"), launcherCmd)
	put(t, filepath.Join(l.npm, "smith.ps1"), launcherPs1)
	put(t, filepath.Join(l.npm, "smith"), launcherSh)
	// Another program named smith on PATH, and a folder that is not there.
	other := filepath.Join(t.TempDir(), "tools")
	put(t, filepath.Join(other, "smith.cmd"), "@ECHO off\r\nother-tool %*\r\n")
	o := l.options("windows", other, l.npm, filepath.Join(l.npm, "missing"), "")

	if !Needed(o) {
		t.Fatal("a smith the npm launcher started does not want a link")
	}
	report, err := Link(o)
	target := filepath.Join(RunDir(l.home), "smith-2.50.0.exe")
	if err != nil || len(report.Changed) != 3 || report.Target != target {
		t.Fatalf("Link = %+v, %v", report, err)
	}
	if !sameFile(target, l.exe) {
		t.Error("the run folder holds no link of the installed version")
	}
	if got := read(t, filepath.Join(l.npm, "smith.cmd")); got != shimText(CmdShim, target) {
		t.Errorf("smith.cmd = %q", got)
	}
	if read(t, filepath.Join(other, "smith.cmd")) != "@ECHO off\r\nother-tool %*\r\n" {
		t.Error("another program named smith was changed")
	}
	if Needed(o) {
		t.Error("needed after the link")
	}
	fromLink := o
	fromLink.Exe = target
	if InstalledExe(l.home, target) != l.exe || !Managed(l.home, InstalledExe(l.home, target)) {
		t.Error("a smith that runs from the link does not know the launchers installed it")
	}
	if st := Check(fromLink); !st.Direct || len(st.Entries) != 3 {
		t.Errorf("Check from the link = %+v", st)
	}
	if _, _, err := Undo(o); err != nil {
		t.Fatal(err)
	}
	if read(t, filepath.Join(l.npm, "smith.cmd")) != launcherCmd || read(t, filepath.Join(l.npm, "smith")) != launcherSh {
		t.Error("undo did not give the npm entries back")
	}
}

func TestLinkRepointsTheNpmLauncherSymlinkOnUnixAndANewerVersionLinksAgain(t *testing.T) {
	l := newLauncherLab(t, "linux")
	main := filepath.Join(l.npm, "lib", "node_modules", "monosmith", "src", "main.mjs")
	put(t, main, "#!/usr/bin/env node\n")
	bin := filepath.Join(l.npm, "bin")
	entry := filepath.Join(bin, "smith")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(main, entry); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	o := l.options("linux", bin)
	if report, err := Link(o); err != nil || len(report.Changed) != 1 || report.Changed[0].Kind != Symlink {
		t.Fatalf("Link = %+v, %v", report, err)
	}
	first := filepath.Join(RunDir(l.home), "smith-2.50.0")
	if got, _ := os.Readlink(entry); got != first {
		t.Fatalf("the symlink points at %q, want %q", got, first)
	}

	// smith update installed 2.51.0, and the link handed a run to it: the entry still
	// calls the link of 2.50.0, which is smith's and is changed again.
	next := filepath.Join(l.home, "versions", "2.51.0", "smith")
	put(t, next, "the program of 2.51.0")
	n := Options{Home: l.home, Exe: next, Version: "2.51.0", GOOS: "linux", PathDirs: []string{bin}, Launcher: "link@2.50.0"}
	if st := Check(n); !st.Direct || len(st.Entries) != 1 || st.Entries[0].Kind != Symlink {
		t.Fatalf("Check of the next version = %+v", st)
	}
	if report, err := Link(n); err != nil || len(report.Changed) != 1 {
		t.Fatalf("Link of the next version = %+v, %v", report, err)
	}
	if got, _ := os.Readlink(entry); got != filepath.Join(RunDir(l.home), "smith-2.51.0") {
		t.Errorf("the symlink points at %q", got)
	}
	if _, _, err := Undo(n); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.Readlink(entry); got != main {
		t.Errorf("undo pointed the symlink at %q, want the npm launcher", got)
	}
}

func TestAProgramTheLaunchersInstalledIsNoNativeEntryAndLooksOnlyThenOnPath(t *testing.T) {
	l := newLauncherLab(t, "windows")
	if isNativeEntry(l.home, l.exe) {
		t.Error("a program in the versions folder is taken for its own entry")
	}
	put(t, filepath.Join(l.npm, "smith.cmd"), launcherCmd)
	dev := filepath.Join(t.TempDir(), "build", "smith.exe")
	put(t, dev, "a program built by hand")
	o := Options{Home: l.home, Exe: dev, Version: "dev", GOOS: "windows", PathDirs: []string{l.npm}}
	if report, err := Link(o); err != nil || len(report.Changed) != 0 {
		t.Fatalf("a program started by its path took the smith command over: %+v, %v", report, err)
	}
	if read(t, filepath.Join(l.npm, "smith.cmd")) != launcherCmd {
		t.Error("the npm entry was changed")
	}
}

func TestALauncherProgramOfALauncherPackageKeepsItsLauncher(t *testing.T) {
	l := newLauncherLab(t, "windows")
	scripts := filepath.Join(t.TempDir(), "Scripts")
	put(t, filepath.Join(scripts, "smith.exe"), "MZ launcher of monosmith_launcher")
	o := l.options("windows", scripts)
	report, err := Link(o)
	if err != nil || len(report.Changed) != 0 || len(report.Skipped) != 1 {
		t.Fatalf("Link = %+v, %v", report, err)
	}
	if read(t, filepath.Join(scripts, "smith.exe")) != "MZ launcher of monosmith_launcher" {
		t.Error("the launcher program was replaced")
	}
	if Needed(o) {
		t.Error("a launcher program that is kept asks for a link at every run")
	}
}
