import type { CommandHost } from '../command-host/index.ts'
import { readFolderData, type ListNode } from '../list-view/index.ts'
import { openSmithTerminal } from '../smith-terminal/index.ts'
import { isWorkspaceTrusted } from '../workspace-trust/index.ts'
import { confirmUpdate, tellUpdateNeedsTrust } from './dependency-prompts.client.ts'
import { pickArguments, updateArguments } from './update-arguments.policy.ts'

/**
 * Starts the servers of the workspace (`smith up`) in a terminal, in the folder the person picks.
 * @param host - the folders and the Output channel of the extension
 */
export async function runUp (host: CommandHost): Promise<void> {
  const state = await host.pickReadyFolder()
  if (state?.location === undefined) return
  openSmithTerminal(state.location.executable, pickArguments(), state.root)
}

/**
 * Updates one dependency of a Dependencies row to its latest version, in a terminal, after a yes.
 * @param host - the folders and the Output channel of the extension
 * @param node - the row the command started on
 */
export async function updateDependency (host: CommandHost, node: ListNode | undefined): Promise<void> {
  const tagged = readFolderData(node?.data)
  const args = tagged === undefined ? undefined : updateArguments(tagged.data)
  const state = tagged === undefined ? undefined : host.registry.folders.get(tagged.root)
  if (tagged === undefined || args === undefined || state?.location === undefined) return
  if (!isWorkspaceTrusted()) {
    tellUpdateNeedsTrust()

    return
  }
  if (await confirmUpdate(tagged.data)) openSmithTerminal(state.location.executable, args, state.root)
}
