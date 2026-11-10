package cachekey

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestTheProfileFileIsThatOfTheShell(t *testing.T) {
	for _, tc := range []struct {
		name, goos, parent, want string
	}{
		{"zsh", "darwin", "zsh", ".zshrc"},
		{"bash on macOS", "darwin", "bash", ".bash_profile"},
		{"bash on Linux", "linux", "bash", ".bashrc"},
		{"fish", "linux", "fish", filepath.Join(".config", "fish", "conf.d", "smith-cache.fish")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := newTempEnv(t, tc.goos, tc.parent)
			got, ok := TargetFor(env)
			if !ok || got.Path != filepath.Join(env.home, tc.want) {
				t.Fatalf("%+v %v, want %s", got, ok, tc.want)
			}
		})
	}
	if _, ok := TargetFor(newTempEnv(t, "linux", "nu")); ok {
		t.Fatal("an unknown shell got a file")
	}
}

func TestTheLineUsesTheSyntaxOfTheShell(t *testing.T) {
	zsh, _ := TargetFor(newTempEnv(t, "linux", "zsh"))
	fish, _ := TargetFor(newTempEnv(t, "linux", "fish"))
	if zsh.Line(KeyVariable, "x") != `export SMITH_CACHE_KEY="x"` || fish.Line(KeyVariable, "x") != `set -gx SMITH_CACHE_KEY "x"` {
		t.Fatalf("%q %q", zsh.Line(KeyVariable, "x"), fish.Line(KeyVariable, "x"))
	}
}

func TestWriteReplaceAndRemoveTheBlock(t *testing.T) {
	for _, shell := range []string{"zsh", "bash", "fish"} {
		t.Run(shell, func(t *testing.T) {
			env := newTempEnv(t, "linux", shell)
			target, _ := TargetFor(env)
			original := "export A=1\n"
			if shell == "fish" {
				original = ""
			} else {
				env.put(t, original, 0o640, strings.TrimPrefix(target.Path, env.home+string(filepath.Separator)))
			}
			if err := WriteProfileKey(env, target, KeyVariable, testKey); err != nil {
				t.Fatal(err)
			}
			if err := WriteProfileKey(env, target, KeyVariable, "second"+testKey); err != nil {
				t.Fatal(err)
			}
			got := env.read(target.Path)
			if strings.Count(got, markersFor(KeyVariable).Begin) != 1 || strings.Contains(got, target.Line(KeyVariable, testKey)) || !strings.Contains(got, target.Line(KeyVariable, "second"+testKey)) {
				t.Fatalf("block not replaced once:\n%s", got)
			}
			if !strings.HasPrefix(got, original) {
				t.Fatalf("the file of the person changed:\n%s", got)
			}
			changed, err := RemoveProfileKeys(env, KeyVariable)
			if err != nil || len(changed) != 1 || changed[0] != target.Path {
				t.Fatalf("%v %v", changed, err)
			}
			if shell == "fish" {
				if env.Exists(target.Path) {
					t.Fatal("the empty file is still there")
				}
			} else if env.read(target.Path) != original {
				t.Fatalf("profile not restored: %q", env.read(target.Path))
			}
		})
	}
}

func TestModeIsKeptAndANewFileIsPrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not meaningful on Windows")
	}
	env := newTempEnv(t, "linux", "zsh")
	target, _ := TargetFor(env)
	if err := WriteProfileKey(env, target, KeyVariable, testKey); err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Stat(target.Path); info.Mode().Perm() != 0o600 {
		t.Fatalf("new file mode %v", info.Mode().Perm())
	}
	env = newTempEnv(t, "linux", "bash")
	target, _ = TargetFor(env)
	path := env.put(t, "x\n", 0o644, ".bashrc")
	if err := WriteProfileKey(env, target, KeyVariable, testKey); err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o644 {
		t.Fatalf("existing file mode %v", info.Mode().Perm())
	}
}

