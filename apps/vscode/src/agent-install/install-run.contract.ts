/** What a run of the smith executable left behind. */
export interface SmithRun {
  /** The exit code; 1 when smith could not be started. */
  code:   number
  stdout: string
  stderr: string
}

/** What to install: the agent with the MCP server, or the MCP server alone. */
export type InstallScope = 'agent' | 'mcp'

/** The files were written; `lines` are smith's `Wrote ...` lines. */
export interface InstallWritten { kind: 'written', lines: string[] }

/** smith refused or failed; `message` is what it said. */
export interface InstallRefused { kind: 'refused', message: string }

/**
 * What the install ended with, for the person who asked: written, `current`
 * (the files were already there), `declined` (the person saw the files and said
 * no), or refused.
 */
export type InstallOutcome = InstallWritten | InstallRefused | { kind: 'current' | 'declined' }

/** What `smith ... install` would write, read from a run without --yes. */
export interface InstallPreview {
  /** One line for each file, as smith words it: `Add .vscode/mcp.json`. */
  actions: string[]
  /** The unified diffs smith printed. */
  diff:    string
}
