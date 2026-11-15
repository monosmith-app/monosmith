import { homedir } from 'node:os'
import type { CommandHost } from '../command-host/index.ts'
import type { FolderState } from '../folder-session/index.ts'
import { openSmithTerminal, smithForTerminal } from '../smith-terminal/index.ts'
import { offerTabCompletion } from './offer-tab-completion.use-case.ts'
import { askTabCompletion } from './tab-completion-prompts.client.ts'
import type { TabCompletionMemory } from './tab-completion-offer.policy.ts'

const MEMORY_KEY = 'monosmith.tabCompletionOffer'

/** What the tab completion commands need from the editor besides the host. */
export interface TabCompletionDeps {
  /** The editor's memory for the person, across workspaces. */
  globalState: { get: <T>(key: string) => T | undefined, update: (key: string, value: unknown) => PromiseLike<void> }
}

/** The commands of tab completion. */
export interface TabCompletionFlows {
  /** Offers the setup once, when smith says tab completion is not set up. Nothing is written. */
  offer: (state: FolderState) => Promise<void>
  /** Opens the terminal on `smith completion install`; the person answers its question there. */
  setUp: (argument: unknown) => Promise<void>
}

/**
 * Builds the commands of tab completion.
 * @param host - the folders and the Output channel of the extension
 * @param deps - the memory of the person
 * @returns the commands
 */
export function createTabCompletionFlows (host: CommandHost, deps: TabCompletionDeps): TabCompletionFlows {
  let askedThisSession = false

  return {
    offer: async (state) => {
      const location = state.location
      if (location === undefined) return
      await offerTabCompletion({
        recall:           () => deps.globalState.get<TabCompletionMemory>(MEMORY_KEY),
        remember:         async (answer) => { await deps.globalState.update(MEMORY_KEY, answer) },
        askedThisSession: () => askedThisSession,
        markAsked:        () => { askedThisSession = true },
        status:           async () => await host.smithIn(state, false)(['completion', 'status', '--json']),
        ask:              askTabCompletion,
        install:          () => { openSmithTerminal(location.executable, ['completion', 'install'], state.root) },
      })
    },
    setUp: async (argument) => {
      const state = await host.folderFor(argument)
      const root = state?.root ?? homedir()
      const executable = await smithForTerminal(root)
      if (executable === undefined) return
      openSmithTerminal(executable, ['completion', 'install'], root)
    },
  }
}
