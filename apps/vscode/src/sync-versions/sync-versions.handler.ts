import type { CommandHost } from '../command-host/index.ts'
import { chooseSyncRule, confirmSync, tellSyncOutcome } from './sync-prompts.client.ts'
import { syncVersions } from './sync-versions.use-case.ts'

/**
 * Asks which version wins, shows smith's sync plan, and writes it on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function syncHere (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite('syncing dependency versions')) return
  const outcome = await syncVersions({ run: host.smithIn(state), choose: chooseSyncRule, confirm: confirmSync })
  host.report('sync', state, outcome)
  tellSyncOutcome(outcome)
  if (outcome.kind === 'synced') await host.refreshFolder(state.root, 'full', false)
}