func TestASecretWithShellCharactersIsQuotedSafely(t *testing.T) {
	zsh, _ := TargetFor(newTempEnv(t, "linux", "zsh"))
	fish, _ := TargetFor(newTempEnv(t, "linux", "fish"))
	for _, tc := range []struct {
		target    ProfileTarget
		value, in string
	}{
		{zsh, "a&b%3D:c", `export V="a&b%3D:c"`},
		{zsh, `a'b$c"d`, `export V='a'\''b$c"d'`},
		{fish, "a&b", `set -gx V "a&b"`},
		{fish, `a'b\c`, `set -gx V 'a\'b\\c'`},
	} {
		if got := tc.target.Line("V", tc.value); got != tc.in {
			t.Errorf("%s: got %s, want %s", tc.target.Shell, got, tc.in)
		}
	}
}

func TestASecretWithALineEndIsRefusedWithoutEchoingIt(t *testing.T) {
	env := newTempEnv(t, "linux", "zsh")
	target, _ := TargetFor(env)
	bad := "abc\nrm -rf ~"
	err := WriteProfileKey(env, target, KeyVariable, bad)
	if err == nil || strings.Contains(err.Error(), bad) || env.Exists(target.Path) {
		t.Fatalf("%v", err)
	}
}

func TestRemoveWithNothingToRemoveSaysSo(t *testing.T) {
	out := &bytes.Buffer{}
	if err := RemoveKey(UserEnvironment{GOOS: "linux", Machine: newTempEnv(t, "linux", "zsh")}, out); err != nil || !strings.Contains(out.String(), "Nothing changed") {
		t.Fatalf("%v %q", err, out)
	}
}

func TestRemoveOnWindowsUnsetsTheUserVariable(t *testing.T) {
	var ran []string
	u := UserEnvironment{GOOS: "windows", Run: func(_ string, argv ...string) (string, error) { ran = argv; return "", nil }}
	out := &bytes.Buffer{}
	if err := RemoveKey(u, out); err != nil || len(ran) == 0 || !strings.Contains(ran[len(ran)-1], "$null, 'User'") || !strings.Contains(out.String(), "no longer set") {
		t.Fatalf("%v %v %q", err, ran, out)
	}
}

// choose runs the key screen on a zsh machine, presses u and then the keys,
// and returns everything the session printed and the profile file.
func choose(t *testing.T, keys ...rune) (out, profile string) {
	t.Helper()
	env := newTempEnv(t, "linux", "zsh")
	clip, buf := &fakeClipboard{readable: true}, &bytes.Buffer{}
	s := newSession(clip, keysOf(append(append([]rune{'u'}, keys...), 'q')...), buf)
	s.Env = UserEnvironment{GOOS: "linux", Machine: env}
	s.Produce = func() (string, error) { return testKey, nil }
	if err := s.Run(); err != nil {
		t.Fatal(err)
	}
	target, _ := TargetFor(env)
	profile = env.read(target.Path)
	if strings.Contains(buf.String(), testKey) {
		t.Fatalf("the key is in the output:\n%s", buf)
	}
	return buf.String(), profile
}

func TestTheConfirmationShowsTheMaskedLinesAndAWarning(t *testing.T) {
	out, profile := choose(t, 27)
	for _, want := range []string{".zshrc", `export SMITH_CACHE_KEY="••••"`, "dotfile git repositories keep the secret", "Enter writes it", "Not written"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	if profile != "" {
		t.Fatalf("a cancelled question wrote the file: %q", profile)
	}
}

func TestOnlyEnterWritesTheFileAndTheKeyIsNotShown(t *testing.T) {
	env := newTempEnv(t, "linux", "zsh")
	clip, buf := &fakeClipboard{readable: true}, &bytes.Buffer{}
	s := newSession(clip, keysOf('u', 'x', ' ', '\r', 'q'), buf)
	s.Env = UserEnvironment{GOOS: "linux", Machine: env}
	s.Produce = func() (string, error) { return testKey, nil }
	if err := s.Run(); err != nil {
		t.Fatal(err)
	}
	target, _ := TargetFor(env)
	if !strings.Contains(env.read(target.Path), target.Line(KeyVariable, testKey)) {
		t.Fatalf("the key is not in the profile:\n%s", env.read(target.Path))
	}
	for _, want := range []string{"smith wrote SMITH_CACHE_KEY to", "new terminal", "smith cache key --remove"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q in\n%s", want, buf)
		}
	}
	if strings.Contains(buf.String(), testKey) {
		t.Fatalf("the key is in the output:\n%s", buf)
	}
}

