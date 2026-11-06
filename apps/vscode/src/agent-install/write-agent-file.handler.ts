import type { CommandHost } from '../command-host/index.ts'
import { confirmAgentFile, tellAgentFileOutcome } from './agent-prompts.client.ts'
import { writeAgentFile } from './write-agent-file.use-case.ts'

/**
 * Writes AGENTS.md, from the agent menu or from the notice after the agent install.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function writeAgentsFile (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite('AGENTS.md')) return
  const outcome = await writeAgentFile({ run: host.smithIn(state), confirm: async (lines) => await confirmAgentFile(lines, state.name) })
  host.report('AGENTS.md', state, outcome)
  tellAgentFileOutcome(outcome, state.root)
}
