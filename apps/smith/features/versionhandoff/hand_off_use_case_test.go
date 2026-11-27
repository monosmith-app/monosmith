package versionhandoff

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/directlink"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/smithreleases"
)

// handOffLab is a smith home with installed versions, a workspace, and a host
// whose Install and Start are recorded, never real: no program is started.
type handOffLab struct {
	home, workspace string
	env             map[string]string
	installed       []string
	started         []string
	startedEnv      []string
	startedArgs     []string
	stderr          bytes.Buffer
}

func newHandOffLab(t *testing.T, versions ...string) *handOffLab {
	t.Helper()
	l := &handOffLab{home: t.TempDir(), workspace: t.TempDir(), env: map[string]string{}}
	for _, v := range versions {
		write(t, smithreleases.ProgramPath(l.home, v, "linux"), "program "+v)
	}
	return l
}

func (l *handOffLab) host(exe, version string) Host {
	return Host{
		Getenv:  func(k string) string { return l.env[k] },
		Environ: []string{"PATH=/bin", "SMITH_LAUNCHER=", "HOME=/home/a"},
		Cwd:     l.workspace, Home: l.home, Exe: exe, Version: version, GOOS: "linux", GOARCH: "amd64",
		Install: func(_ context.Context, home, v, goos, _ string) (string, error) {
			l.installed = append(l.installed, v)
			return smithreleases.ProgramPath(home, v, goos), nil
		},
		Start: func(program string, args, env []string) (int, error) {
			l.started = append(l.started, program)
			l.startedArgs, l.startedEnv = args, env
			return 7, nil
		},
		Stderr: &l.stderr,
	}
}

func (l *handOffLab) link(version string) string {
	return filepath.Join(directlink.RunDir(l.home), "smith-"+version)
}

func TestProgramChoosesTheVersionLikeTheLaunchers(t *testing.T) {
	cases := []struct {
		name, running, smithVersion, pin, launcher string
		fromLink                                   bool
		want, install                              string
	}{
		{name: "the newest installed is the running one", running: "2.51.0", fromLink: true},
		{name: "an older link hands off to the newest installed", running: "2.50.0", fromLink: true, want: "2.51.0"},
		{name: "the workspace pin wins over the newest", running: "2.51.0", pin: "2.50.0", fromLink: true, want: "2.50.0"},
		{name: "SMITH_VERSION wins over the pin", running: "2.51.0", pin: "2.50.0", smithVersion: "2.49.0", fromLink: true, want: "2.49.0"},
		{name: "SMITH_VERSION latest ignores the pin", running: "2.50.0", pin: "2.50.0", smithVersion: "latest", fromLink: true, want: "2.51.0"},
		{name: "a missing pinned version is installed", running: "2.51.0", pin: "2.30.0", fromLink: true, want: "2.30.0", install: "2.30.0"},
		{name: "a pin of the running version runs here", running: "2.50.0", pin: "2.50.0", fromLink: true},
		{name: "a smith not started by the link runs here", running: "2.50.0", pin: "2.40.0"},
		{name: "a smith a launcher started runs here", running: "2.50.0", pin: "2.40.0", fromLink: true, launcher: "npm@3.0.0"},
		{name: "a smith the link handed off to runs here", running: "2.40.0", fromLink: true, launcher: "link@2.50.0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l := newHandOffLab(t, "2.49.0", "2.50.0", "2.51.0")
			if c.pin != "" {
				write(t, filepath.Join(l.workspace, "smith.toml"), "[workspace]\nsmith = \""+c.pin+"\"\n")
			}
			l.env["SMITH_VERSION"], l.env["SMITH_LAUNCHER"] = c.smithVersion, c.launcher
			exe := smithreleases.ProgramPath(l.home, c.running, "linux")
			if c.fromLink {
				exe = l.link(c.running)
			}
			got, err := Program(context.Background(), l.host(exe, c.running))
			want := ""
			if c.want != "" {
				want = smithreleases.ProgramPath(l.home, c.want, "linux")
			}
			if err != nil || got != want {
				t.Fatalf("Program = %q, %v; want %q", got, err, want)
			}
			if c.install == "" && len(l.installed) > 0 || c.install != "" && !slices.Equal(l.installed, []string{c.install}) {
				t.Errorf("installed %v, want %q", l.installed, c.install)
			}
			if c.install != "" && !strings.Contains(l.stderr.String(), "downloading smith "+c.install) {
				t.Errorf("no progress line for the download: %q", l.stderr.String())
			}
		})
	}
}

func TestHandOffStartsTheChosenProgramOnceWithTheLinkAsLauncher(t *testing.T) {
	l := newHandOffLab(t, "2.50.0", "2.51.0")
	code, handed := HandOff(context.Background(), l.host(l.link("2.50.0"), "2.50.0"), []string{"run", "test", "--yes"})
	if !handed || code != 7 {
		t.Fatalf("HandOff = %d, %v", code, handed)
	}
	if want := smithreleases.ProgramPath(l.home, "2.51.0", "linux"); !slices.Equal(l.started, []string{want}) {
		t.Errorf("started %v, want %s", l.started, want)
	}
	if !slices.Equal(l.startedArgs, []string{"run", "test", "--yes"}) {
		t.Errorf("arguments %v", l.startedArgs)
	}
	launchers := 0
	for _, kv := range l.startedEnv {
		if strings.HasPrefix(kv, "SMITH_LAUNCHER=") {
			launchers++
			if kv != "SMITH_LAUNCHER=link@2.50.0" {
				t.Errorf("the next program gets %s", kv)
			}
		}
	}
	if launchers != 1 || !slices.Contains(l.startedEnv, "PATH=/bin") {
		t.Errorf("environment %v", l.startedEnv)
	}

	// The program it started sees the variable and runs the command itself: no loop.
	l.env["SMITH_LAUNCHER"] = "link@2.50.0"
	if _, handed := HandOff(context.Background(), l.host(l.link("2.50.0"), "2.50.0"), nil); handed || len(l.started) != 1 {
		t.Error("a handed-off run hands off again")
	}
}

func TestHandOffSaysWhatWentWrong(t *testing.T) {
	l := newHandOffLab(t, "2.50.0")
	l.env["SMITH_VERSION"] = "two"
	if code, handed := HandOff(context.Background(), l.host(l.link("2.50.0"), "2.50.0"), nil); !handed || code != 1 || !strings.Contains(l.stderr.String(), `SMITH_VERSION is "two"`) {
		t.Errorf("a bad SMITH_VERSION: %d, %v, %q", code, handed, l.stderr.String())
	}

	l = newHandOffLab(t, "2.50.0")
	l.env["SMITH_VERSION"] = "2.30.0"
	h := l.host(l.link("2.50.0"), "2.50.0")
	h.Install = func(context.Context, string, string, string, string) (string, error) {
		return "", errors.New("smith 2.30.0 was not installed: checksums.txt does not list it")
	}
	if code, handed := HandOff(context.Background(), h, nil); !handed || code != 1 || len(l.started) != 0 || !strings.Contains(l.stderr.String(), "was not installed") {
		t.Errorf("a failed download: %d, %v, %q", code, handed, l.stderr.String())
	}

	l = newHandOffLab(t, "2.50.0", "2.51.0")
	h = l.host(l.link("2.50.0"), "2.50.0")
	h.Start = func(string, []string, []string) (int, error) { return 1, errors.New("exec format error") }
	if code, handed := HandOff(context.Background(), h, nil); !handed || code != 1 || !strings.Contains(l.stderr.String(), "smith link --undo") {
		t.Errorf("a program that cannot start: %d, %v, %q", code, handed, l.stderr.String())
	}
}