func TestShowHowByHandWritesNothing(t *testing.T) {
	out, profile := choose(t, 'h')
	if profile != "" || !strings.Contains(out, `export SMITH_CACHE_KEY="<the key>"`) || !strings.Contains(out, "smith changed nothing") {
		t.Fatalf("profile %q\n%s", profile, out)
	}
}

func TestAWriteErrorNeverHoldsTheKey(t *testing.T) {
	env := newTempEnv(t, "linux", "zsh")
	target, _ := TargetFor(env)
	// A folder where the file should be makes the write fail.
	if err := os.MkdirAll(target.Path, 0o755); err != nil {
		t.Fatal(err)
	}
	err := WriteProfileKey(env, target, KeyVariable, testKey)
	if err == nil || strings.Contains(err.Error(), testKey) {
		t.Fatalf("%v", err)
	}
}

func TestAWindowsSetErrorNeverHoldsTheKey(t *testing.T) {
	u := UserEnvironment{GOOS: "windows", Run: func(stdin string, _ ...string) (string, error) {
		return "", errors.New("failed on " + strings.TrimSpace(stdin))
	}}
	if err := u.Set(KeyVariable, testKey); err == nil || strings.Contains(err.Error(), testKey) {
		t.Fatalf("%v", err)
	}
}

func TestTheSASTokenGetsItsOwnVariableAndBlockAndRemovalUndoesBoth(t *testing.T) {
	env := newTempEnv(t, "linux", "zsh")
	target, _ := TargetFor(env)
	const token = "sv=2022-11-02&sig=abc"
	if err := WriteProfileKey(env, target, KeyVariable, testKey); err != nil {
		t.Fatal(err)
	}
	if err := WriteProfileKey(env, target, SASVariable, token); err != nil {
		t.Fatal(err)
	}
	got := env.read(target.Path)
	if !strings.Contains(got, `export AZURE_STORAGE_SAS_TOKEN="sv=2022-11-02&sig=abc"`) || !strings.Contains(got, "# >>> smith azure_storage_sas_token >>>") || !strings.Contains(got, target.Line(KeyVariable, testKey)) {
		t.Fatalf("both blocks must be there:\n%s", got)
	}
	out := &bytes.Buffer{}
	if err := RemoveKey(UserEnvironment{GOOS: "linux", Machine: env}, out); err != nil {
		t.Fatal(err)
	}
	if env.Exists(target.Path) || !strings.Contains(out.String(), "Removed SMITH_CACHE_KEY") || !strings.Contains(out.String(), "Removed AZURE_STORAGE_SAS_TOKEN") || strings.Contains(out.String(), testKey) {
		t.Fatalf("%q", out)
	}
}

func TestTheSASScreenNamesItsVariableAndKeepsTheAcronym(t *testing.T) {
	env := newTempEnv(t, "linux", "zsh")
	buf := &bytes.Buffer{}
	s := newSession(&fakeClipboard{readable: true}, keysOf('u', '\r', 'q'), buf)
	s.Env = UserEnvironment{GOOS: "linux", Machine: env}
	s.Noun, s.Variable = "SAS token", SASVariable
	s.Produce = func() (string, error) { return "sv-2022", nil }
	if err := s.Run(); err != nil {
		t.Fatal(err)
	}
	target, _ := TargetFor(env)
	out := buf.String()
	for _, want := range []string{"SAS token copied to the clipboard (not shown). smith reads it from AZURE_STORAGE_SAS_TOKEN", "if it still holds the SAS token", `export AZURE_STORAGE_SAS_TOKEN="••••"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	if strings.Contains(out, "sv-2022") || !strings.Contains(env.read(target.Path), `export AZURE_STORAGE_SAS_TOKEN="sv-2022"`) {
		t.Fatalf("output %s\nprofile %s", out, env.read(target.Path))
	}
}

func TestTheWindowsUnsetNamesBothVariables(t *testing.T) {
	var ran []string
	u := UserEnvironment{GOOS: "windows", Run: func(_ string, argv ...string) (string, error) {
		ran = append(ran, argv[len(argv)-1])
		return "", nil
	}}
	if err := RemoveKey(u, &bytes.Buffer{}); err != nil || len(ran) != 2 || !strings.Contains(ran[1], "AZURE_STORAGE_SAS_TOKEN") {
		t.Fatalf("%v %v", err, ran)
	}
}
