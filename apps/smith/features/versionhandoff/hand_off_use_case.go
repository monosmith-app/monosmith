package versionhandoff

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/directlink"
	"github.com/monosmith-app/monosmith/apps/smith/features/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/smithreleases"
)

// launcherVariable tells smith what started it; the launchers set it, and so
// does the hand-off, which also stops a second hand-off in the program it starts.
const launcherVariable = "SMITH_LAUNCHER"

// Host is what a hand-off reads and does. ForThisProcess fills it for the
// running smith; tests give fakes.
type Host struct {
	// Getenv reads the environment.
	Getenv func(string) string
	// Environ is the environment the next program gets, as KEY=value.
	Environ []string
	// Cwd is the current folder, where the workspace pin is looked for.
	Cwd string
	// Home is the smith home.
	Home string
	// Exe is the running program.
	Exe string
	// Version is the running smith.
	Version string
	// GOOS and GOARCH are the system smith runs on.
	GOOS, GOARCH string
	// Install downloads, checks and installs one version, and returns its program.
	Install func(ctx context.Context, home, version, goos, goarch string) (string, error)
	// Start runs a program with the arguments and environment, on this terminal,
	// and returns its exit code. On Unix it replaces this process and returns
	// only when it could not.
	Start func(program string, args, env []string) (int, error)
	// Stderr is where progress and problems go.
	Stderr io.Writer
}

// Program is the program a run must hand off to, or "" when the running smith
// is the one to run: it was not started by the link, a launcher started it, no
// version is installed, or the chosen version is the running one. A pinned
// version that is missing is downloaded and checked first, with a line on
// Stderr; this is the only case that uses the network.
//
// Parameters: ctx bounds the download; h is the host.
// Returns: the program, or ""; an error for a bad SMITH_VERSION or a pinned
// version that cannot be installed.
func Program(ctx context.Context, h Host) (string, error) {
	if h.Getenv(launcherVariable) != "" || !directlink.RunningFromLink(h.Home, h.Exe) {
		return "", nil
	}
	wanted, err := ChooseVersion(h.Getenv("SMITH_VERSION"), func() string { return WorkspacePin(h.Cwd) })
	if err != nil {
		return "", err
	}
	pinned := wanted != ""
	if !pinned {
		wanted = NewestInstalled(h.Home, h.GOOS)
	}
	if wanted == "" || wanted == h.Version {
		return "", nil
	}
	program := smithreleases.ProgramPath(h.Home, wanted, h.GOOS)
	if _, err := os.Stat(program); err == nil || !pinned {
		return program, nil
	}
	fmt.Fprintf(h.Stderr, "smith: downloading smith %s, the version this run asks for\n", wanted)
	return h.Install(ctx, h.Home, wanted, h.GOOS, h.GOARCH)
}

// HandOff runs the version the workspace pins, or the newest installed, when the
// link started a different one. The next program gets SMITH_LAUNCHER=link@<this
// version>, so it does not hand off again and smith update installs into the
// versions folder.
//
// Parameters: ctx bounds a download; h is the host; args are the arguments of
// the run.
// Returns: the exit code of the run, and true when the run was handed off (or
// failed before it could be); false means the running smith runs the command.
func HandOff(ctx context.Context, h Host, args []string) (int, bool) {
	program, err := Program(ctx, h)
	if err != nil {
		fmt.Fprintln(h.Stderr, "smith: "+err.Error())
		return 1, true
	}
	if program == "" {
		return 0, false
	}
	env := make([]string, 0, len(h.Environ)+1)
	for _, kv := range h.Environ {
		if !strings.HasPrefix(strings.ToUpper(kv), launcherVariable+"=") {
			env = append(env, kv)
		}
	}
	env = append(env, launcherVariable+"="+directlink.LauncherPrefix+h.Version)
	code, err := h.Start(program, args, env)
	if err != nil {
		fmt.Fprintf(h.Stderr, "smith: could not start %s: %v\n", program, err)
		fmt.Fprintln(h.Stderr, "smith: run smith link --undo to start smith through its launcher again")
		return 1, true
	}
	return code, true
}

// ForThisProcess is the host of the running smith.
//
// Parameters: version is the running smith.
// Returns: the host.
func ForThisProcess(version string) Host {
	o := directlink.ForThisProcess(version)
	cwd, _ := os.Getwd()
	return Host{
		Getenv: os.Getenv, Environ: os.Environ(), Cwd: cwd,
		Home: plugintrust.Home(), Exe: o.Exe, Version: version,
		GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		Install: func(ctx context.Context, home, v, goos, goarch string) (string, error) {
			return smithreleases.Install(ctx, home, v, goos, goarch, smithreleases.PublicKey)
		},
		Start:  startProgram,
		Stderr: os.Stderr,
	}
}
