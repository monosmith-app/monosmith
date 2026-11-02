package agentfiles

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"golang.org/x/term"
)

// Install is what smith agent install and smith mcp install do.
type Install struct {
	// Agent includes the agent file.
	Agent bool
	// Claude and VSCode choose the clients; neither means the ones present.
	Claude, VSCode bool
	Yes            bool
	In             io.Reader
	Out            io.Writer
	// Interactive says a person can answer; tests set it.
	Interactive func() bool
}

func terminalAttached() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
}

// Run previews the files, asks (unless Yes), and writes them. It writes
// nothing when a file is the person's: it shows the difference and stops.
func (i Install) Run() error {
	root, err := workspace.FindRoot(".")
	if err != nil {
		return err
	}
	claude, vscode := i.Claude, i.VSCode
	if !claude && !vscode {
		claude, vscode = PresentClients(root, exec.LookPath)
	}
	exe, _ := os.Executable()
	files, err := Plan(Options{Root: root, Claude: claude, VSCode: vscode, Agent: i.Agent, Launcher: LauncherFor(root, exe)})
	if err != nil {
		if errors.Is(err, ErrNotJSON) {
			return problem.New(err.Error(),
				"smith does not edit around comments: add the smith server by hand (docs/manual/agent.md shows the entry)")
		}
		return err
	}
	var edited, todo []File
	for _, f := range files {
		switch {
		case f.Status == Edited:
			edited = append(edited, f)
		case f.Status.Writable():
			todo = append(todo, f)
		}
	}
	if len(edited) > 0 {
		for _, f := range edited {
			fmt.Fprintf(i.Out, "● %s differs from what smith writes\n%s\n", f.Path, textdiff.UnifiedDiff(f.Path, f.Old, f.New))
		}
		return problem.New(edited[0].Path+" was changed after smith wrote it, so smith left everything as it is",
			"keep your version: nothing to do (smith doctor accepts it)",
			"take smith's: delete "+edited[0].Path+" (or its smith entry in an MCP file) and run the command again")
	}
	if len(todo) == 0 {
		fmt.Fprintf(i.Out, "✓ The MonoSmith agent is already installed (%d files)\n", len(files))
		return nil
	}
	for _, f := range todo {
		fmt.Fprintf(i.Out, "● %s\n%s\n", f.action(), textdiff.UnifiedDiff(f.Path, f.Old, f.New))
	}
	if !i.Yes {
		interactive := i.Interactive
		if interactive == nil {
			interactive = terminalAttached
		}
		if !interactive() {
			return problem.New("nothing was written: there is no terminal to ask on",
				"run it again with --yes to write the files above")
		}
		fmt.Fprintf(i.Out, "Write %d file%s? [y/N] ", len(todo), map[bool]string{true: "", false: "s"}[len(todo) == 1])
		answer, _ := bufio.NewReader(i.In).ReadString('\n')
		if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
			fmt.Fprintln(i.Out, "Nothing written.")
			return nil
		}
	}
	written, err := Write(root, todo)
	for _, w := range written {
		fmt.Fprintf(i.Out, "✓ Wrote %s\n", w)
	}
	if err != nil {
		return err
	}
	i.next(claude, vscode)
	return nil
}

func (i Install) next(claude, vscode bool) {
	if claude && i.Agent {
		fmt.Fprintln(i.Out, "→ Claude Code: run `smith agent`, or start claude and pick the smith agent. It asks to trust the smith server in .mcp.json the first time.")
	} else if claude {
		fmt.Fprintln(i.Out, "→ Claude Code: restart it; it asks to trust the smith server in .mcp.json the first time.")
	}
	if vscode {
		fmt.Fprintln(i.Out, "→ VS Code: open .vscode/mcp.json and choose Start on the smith server, then use it in agent mode.")
	}
}
