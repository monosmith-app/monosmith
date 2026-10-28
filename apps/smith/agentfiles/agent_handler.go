package agentfiles

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

func installFlags(cmd *cobra.Command, i *Install) {
	cmd.Flags().BoolVar(&i.Claude, "claude", false, "for Claude Code: .mcp.json"+map[bool]string{true: " and .claude/agents/smith.md", false: ""}[i.Agent])
	cmd.Flags().BoolVar(&i.VSCode, "vscode", false, "for VS Code: .vscode/mcp.json")
	cmd.Flags().BoolVarP(&i.Yes, "yes", "y", false, "write without asking")
}

func installCommand(agent bool, use, short, long, example string) *cobra.Command {
	i := Install{Agent: agent}
	cmd := &cobra.Command{
		Use: use, Short: short, Long: long, Example: example,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			i.In, i.Out = cmd.InOrStdin(), cmd.OutOrStdout()
			return i.Run()
		},
	}
	installFlags(cmd, &i)
	return cmd
}

// AgentInstallCommand is `smith agent install`.
func AgentInstallCommand() *cobra.Command {
	return installCommand(true, "install", "Install the MonoSmith agent and the smith MCP server in this workspace",
		`Write the MonoSmith agent (.claude/agents/smith.md, for Claude Code) and register
smith mcp where the clients read it (.mcp.json for Claude Code, .vscode/mcp.json
for VS Code). An existing MCP file is merged: your other servers stay as they are.

It shows each file as a diff and asks before writing; --yes writes without asking.
Run again, it changes nothing when the files are current. A file you edited is
never overwritten: smith shows the difference and stops.

Without --claude or --vscode it writes for the clients the workspace or this
machine has, and for both when it finds neither. To keep a workspace without the
files, set enabled = false under [agent] in smith.toml and smith doctor stops asking.`,
		`  smith agent install
  smith agent install --claude --yes`)
}

// McpInstallCommand is `smith mcp install`: the MCP registration alone.
func McpInstallCommand() *cobra.Command {
	return installCommand(false, "install", "Install the smith MCP server for Claude Code and VS Code",
		`Register the smith MCP server for the clients that read it: .mcp.json for Claude
Code and .vscode/mcp.json for VS Code. It merges into an existing file and keeps your
other servers, shows the change as a diff and asks before writing (--yes writes
without asking), and never overwrites a smith entry you edited.

smith agent install does this and also writes the agent for Claude Code.`,
		`  smith mcp install
  smith mcp install --vscode --yes`)
}

// Command is `smith agent`: with no subcommand it starts the agent in Claude Code.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Start the MonoSmith agent in Claude Code; agent install adds it to a workspace",
		Long: `Start Claude Code with the MonoSmith agent (claude --agent smith) in this workspace.
It knows what smith can do and works through smith's own MCP server. It reads
projects, the doctor and dry runs on its own and asks before anything that writes,
installs, tags or publishes.

If the agent is not in the workspace yet, smith agent install adds it. Without the
claude command on the PATH, this prints how to install it.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return launch(cmd) },
	}
	cmd.AddCommand(AgentInstallCommand())
	return cmd
}

func launch(cmd *cobra.Command) error {
	root, err := workspace.FindRoot(".")
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(AgentPath))); err != nil {
		return problem.New("this workspace has no MonoSmith agent yet", "smith agent install")
	}
	claude, err := exec.LookPath("claude")
	if err != nil {
		return problem.New("Claude Code is not installed (no claude command on the PATH)",
			"install Claude Code from https://claude.com/claude-code, then run smith agent again",
			"in VS Code, the MonoSmith extension and .vscode/mcp.json give the same tools to its agent mode")
	}
	c := exec.Command(claude, "--agent", "smith")
	c.Dir = root
	c.Stdin, c.Stdout, c.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
	if err := c.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return problem.New("claude stopped with exit code " + strconv.Itoa(exit.ExitCode()))
		}
		return err
	}
	return nil
}
