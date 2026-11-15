import type { CommandHost } from '../command-host/index.ts'
import { openSmithTerminal } from '../smith-terminal/index.ts'

/**
 * Opens the remote cache setup (`smith cache setup`) in a MonoSmith terminal. The setup asks
 * questions and shows each cloud command before it runs, which a notice or a pick list cannot do
 * for a command smith runs for the person.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function setUpRemoteCache (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined) return
  openSmithTerminal(state.location.executable, ['cache', 'setup'], state.root)
}
