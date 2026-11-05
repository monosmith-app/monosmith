package cachekey

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/clipboard"
)

// tool is one program that writes and reads the clipboard.
type tool struct {
	name  string
	write []string
	read  []string
}

// toolsFor lists, in order of preference, the clipboard programs of an
// operating system. getenv decides between Wayland and X on Linux.
func toolsFor(goos string, getenv func(string) string) []tool {
	switch goos {
	case "windows":
		return []tool{{"clip", []string{"clip"}, []string{"powershell", "-NoProfile", "-Command", "Get-Clipboard -Raw"}}}
	case "darwin":
		return []tool{{"pbcopy", []string{"pbcopy"}, []string{"pbpaste"}}}
	}
	var tools []tool
	if getenv("WAYLAND_DISPLAY") != "" {
		tools = append(tools, tool{"wl-copy", []string{"wl-copy"}, []string{"wl-paste", "--no-newline"}})
	}
	return append(tools,
		tool{"xclip", []string{"xclip", "-selection", "clipboard"}, []string{"xclip", "-selection", "clipboard", "-o"}},
		tool{"xsel", []string{"xsel", "--clipboard", "--input"}, []string{"xsel", "--clipboard", "--output"}})
}

// CommandRunner runs a program with text on its standard input and returns
// its standard output. It is the port to the operating system: tests pass a
// fake, so no test touches a real clipboard.
type CommandRunner func(stdin string, argv ...string) (string, error)

// ExecRunner is the CommandRunner that starts a real process.
func ExecRunner(stdin string, argv ...string) (string, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.Output()
	return string(out), err
}

type toolClipboard struct {
	t   tool
	run CommandRunner
}

func (c toolClipboard) Name() string { return c.t.name }
func (c toolClipboard) Write(text string) error {
	_, err := c.run(text, c.t.write...)
	return err
}
func (c toolClipboard) Read() (string, error) {
	out, err := c.run("", c.t.read...)
	return strings.TrimRight(out, "\r\n"), err
}
func (c toolClipboard) CanRead() bool { return true }

type terminalClipboard struct{ out io.Writer }

func (c terminalClipboard) Name() string { return "the terminal" }
func (c terminalClipboard) Write(text string) error {
	if text == "" {
		return nil
	}
	return clipboard.Copy(c.out, text)
}
func (c terminalClipboard) Read() (string, error) {
	return "", errors.New("the terminal clipboard cannot be read")
}
func (c terminalClipboard) CanRead() bool { return false }

// Find picks the clipboard to use: the first program of the operating system
// that is installed, else the terminal's OSC 52 sequence when out is a
// terminal. It returns nil when there is none; the caller then offers to print
// the key.
//
// Parameters: goos is runtime.GOOS; getenv reads the environment; lookPath
// finds a program on PATH; run starts one; out is the terminal, or nil.
func Find(goos string, getenv func(string) string, lookPath func(string) (string, error), run CommandRunner, out io.Writer) Clipboard {
	for _, t := range toolsFor(goos, getenv) {
		if _, err := lookPath(t.write[0]); err == nil {
			return toolClipboard{t, run}
		}
	}
	if out != nil {
		return terminalClipboard{out}
	}
	return nil
}

// FindSystem is Find for this machine.
//
// Parameters: out is the terminal, or nil when standard output is not one.
func FindSystem(out io.Writer) Clipboard {
	return Find(runtime.GOOS, os.Getenv, exec.LookPath, ExecRunner, out)
}
