import type { CommandHost } from '../command-host/index.ts'
import { confirmLauncher, tellLauncherOutcome } from './launcher-prompts.client.ts'
import { writeLauncher } from './write-launcher.use-case.ts'

/**
 * Writes the launcher (`smith bootstrap`) that makes a laptop and CI run the same smith.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function bootstrapHere (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite('the smith launcher')) return
  const outcome = await writeLauncher({ run: host.smithIn(state), confirm: async (plan) => await confirmLauncher(state.name, plan) })
  host.report('launcher', state, outcome)
  tellLauncherOutcome(outcome)
  if (outcome.kind === 'written') await host.refreshFolder(state.root, 'full', false)
}
