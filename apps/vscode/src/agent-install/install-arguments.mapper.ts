import type { InstallScope } from './install-run.contract.ts'

/**
 * The arguments of the smith command that installs.
 * @param scope - the agent and the MCP server, or the MCP server alone
 * @param write - `--yes`: write; without it smith only prints the diffs
 */
export function installArguments (scope: InstallScope, write: boolean): string[] {
  const files = scope === 'agent' ? ['agent', 'install', '--claude', '--vscode'] : ['mcp', 'install', '--vscode']

  return write ? [...files, '--yes'] : files
}

/** The arguments of the command that records that this workspace wants no agent files. */
export function optOutArguments (): string[] {
  return ['config', 'set', 'agent.enabled', 'false']
}
