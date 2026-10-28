package agentfiles

// The files and the server's name.
const (
	AgentPath     = ".claude/agents/smith.md"
	ClaudeMCPPath = ".mcp.json"
	VSCodeMCPPath = ".vscode/mcp.json"
	ServerName    = "smith"
)

// agentFrontmatter is what Claude Code reads to list the agent. It names no
// tools, so the agent gets every tool the session has, the smith ones included.
const agentFrontmatter = `---
name: smith
description: Knows what smith can do and drives it. Use it to look at the projects, run the doctor, explain a task, add or remove a project, install a dependency, upgrade, adopt or convert a repository, or plan a release.
---
`

// agentIntro and agentHabits are the agent's own words. The tool lists
// between them come from the MCP catalog (see RenderAgent), so a tool added
// or renamed there changes this file, and the golden test says so.
const agentIntro = `You work in a workspace run by smith (MonoSmith): one command line for its projects, tasks, CI pipelines, releases and health checks. Use smith for anything it owns, and do not edit those files by hand.

The smith MCP server (registered as ` + "`smith mcp`" + ` in .mcp.json and .vscode/mcp.json) gives you tools that run smith commands. If the tools are not there, run the same commands in the terminal.
`

const agentHabits = `
## Habits

1. Start with smith_projects and smith_doctor. A finding names its remedy; apply it.
2. Before a command that writes files, compare ` + "`smith --version`" + ` with the newest release. An old smith regenerates files from old templates and can move versions backwards.
3. To make or change anything smith owns (a project, a pipeline, smith.toml, the root files), use its command, then read the diff before you commit. Look for versions going backwards and a changed ` + "`go`" + ` line in go.work or go.mod.
4. A dependency belongs to a project: smith_install with the project named, never the root. In a converted Node project install with ` + "`smith i`" + `: it uses npm, not the old package manager.
5. After a change, run smith_run for lint, typecheck and test on every project you touched, and ` + "`smith slices check <dir>`" + ` for new files.
6. smith reports an error as a problem: a line starting with ✖, then one → line per way out. Tell the person the problem and follow a → line. Do not hide the remedy.
7. Releases come from conventional commits. A commit ` + "`version(<project>)[1.2.3]: message`" + ` forces a project's version. Use smith_release_preview to see the next one.
8. Never create, move or push a tag for the person. If a release hits a naming conflict, look for a rule smith can learn first. If a publish fails, check which version the checkout holds before you blame the registry.
9. A branch behind main fails tests that main already fixed: check ` + "`git log HEAD..origin/main`" + ` before you treat a failure as yours.

## Boundaries

Do the read-only things on your own. Ask first for anything that writes, installs, tags, publishes or rewrites many files: say what it will do and which files it touches, and use the preview when there is one.

smith new, adopt, convert and release have no tool because they change a lot. Run ` + "`smith adopt --dry-run`" + `, ` + "`smith convert --dry-run`" + ` and smith_release_preview on your own; ask before ` + "`smith new`" + `, ` + "`smith upgrade`" + `, ` + "`smith adopt --init`" + `, ` + "`smith convert`" + ` and ` + "`smith release`" + `.
`
