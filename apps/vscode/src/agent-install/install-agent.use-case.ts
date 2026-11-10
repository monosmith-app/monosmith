import { installArguments } from './install-arguments.mapper.ts'
import { isPreview, readPreview, readWrite } from './install-outcome.mapper.ts'
import type { SmithRun } from '../smith-run/index.ts'
import type { InstallOutcome, InstallPreview, InstallScope } from './install-run.contract.ts'

/** What installing needs from the editor. */
export interface InstallDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:     (args: string[]) => Promise<SmithRun>
  /** Shows what would be written and answers whether to write it. */
  confirm: (preview: InstallPreview) => Promise<boolean>
}

/**
 * Installs the MonoSmith agent and the MCP server (or the server alone) in the
 * workspace: asks smith what it would write, shows that, and only on a yes
 * writes it. smith never overwrites a file the team edited, so a refusal is
 * returned with smith's own words.
 * @param deps - how to run smith and ask the person
 * @param scope - the agent with the server, or the server alone
 * @param confirmed - the person already said yes (a command argument), so only the preview is read
 */
export async function installAgent (deps: InstallDeps, scope: InstallScope, confirmed = false): Promise<InstallOutcome> {
  const preview = readPreview(await deps.run(installArguments(scope, false)))
  if (!isPreview(preview)) return preview

  const agreed = confirmed || await deps.confirm(preview)

  return agreed ? readWrite(await deps.run(installArguments(scope, true))) : { kind: 'declined' }
}
